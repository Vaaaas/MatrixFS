package main

import (
	"github.com/golang/glog"
	"flag"
	"Vaaaas/MatrixFS/SysConfig"
	"Vaaaas/MatrixFS/File"
	"fmt"
	"net/http"
	"log"
)

func main() {
	var fault, row int

	//初始化命令行参数
	flag.IntVar(&fault, "fault", 2, "系统容错数量")
	flag.IntVar(&row, "row", 2, "文件阵列行数")
	//初始化命令行参数
	//调试时log_dir="./log"
	flag.Parse()
	//退出时调用，确保日志写入文件
	defer glog.Flush()
	//初始化系统配置
	SysConfig.InitConfig(fault, row)
	glog.Info("Server start here")

	testFileHandle()

	//http.Handle("/view/", http.FileServer(http.Dir("./view/")))
	http.Handle("/view/", http.StripPrefix("/view/", http.FileServer(http.Dir("./view/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))


	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Panic(err)
	}
}

func testFileHandle() {
	//SysConfig & File 包测试
	var file01 File.File
	file01.Init("/Users/vaaaas/Desktop/READING/PDF.pdf")
	name, ext := file01.SliceFileName()
	fmt.Println(name + " , " + ext)
	file01.InitDataFiles()
	file01.InitRddtFiles()
	file01.GetFile("/Users/vaaaas/Desktop/WRITING/")
}

//todo : 用bee搭建服务器