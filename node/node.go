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

	"github.com/golang/glog"
)

//Node 节点结构体
type Node struct {
	ID       uint    `json:"ID"`
	Address  net.IP  `json:"Address"`
	Port     int     `json:"Port"`
	Volume   float64 `json:"Volume"`
	Status   bool    `json:"Status"`
	LastTime int64   `json:"Lasttime"`
}

var nodeInfo Node

var masterAdd *net.TCPAddr
var nodeAdd *net.TCPAddr
var storePath string

func main() {
	var local string
	var master string

	flag.StringVar(&master, "master", "127.0.0.1:8080", "Master server IP & Port")
	flag.StringVar(&local, "node", "127.0.0.1:9090", "Local Node IP & Port")
	flag.StringVar(&storePath, "stpath", "./storage", "Local Storage Path")
	//log_dir="./log"
	flag.Parse()

	//退出时将日志写入文件
	defer glog.Flush()
	err := os.MkdirAll(flag.Lookup("log_dir").Value.String(), 0766)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	err = os.MkdirAll(storePath, 0766)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}

	//初始化中心节点和本节点的IP
	masterAdd, err = net.ResolveTCPAddr("tcp4", master)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	nodeAdd, err = net.ResolveTCPAddr("tcp4", local)
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
	diskUsage := NewDiskUsage(dir)
	volume := (float64)(diskUsage.Available()) / (float64)(1024*1024*1024)
	glog.Infof("Node start here : %d", nodeAdd.Port)
	glog.Infof("Store Path: %s, Free space: %f", storePath, volume)
	nodeInfo.ID = 0
	initStruct(&nodeInfo, nodeAdd.IP, nodeAdd.Port, volume)
	go func() {
		for {
			connectMaster()
			time.Sleep(3 * time.Second)
		}
	}()

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/resetid", resetIDHandler)
	http.Handle("/download/", http.StripPrefix("/download/", http.FileServer(http.Dir(storePath))))
	if err := http.ListenAndServe(":"+strconv.Itoa(nodeAdd.Port), nil); err != nil {
		glog.Errorln(err)
		panic(err)
	}
}

func uploadHandler(_ http.ResponseWriter, r *http.Request) {
	glog.Infoln("[UPLOAD] method: " + r.Method)
	if r.Method == "GET" {
		glog.Warningln("[/UPLOAD] GET" + r.URL.Path)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			glog.Error(err)
			panic(err)
		}
		defer file.Close()
		longSpl := strings.Split(handler.Filename, "/")
		err = os.MkdirAll(storePath+"/", 0766)
		if err != nil {
			glog.Errorln(err)
			panic(err)
		}
		f, err := os.Create(storePath + "/" + longSpl[len(longSpl)-1])
		if err != nil {
			glog.Error(err)
			panic(err)
		}
		defer f.Close()
		io.Copy(f, file)
	}
}

func deleteHandler(_ http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		fileName := r.Header.Get("fileName")
		glog.Info("File to Delete : " + fileName)
		if _, err := os.Stat(storePath + "/" + fileName); os.IsNotExist(err) {
			glog.Warningf("[File to Delete NOT EXIST] %s", storePath+"/"+fileName)
		} else {
			err := os.Remove(storePath + "/" + fileName)
			if err != nil {
				glog.Errorln(err)
			}
		}

	} else {
		glog.Warningln("[/DELETE] else" + r.URL.Path)
	}
}

func resetIDHandler(_ http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		newID := r.Header.Get("NewID")
		tmpID, err := strconv.Atoi(newID)
		glog.Infof("NewID : %d, New Status : %t", tmpID, false)
		if err != nil {
			glog.Error(err)
		}
		nodeInfo.ID = (uint)(tmpID)
		nodeInfo.Status = false
	} else {
		glog.Warningln("[/ResetID] else" + r.URL.Path)
	}
}

func connectMaster() error {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dir = strings.Replace(dir, "\\", "/", -1)
	diskUsage := NewDiskUsage(dir)
	volume := (float64)(diskUsage.Available()) / (float64)(1024*1024*1024)
	nodeInfo.Volume = volume
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(nodeInfo)
	res, err := http.Post("http://"+masterAdd.String()+"/greet", "application/json; charset=utf-8", b)
	if err != nil {
		glog.Error(err)
	}
	id, err := strconv.Atoi(res.Header.Get("ID"))
	nodeInfo.ID = (uint)(id)
	if err != nil {
		glog.Error(err)
	}
	return nil
}

func initStruct(nodeInfo *Node, address net.IP, port int, volume float64) error {
	nodeInfo.Address = address
	nodeInfo.Port = port
	nodeInfo.Volume = volume
	nodeInfo.Status = true

	return nil
}
