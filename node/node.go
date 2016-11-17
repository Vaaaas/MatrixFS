package main

import (
	"flag"
	"os"
	"github.com/golang/glog"
	"net/http"
	"net"
)

func main() {
	var master string
	var localPort string

	flag.StringVar(&master, "master", "localhost:8080", "Master server IP & Port")
	flag.StringVar(&localPort, "port", "9090", "Local Node Port")

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

	connectMaster(master)

	if err := http.ListenAndServe(":" + localPort, nil); err != nil {
		glog.Errorln(err)
		panic(err)
	}
}

func connectMaster(master string) error {
	//glog.Infof("Master IP : %s", master)

	masterAddr, err := net.ResolveTCPAddr("tcp4", master)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	conn, err := net.DialTCP("tcp", nil, masterAddr)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	_, err = conn.Write([]byte("HEAD / HTTP/1.0"))

	return nil
}