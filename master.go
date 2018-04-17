package main

import (
	"flag"
	"net/http"
	"os"
	"github.com/Vaaaas/MatrixFS/filehandler"
	"github.com/Vaaaas/MatrixFS/nodehandler"
	"github.com/Vaaaas/MatrixFS/server"
	"github.com/Vaaaas/MatrixFS/util"
	"github.com/golang/glog"
)

func main() {
	//log_dir="./log"
	flag.Parse()
	//退出时调用 将日志写入文件
	defer glog.Flush()

	glog.Info("Master 服务器启动")
	//建立临时文件存储文件夹
	err := os.MkdirAll("./temp", 0766)
	if err != nil {
		glog.Errorln(err)
	}
	//建立日志存储文件夹
	err = os.MkdirAll("./log", 0766)
	if err != nil {
		glog.Errorln(err)
	}

	nodehandler.IDCounter = util.NewSafeID()
	//初始化节点Map
	nodehandler.AllNodes = util.NewSafeMap()
	//初始化文件Map
	filehandler.AllFiles = util.NewSafeMap()

	//页面处理方法
	http.HandleFunc("/", server.RootHandler)
	http.HandleFunc("/index", server.IndexPageHandler)
	http.HandleFunc("/file", server.FilePageHandler)
	http.HandleFunc("/node", server.NodeEnterHandler)

	//功能处理方法
	http.HandleFunc("/greet", server.GreetHandler)
	http.HandleFunc("/upload", server.UploadHandler)
	http.HandleFunc("/delete", server.DeleteHandler)
	http.HandleFunc("/download", server.DownloadHandler)
	http.HandleFunc("/restore", server.RestoreHandler)

	//定时遍历所有节点，比较最后访问时间
	go func() {
		nodehandler.NodeStatusDetect()
	}()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		glog.Errorln(err)
	}
}
