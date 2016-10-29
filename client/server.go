package main

import (
	"github.com/golang/glog"
	"flag"
	"Vaaaas/MatrixFS/SysConfig"
	"Vaaaas/MatrixFS/File"
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

	glog.Info("Server start here")

	//其他日志类型
	//glog.Warning("Warning!")
	//glog.Error("Error!")

	//SysConfig & File 包测试
	SysConfig.InitConfig(fault, row)
	var file01 File.File
	file01.Init("/Users/vaaaas/Desktop/READING/FILE01")
	//name, ext := file01.SliceFileName()
	//fmt.Println(name + " , " + ext)
	file01.InitDataFiles()
	file01.InitRddtFiles()
	file01.GetFile("/Users/vaaaas/Desktop/WRITING/")
}

//todo : 用bee搭建服务器