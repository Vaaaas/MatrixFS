package main

import (
	"flag"
	"os"
	"github.com/golang/glog"
	"net/http"
	"net"
	"github.com/Vaaaas/MatrixFS/NodeStruct"
	"strconv"
	"bytes"
	"encoding/json"
	"path/filepath"
	"log"
	"strings"
	"io"
)

var nodeInfo NodeStruct.Node
var MasterAdd *net.TCPAddr
var NodeAdd *net.TCPAddr

var StorePath string

func main() {
	var local string
	var master string
	flag.StringVar(&master, "master", "127.0.0.1:8080", "Master server IP & Port")
	flag.StringVar(&local, "node", "127.0.0.1:9090", "Local Node IP & Port")
	flag.StringVar(&StorePath, "stpath", "./storage", "Local Storage Path")
	//when debug, log_dir="./log"
	flag.Parse()
	//Trigger on exit, write log into files
	defer glog.Flush()
	glog.Info("Node start here")
	err := os.MkdirAll("./log", 0766)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	err = os.MkdirAll(StorePath, 0766)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}

	//Init ID Counter, IDCounter++ when create new Node
	NodeStruct.IDCounter = 0

	//Init Master & Node Address
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
	//Get Free Space of Storage Path
	volume := NodeStruct.DiskUsage(dir)
	glog.Infof("Store Path: %s, Free space: %d", StorePath, volume)

	InitStruct(&nodeInfo, NodeAdd.IP, NodeAdd.Port, volume)

	connectMaster(MasterAdd)

	http.HandleFunc("/upload", uploadHandler)

	if err := http.ListenAndServe(":" + strconv.Itoa(NodeAdd.Port), nil); err != nil {
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
		//fmt.Fprintf(w, "%v", handler.Header)
		longSpl := strings.Split(handler.Filename, "/")
		typeSpl := strings.Split(longSpl[len(longSpl) - 2], ".")

		err = os.MkdirAll(StorePath + "/" + typeSpl[0] + "/", 0766)
		if err != nil {
			glog.Errorln(err)
			panic(err)
		}

		f, err := os.OpenFile(StorePath + "/" + typeSpl[0] + "/" + longSpl[len(longSpl) - 1], os.O_WRONLY | os.O_CREATE, 0666)
		if err != nil {
			glog.Error(err)
			panic(err)
		}
		defer f.Close()
		io.Copy(f, file)
	}
}

func connectMaster(master *net.TCPAddr) error {
	glog.Infof("Connect to Master IP : %s", master.String())

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(nodeInfo)
	res, err := http.Post("http://" + MasterAdd.String() + "/greet", "application/json; charset=utf-8", b)
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

func InitStruct(nodeInfo *NodeStruct.Node, address net.IP, port int, volume float64) error {
	nodeInfo.Address = address
	nodeInfo.Port = port
	nodeInfo.Volume = volume
	nodeInfo.Status = true

	return nil
}