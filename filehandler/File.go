package filehandler

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/Vaaaas/MatrixFS/sysTool"
	"github.com/golang/glog"
)

//AllFiles 所有文件对象列表
var AllFiles []File

//File 文件类
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

//FileExistedInCenter 查看某个文件是否在磁盘中存在
func FileExistedInCenter(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

//FileExistInNode 查看某个文件能否从节点获取
func FileExistInNode(file File, isData bool, nodeID uint, posiX, posiY, rddtNodePos int) bool {
	return getOneFile(file, isData, nodeID, posiX, posiY, rddtNodePos)
}

//StructSliceFileName 利用存储根目录，原始文件名，数据/校验，存储节点编号，数据节点编号/斜率，行号/起始数据节点编号构建存储路径
func StructSliceFileName(storagePosition string, isDataSlice bool, nodePos int, fileName string, dataNodeNumK int, rowStartDataPos int) string {
	var dataOrRddt string
	if isDataSlice {
		dataOrRddt = "Data."
	} else {
		dataOrRddt = "Rddt."
	}
	return "./" + storagePosition + "/" + dataOrRddt + strconv.Itoa(nodePos) + "/" + fileName + "." + strconv.Itoa(dataNodeNumK) + strconv.Itoa(rowStartDataPos)
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
