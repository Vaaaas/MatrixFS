package sysTool

import (
	"os"
	"strconv"

	"github.com/golang/glog"
)

//SysConfig 首字母大写为导出成员，可被包外引用
var SysConfig struct {
	FaultNum int
	RowNum   int
	SliceNum int
	DataNum  int
	RddtNum  int

	Status bool
}

//SysConfigured 判断是否已经配置最大容错数和阵列行数
func SysConfigured() bool {
	return SysConfig.FaultNum != 0 && SysConfig.RowNum != 0
}

//InitConfig 初始化容错数和行数配置
func InitConfig(fault, row int) {
	if fault <= 1 || row <= 1 {
		glog.Error("行数和容错数都应大于1")
		panic("行数和容错数都应大于1")
	} else {
		SysConfig.FaultNum = fault
		SysConfig.RowNum = row
		SysConfig.DataNum = SysConfig.RowNum*SysConfig.FaultNum - SysConfig.FaultNum + 1
		SysConfig.SliceNum = SysConfig.DataNum * SysConfig.RowNum
		if (fault*SysConfig.DataNum)%row != 0 {
			SysConfig.RddtNum = (fault*SysConfig.DataNum)/row + 1

		} else {
			SysConfig.RddtNum = (fault * SysConfig.DataNum) / row
		}

		SysConfig.Status = true

		glog.Infof("系统参数配置完成：容错数 %d , 行数 %d , 数据分块数 %d , 数据列数 %d , 冗余列数 %d", SysConfig.FaultNum, SysConfig.RowNum, SysConfig.SliceNum, SysConfig.DataNum, SysConfig.RddtNum)

		err := initTempFolders(fault, SysConfig.DataNum, row)
		if err != nil {
			glog.Error(err)
			panic(err)
		}
	}
}

//初始化临时文件夹
func initTempFolders(fault, dataNum, row int) error {
	err := initDataFolders(dataNum)
	if err != nil {
		glog.Error(err)
		return err
	}
	if (fault*dataNum)%row != 0 {
		initRddtFolders((fault*dataNum)/row + 1)

	} else {
		initRddtFolders((fault * dataNum) / row)
	}

	return nil
}

func initDataFolders(dataNum int) error {
	for i := 0; i < dataNum; i++ {
		err := os.MkdirAll("./temp/DATA."+strconv.Itoa(i), 0766)
		if err != nil {
			return err
		}
	}

	return nil
}

func initRddtFolders(rddtNum int) error {
	for i := 0; i < rddtNum; i++ {
		err := os.MkdirAll("./temp/RDDT."+strconv.Itoa(i), 0766)
		if err != nil {
			return err
		}
	}

	return nil
}

//GetIndexInAll 利用编号在列表中获取对象
func GetIndexInAll(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}
