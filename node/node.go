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
	"io"
)

var nodeInfo NodeStruct.Node
var MasterAdd *net.TCPAddr
var NodeAdd *net.TCPAddr

func main() {
	var local string
	var storePath string
	var master string
	flag.StringVar(&master, "master", "127.0.0.1:8080", "Master server IP & Port")
	flag.StringVar(&local, "node", "127.0.0.1:9090", "Local Node IP & Port")
	flag.StringVar(&storePath, "stpath", "./storage", "Local Storage Path")
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
	err = os.MkdirAll(storePath, 0766)
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
	//Get Free Space of Storage Path
	volume := NodeStruct.DiskUsage(storePath)
	glog.Infof("Store Path: %s, Free space: %d", storePath, volume)

	InitStruct(&nodeInfo, NodeAdd.IP, NodeAdd.Port, volume)

	connectMaster(MasterAdd)

	http.HandleFunc("/getID", getIDHandler)

	if err := http.ListenAndServe(":" + strconv.Itoa(NodeAdd.Port), nil); err != nil {
		glog.Errorln(err)
		panic(err)
	}
}

func getIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&nodeInfo.ID)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	glog.Infoln("ID Refreshed! %d", nodeInfo.ID)
}

func connectMaster(master *net.TCPAddr) error {
	glog.Infof("Connect to Master IP : %s", master.String())

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(nodeInfo)
	res, err := http.Post("http://" + MasterAdd.String() + "/greet", "application/json; charset=utf-8", b)
	if err != nil {
		glog.Error(err)
	}
	io.Copy(os.Stdout, res.Body)

	return nil
}

func InitStruct(nodeInfo *NodeStruct.Node, address net.IP, port int, volume float64) error {
	nodeInfo.Address = address
	nodeInfo.Port = port
	nodeInfo.Volume = volume
	nodeInfo.Status = true

	return nil
}