package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Vaaaas/MatrixFS/nodehandler"
	"github.com/Vaaaas/go-disk-usage/du"
	"github.com/golang/glog"
)

var nodeInfo nodehandler.Node
var MasterAdd *net.TCPAddr
var NodeAdd *net.TCPAddr

var StorePath string

func main() {
	var local string
	var master string
	flag.StringVar(&master, "master", "127.0.0.1:8080", "Master server IP & Port")
	flag.StringVar(&local, "node", "127.0.0.1:9090", "Local Node IP & Port")
	flag.StringVar(&StorePath, "stpath", "./storage", "Local Storage Path")
	//log_dir="./log"
	flag.Parse()

	//退出时将日志写入文件
	defer glog.Flush()
	err := os.MkdirAll(flag.Lookup("log_dir").Value.String(), 0766)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	err = os.MkdirAll(StorePath, 0766)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}

	//初始化中心节点和本节点的IP
	MasterAdd, err = net.ResolveTCPAddr("tcp4", master)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	NodeAdd, err = net.ResolveTCPAddr("tcp4", local)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dir = strings.Replace(dir, "\\", "/", -1)
	//获取存储节点本地存储路径的可用空间
	diskUsage := du.NewDiskUsage(dir)
	volume := (float64)(diskUsage.Available()) / (float64)(1024*1024*1024)
	glog.Infof("Node start here : %d", NodeAdd.Port)

	glog.Infof("Store Path: %s, Free space: %f", StorePath, volume)

	nodeInfo.ID = 0
	initStruct(&nodeInfo, NodeAdd.IP, NodeAdd.Port, volume)

	go func() {
		for {
			//glog.Info("Connect Master!")
			connectMaster(MasterAdd)
			time.Sleep(4 * time.Second)
		}
	}()

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/resetid", resetIDHandler)

	http.Handle("/download/", http.StripPrefix("/download/", http.FileServer(http.Dir(StorePath))))

	if err := http.ListenAndServe(":"+strconv.Itoa(NodeAdd.Port), nil); err != nil {
		glog.Errorln(err)
		panic(err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[UPLOAD] method: ", r.Method)
	if r.Method == "GET" {
		glog.Infoln("[/UPLOAD] " + r.URL.Path)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			glog.Error(err)
			panic(err)
		}
		defer file.Close()
		longSpl := strings.Split(handler.Filename, "/")

		err = os.MkdirAll(StorePath+"/", 0766)
		if err != nil {
			glog.Errorln(err)
			panic(err)
		}

		f, err := os.OpenFile(StorePath+"/"+longSpl[len(longSpl)-1], os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			glog.Error(err)
			panic(err)
		}
		defer f.Close()
		io.Copy(f, file)
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[DELETE] method: ", r.Method)
	if r.Method == "DELETE" {
		glog.Infoln("[/DELETE] " + r.URL.Path)

		fileName := r.Header.Get("fileName")
		glog.Info("File to Delete : " + fileName)

		glog.Info("File to [Delete] Name: " + fileName)
		if _, err := os.Stat(StorePath + "/" + fileName); os.IsNotExist(err) {
			glog.Warningf("[File to Delete NOT EXIST] %s", StorePath+"/"+fileName)
		} else {
			err := os.Remove(StorePath + "/" + fileName)
			if err != nil {
				glog.Errorln(err)
			} else {
				glog.Infof("[File to Delete] %s", StorePath+"/"+fileName)
			}
		}

	} else {
		glog.Infoln("[/DELETE] else" + r.URL.Path)
	}
}

func resetIDHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[ResetID] method: ", r.Method)
	if r.Method == "POST" {
		glog.Infoln("[/ResetID] " + r.URL.Path)

		newID := r.Header.Get("NewID")

		tmpID, err := strconv.Atoi(newID)

		glog.Infof("NewID : %d, New Status : %t", tmpID, false)

		if err != nil {
			glog.Error(err)
		}
		nodeInfo.ID = (uint)(tmpID)
		nodeInfo.Status = false
	} else {
		glog.Infoln("[/ResetID] else" + r.URL.Path)
	}
}

func connectMaster(master *net.TCPAddr) error {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dir = strings.Replace(dir, "\\", "/", -1)
	//Get Free Space of Storage Path
	diskUsage := du.NewDiskUsage(dir)
	volume := (float64)(diskUsage.Available()) / (float64)(1024*1024*1024)
	//glog.Infof("Store Path: %s, Free space: %f", StorePath, volume)

	nodeInfo.Volume = volume
	//glog.Infof("Connect to Master IP : %s", master.String())

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(nodeInfo)
	res, err := http.Post("http://"+MasterAdd.String()+"/greet", "application/json; charset=utf-8", b)
	if err != nil {
		glog.Error(err)
	}
	id, err := strconv.Atoi(res.Header.Get("ID"))
	nodeInfo.ID = (uint)(id)
	if err != nil {
		glog.Error(err)
	}

	//glog.Info("Connect Finished")
	return nil
}

func initStruct(nodeInfo *nodehandler.Node, address net.IP, port int, volume float64) error {
	nodeInfo.Address = address
	nodeInfo.Port = port
	nodeInfo.Volume = volume
	nodeInfo.Status = true

	return nil
}
