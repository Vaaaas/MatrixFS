package main

import (
	"flag"
	"os"
	"github.com/golang/glog"
	"net/http"
	"strings"
)

func main() {
	var master string
	var localPort string

	flag.StringVar(&master, "master", "localhost:8080", "Master server IP & Port")
	flag.StringVar(&localPort, "port", ":9090", "Local Node Port")

	//when debug, log_dir="./log"
	flag.Parse()
	//Trigger on exit, write log into files
	defer glog.Flush()
	glog.Info("Node start here")


	err := os.MkdirAll("./log", 0766)
	if err != nil {
		glog.Errorln(err)
	}

	ipNport := strings.Split(master, ":")
	connectMaster(ipNport[0], ipNport[1])

	if err := http.ListenAndServe(":" + localPort, nil); err != nil {
		glog.Errorln(err)
	}
}

func connectMaster(ip, port string) error {
	glog.Infof("Master IP : %s, Port : %s", ip, port)



	return nil
}