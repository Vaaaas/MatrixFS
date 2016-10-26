package main

import (
	"github.com/golang/glog"
	"flag"
	"strings"
	"os"
	"errors"
	"fmt"
)

//首字母大写为导出成员，可被包外引用
var SysConfig struct {
	FaultNum int
	RowNum   int
	SliceNum int16
}

func initConfig(fault, row int) {
	if (fault <= 1 || row <= 1) {
		glog.Error("行数和容错数都应大于1")
	} else {
		SysConfig.FaultNum = fault
		SysConfig.RowNum = row
		dataNum := SysConfig.RowNum * SysConfig.FaultNum - SysConfig.FaultNum + 1
		SysConfig.SliceNum = (int16)(dataNum * SysConfig.RowNum)
		glog.Infof("系统参数配置完成：容错数 %d , 行数 %d , 数据分块数 %d", SysConfig.FaultNum, SysConfig.RowNum, SysConfig.SliceNum)
	}
}

type File struct {
	FileFullName string
	Size         int64
	FillLast     bool
	FillSize     int16
	SliceSize    int64
}

func (file *File) init(source string) error {
	fileInfo, err := os.Stat(source)
	if err != nil {
		panic(err)
	}
	if fileInfo.IsDir() {
		glog.Error("初始化失败：该路径指向的是文件夹")
		return errors.New("初始化失败：该路径指向的是文件夹")
	} else {
		file.FileFullName = fileInfo.Name()
		file.Size = fileInfo.Size()
		if ((file.Size % (int64)(SysConfig.SliceNum)) != 0) {
			file.FillLast = true
			file.SliceSize = file.Size / (int64)(SysConfig.SliceNum) + 1
			file.FillSize = (int16)(file.SliceSize * (int64)(SysConfig.SliceNum) - file.Size)
		} else {
			file.FillLast = false
			file.SliceSize = file.Size / (int64)(SysConfig.SliceNum)
			file.FillSize = 0
		}
		glog.Infof("%+v ", file)
		return nil
	}
}

func (file File) sliceFileName() (string, string) {
	if file.FileFullName != "" {
		slices := strings.Split(file.FileFullName, ".")
		if len(slices) > 1 {
			name := strings.Join(slices, "")
			exten := slices[len(slices) - 1]
			return name, exten
		} else {
			glog.Error("分割文件名失败：该文件不含扩展名")
			return "", ""
		}
	} else {
		glog.Error("分割文件名失败：未定义文件名")
		return "", ""
	}
}

func main() {
	var fault, row int

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

	initConfig(fault, row)
	var file01 File
	file01.init("/Users/vaaaas/Desktop/READING/MP3.mp3")
	name, ext := file01.sliceFileName()
	fmt.Println(name + " , " + ext)
}