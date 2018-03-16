package fileHandler

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/Vaaaas/MatrixFS/sysTool"
)

//AllFiles 所有文件对象列表
var AllFiles []File

type File struct {
	FileFullName string
	Size         int64
	FillLast     bool
	FillSize     int64
	SliceSize    int64
}

//FindFileInAll 在全部文件列表中通过文件名查找文件对象
func FindFileInAll(name string) *File {
	for _, tempFile := range AllFiles {
		if tempFile.FileFullName == name {
			return &tempFile
		}
	}
	return nil
}

//FileExisted 查看某个文件是否在磁盘中存在
func FileExisted(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

//StructSliceFileName 利用存储根目录，原始文件名，数据/校验，存储节点编号，数据节点编号/斜率，行号/起始数据节点编号构建存储路径
func StructSliceFileName(storagePosition string, isDataSlice bool, nodeNum int, fileName string, varNum1 int, varNum2 int) string {
	var dataOrRddt string
	if isDataSlice {
		dataOrRddt = "Data."
	} else {
		dataOrRddt = "Rddt."
	}
	return "./" + storagePosition + "/" + dataOrRddt + strconv.Itoa(nodeNum) + "/" + fileName + "." + strconv.Itoa(varNum1) + strconv.Itoa(varNum2)
}

//Init 初始化文件对象
func (file *File) Init(source string) error {
	//获取原始文件属性
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
	//判断是否需要在文件末尾补0
	if (file.Size % (int64)(sysTool.SysConfig.SliceNum)) != 0 {
		file.FillLast = true
		file.SliceSize = file.Size/(int64)(sysTool.SysConfig.SliceNum) + 1
		file.FillSize = file.SliceSize*(int64)(sysTool.SysConfig.SliceNum) - file.Size
	} else {
		file.FillLast = false
		file.SliceSize = file.Size / (int64)(sysTool.SysConfig.SliceNum)
		file.FillSize = 0
	}
	glog.Infof("%+v ", file)
	//添加入所有文件列表
	AllFiles = append(AllFiles, *file)
	glog.Infof("All Files: %+v  ,len: %d", AllFiles[len(AllFiles)-1], len(AllFiles))
	return nil
}

//SliceFileName 分割原始文件名，分别返回文件名和后缀名
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
			for _, s := range slices[1: len(slices)-1] {
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

