package filehandler

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/Vaaaas/MatrixFS/util"
	"github.com/golang/glog"
)

// AllFiles key：节点ID Value：节点对象
var AllFiles *util.SafeMap

// File 文件类
type File struct {
	FileFullName string
	Size         int64
	fillLast     bool
	fillSize     int64
	sliceSize    int64
}

// fileExistedInCenter 查看某个文件是否在磁盘中存在
func fileExistedInCenter(filePath string) bool {
	//glog.Infoln("[File to Delete] path  "+filePath)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// Init 初始化文件对象
func (file *File) Init(source string) error {
	// 获取原始文件属性
	fileInfo, err := os.Stat(source)
	if err != nil {
		glog.Error(err)
		panic(err)
	}
	if fileInfo.IsDir() {
		glog.Error("初始化失败：该路径指向的是文件夹")
		return errors.New("初始化失败：该路径指向的是文件夹")
	}
	file.FileFullName = fileInfo.Name()
	file.Size = fileInfo.Size()
	// 判断是否需要在文件末尾补0
	if (file.Size % (int64)(util.SysConfig.SliceNum)) != 0 {
		file.fillLast = true
		file.sliceSize = file.Size/(int64)(util.SysConfig.SliceNum) + 1
		file.fillSize = file.sliceSize*(int64)(util.SysConfig.SliceNum) - file.Size
	} else {
		file.fillLast = false
		file.sliceSize = file.Size / (int64)(util.SysConfig.SliceNum)
		file.fillSize = 0
	}
	glog.Infof("%+v ", file)
	// 添加入所有文件列表
	AllFiles.Set(file.FileFullName, file)
	glog.Infof("All Files: %d", AllFiles.Count())
	return nil
}

// SliceFileName 分割原始文件名，分别返回文件名和后缀名
func (file File) SliceFileName() (string, string) {
	if file.FileFullName != "" {
		slices := strings.Split(file.FileFullName, ".")
		if len(slices) > 1 {
			n := len(".") * (len(slices) - 1)
			for i := 0; i < len(slices)-1; i++ {
				n += len(slices[i])
			}
			name := make([]byte, n-1)
			bp := copy(name, slices[0])
			for _, s := range slices[1 : len(slices)-1] {
				bp += copy(name[bp:], ".")
				bp += copy(name[bp:], s)
			}
			exten := slices[len(slices)-1]
			return string(name), exten
		}
		glog.Error("分割文件名失败：该文件不含扩展名")
		return "", ""
	}
	glog.Error("分割文件名失败：未定义文件名")
	return "", ""
}

// structSliceFileName 利用存储根目录，原始文件名，数据/校验，存储节点编号，数据节点编号/斜率，行号/起始数据节点编号构建存储路径
func structSliceFileName(storagePosition string, isDataSlice bool, nodePos int, fileName string, dataNodeNumK int, rowStartDataPos int) string {
	var dataOrRddt string
	if isDataSlice {
		dataOrRddt = "DATA."
	} else {
		dataOrRddt = "RDDT."
	}
	return storagePosition + "/" + dataOrRddt + strconv.Itoa(nodePos) + "/" + fileName + "." + strconv.Itoa(dataNodeNumK) + strconv.Itoa(rowStartDataPos)
}
