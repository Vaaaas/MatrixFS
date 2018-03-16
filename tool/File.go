package tool

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

//AllFiles 所有文件对象列表
var AllFiles []File

//FindFileInAll 在全部文件列表中通过文件名查找文件对象
func FindFileInAll(name string) *File {
	for _, tempFile := range AllFiles {
		if tempFile.FileFullName == name {
			return &tempFile
		}
	}
	return nil
}

//File 使用时需要通过File(包名).File(结构体名)来访问
type File struct {
	FileFullName string
	Size         int64
	FillLast     bool
	FillSize     int64
	SliceSize    int64
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
	if (file.Size % (int64)(SysConfig.SliceNum)) != 0 {
		file.FillLast = true
		file.SliceSize = file.Size/(int64)(SysConfig.SliceNum) + 1
		file.FillSize = file.SliceSize*(int64)(SysConfig.SliceNum) - file.Size
	} else {
		file.FillLast = false
		file.SliceSize = file.Size / (int64)(SysConfig.SliceNum)
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

//GetFileIndexInAll 利用编号在全部文件列表中获取文件对象（仅用于删除文件时）
func GetFileIndexInAll(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

//InitDataFiles 初始化数据分块
func (file File) InitDataFiles() error {
	source := "./temp/" + file.FileFullName
	//打开原始文件
	sourceFile, err := os.Open(source)
	if err != nil {
		glog.Error("分割时打开源文件失败")
		panic(err)
	}
	defer sourceFile.Close()

	if file.Size <= 1000 {
		//原始文件大小不大于1000 Byte
		for i := 0; i < SysConfig.DataNum; i++ {
			//为每个数据节点拷贝一份副本
			if file.copyFile(true, i, sourceFile) != nil {
				glog.Errorf("生成副本文件失败 i=%d", i)
				panic(err)
			}
		}
	} else {
		//原始文件大于1000 Byte
		for i := 0; i < SysConfig.DataNum; i++ {
			for j := 0; j < SysConfig.RowNum; j++ {
				//生成一个数据分块
				if file.initOneDataFile(i, j, sourceFile) != nil {
					glog.Errorf("生成单个数据文件文件失败 i=%d j=%d", i, j)
					panic(err)
				}
			}
		}
	}

	sourceFile.Close()
	return nil
}

//复制文件，可用于生成数据副本和校验副本
func (file File) copyFile(isData bool, col int, sourceFile *os.File) error {
	//构造副本文件名
	fileName := StructSliceFileName("temp", isData, col, file.FileFullName, col, 0)
	//打开副本文件
	outFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		glog.Error("新建副本文件失败 " + strconv.Itoa(col) + "/" + file.FileFullName + "." + strconv.Itoa(col) + strconv.Itoa(0))
		panic(err)
	}
	defer outFile.Close()

	buffer := make([]byte, file.SliceSize)
	//将原始文件读入buffer
	_, err = sourceFile.Read(buffer)
	if err != nil && err != io.EOF {
		glog.Errorf("buffer读取文件失败 col=%d", col, 0)
		panic(err)
	}

	//将buffer写入副本文件
	if _, err := outFile.Write(buffer[:file.SliceSize]); err != nil {
		glog.Errorf("复制文件失败 col=%d", col)
		panic(err)
	}
	return nil
}

//生成一个数据分块
func (file File) initOneDataFile(col int, row int, sourceFile *os.File) error {
	//构造分块文件名
	fileName := StructSliceFileName("temp", true, col, file.FileFullName, col, row)
	//建立分块文件
	outFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		glog.Error("新建数据分块文件失败 " + "./temp/Data." + strconv.Itoa(col) + "/" + file.FileFullName + "." + strconv.Itoa(col) + strconv.Itoa(row))
		panic(err)
	}
	defer outFile.Close()

	//初始化数据buffer 变量
	buffer := make([]byte, file.SliceSize)

	//判断哪一个分块是原始数据文件的结尾，那么该分块仍需要读取文件，剩下的分块就只需要填充
	fillSliceCount := (int)(file.FillSize/file.SliceSize) + 1

	if file.FillLast && row*SysConfig.DataNum+col == SysConfig.SliceNum-fillSliceCount {
		//需要补充 && 第一个补充分块混合原文件和0

		//先读取原文件中剩余的数据
		n, err := sourceFile.Read(buffer)
		if err != nil && err != io.EOF {
			glog.Errorf("buffer读取文件失败 col=%d row=%d", col, row)
			panic(err)
		}
		//剩余部分用0补齐
		for j := n; (int64)(j) < file.SliceSize; j++ {
			buffer[j] = (byte)(0)
		}
	} else if file.FillLast && row*SysConfig.DataNum+col >= SysConfig.SliceNum-fillSliceCount {
		//需要补充 && (混合点以后 或 第一个即为全0点)
		for j := 0; (int64)(j) < file.FillSize; j++ {
			buffer[j] = (byte)(0)
			if _, err := outFile.Write(buffer[:file.SliceSize]); err != nil {
				glog.Errorf("写入数据分块失败 col=%d row=%d", col, row)
				panic(err)
			}
			return nil
		}
	} else {
		//不需要补充 || 位于混合点以前
		_, err := sourceFile.Read(buffer)
		if err != nil && err != io.EOF {
			glog.Errorf("buffer读取文件失败 col=%d row=%d", col, row)
			panic(err)
		}
	}

	//将buffer 中的数据写入文件
	if _, err := outFile.Write(buffer[:file.SliceSize]); err != nil {
		glog.Errorf("写入数据分块失败 col=%d row=%d", col, row)
		panic(err)
	}
	return nil
}

//SendToNode 将某个文件发送至对应存储节点
func (file File) SendToNode() {
	//发送数据分块
	//todo : 区分是否<=1000
	for i := 0; i < SysConfig.DataNum; i++ {
		//当有节点丢失且当前分块需要发往丢失节点时，直接跳过
		if SysConfig.Status == false && AllNodes[DataNodes[i]].Status == false {
			continue
		}
		for j := 0; j < SysConfig.RowNum; j++ {
			postOneFile(file, true, DataNodes[i], i, j, 0)
		}
	}

	//发送校验分块
	nodeCounter := 0
	fileCounter := 0
	rddtFileCounter := 0
	for xx := 0; xx < SysConfig.FaultNum; xx++ {
		k := (int)((xx + 2) / 2 * (int)(math.Pow(-1, (float64)(xx+2))))
		for fileCounter < SysConfig.DataNum {
			glog.Infof("Rddt Node Num : %d \t k : %d \t fileCounter : %d \t nodeCounter : %d\n", nodeCounter, k, fileCounter, nodeCounter)
			if SysConfig.Status == true && AllNodes[RddtNodes[nodeCounter]].Status == true {
				postOneFile(file, false, RddtNodes[nodeCounter], k, fileCounter, nodeCounter)
			}
			fileCounter++
			rddtFileCounter++
			if rddtFileCounter%(SysConfig.SliceNum/SysConfig.DataNum) == 0 {
				nodeCounter++
				rddtFileCounter = 0
			}
			if fileCounter != SysConfig.DataNum {
				continue
			}
			fileCounter = 0
			break
		}
	}
}

//将某个文件发送至对应存储节点
func postOneFile(file File, isData bool, nodeID uint, posiX, posiY, nodeCounter int) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	var filePath string
	if isData {
		filePath = StructSliceFileName("temp", true, posiX, file.FileFullName, posiX, posiY)
	} else {
		filePath = StructSliceFileName("temp", false, nodeCounter, file.FileFullName, posiX, posiY)
	}
	//编写请求body
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filePath)
	if err != nil {
		glog.Errorf("error writing to buffer + %s", err)
		panic(err)
	}
	//打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		glog.Errorln("error opening file + %s", err)
		panic(err)
	}
	defer fh.Close()
	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	//设定url
	url := "http://" + AllNodes[nodeID].Address.String() + ":" + strconv.Itoa(AllNodes[nodeID].Port) + "/upload"
	resp, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	defer resp.Body.Close()
	//获取response
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	glog.Info(resp.Status)
	glog.Info(string(respbody))
}

//GetFile 将收集到的数据分块写入中心节点中的临时副本文件，用于发送给用户
func (file File) GetFile(targetFolder string) error {
	glog.Infof("[Get File] %s", targetFolder+file.FileFullName)
	target, err := os.Create(targetFolder + file.FileFullName)
	if err != nil {
		glog.Error("无法新建目标文件")
		panic(err)
	}
	defer target.Close()

	if file.Size <= 1000 {
		//直接读取（0，0）处的副本
		buffer := make([]byte, file.SliceSize)
		filePath := StructSliceFileName("temp", true, 0, file.FileFullName, 0, 0)
		dataFile, err := os.Open(filePath)
		if err != nil {
			glog.Error("读取数据分块文件失败 " + filePath)
			panic(err)
		}
		defer dataFile.Close()
		n, err := dataFile.Read(buffer)
		if err != nil && err != io.EOF {
			glog.Error("buffer读取副本失败")
			panic(err)
		}
		if _, err := target.Write(buffer[:n]); err != nil {
			glog.Error("写入副本失败 ")
			panic(err)
		}
	} else {
		var realSliceNum int64
		//确定分块数
		if file.Size%file.SliceSize != 0 {
			realSliceNum = file.Size/file.SliceSize + 1
		} else {
			realSliceNum = file.Size / file.SliceSize
		}
		buffer := make([]byte, file.SliceSize)
		for i := 0; (int64)(i) < realSliceNum; i++ {
			dataPosition := i / SysConfig.RowNum
			rowPosition := i % SysConfig.RowNum
			//读取分块
			filePath := StructSliceFileName("temp", true, dataPosition, file.FileFullName, dataPosition, rowPosition)
			dataFile, err := os.Open(filePath)
			if err != nil {
				glog.Error("读取数据分块文件失败 " + filePath)
				panic(err)
			}
			defer dataFile.Close()
			n, err := dataFile.Read(buffer)
			if err != nil && err != io.EOF {
				glog.Error("buffer读取文件失败 " + strconv.Itoa(i))
				panic(err)
			}

			//判断哪一个分块是原始数据文件的结尾，那么该分块仍需要读取文件，剩下的分块就只需要填充
			fillSliceCount := (int)(file.FillSize/file.SliceSize) + 1
			//todo ： 这里需要去掉的分块不一定是一个
			if file.FillLast && rowPosition*SysConfig.DataNum+dataPosition == SysConfig.SliceNum-fillSliceCount {
				//需要补充 && 第一个补充分块混合原文件和0
				bufferNeeded := file.SliceSize - (file.SliceSize - file.FillSize%file.SliceSize)
				if _, err := target.Write(buffer[:bufferNeeded]); err != nil {
					glog.Error("写入数据分块失败 " + strconv.Itoa(i))
					panic(err)
				}
				return nil
			}
			//不需要补充 || 位于混合点以前
			if _, err := target.Write(buffer[:n]); err != nil {
				glog.Error("写入数据分块失败 " + strconv.Itoa(i))
				panic(err)
			}
		}
	}
	return nil
}

//InitRddtFiles 生成校验文件
func (file File) InitRddtFiles() error {
	if file.Size <= 1000 {
		source := "./temp/" + file.FileFullName
		//打开原始文件
		sourceFile, err := os.Open(source)
		if err != nil {
			glog.Error("备份原始文件副本时打开源文件失败")
			panic(err)
		}
		defer sourceFile.Close()
		for i := 0; i < SysConfig.DataNum; i++ {
			//为每个数据节点拷贝一份副本
			if file.copyFile(false, i, sourceFile) != nil {
				glog.Errorf("生成副本文件失败 i=%d", i)
				panic(err)
			}
		}
	} else {
		rddtFolderCounter := 0
		rddtRowCounter := 0
		for faultCount := 0; faultCount < SysConfig.FaultNum; faultCount++ {
			k := (int)((faultCount + 2) / 2 * (int)(math.Pow(-1, (float64)(faultCount+2))))
			for i := 0; i < SysConfig.DataNum; i++ {
				if rddtRowCounter%SysConfig.RowNum == 0 && (i != 0 || faultCount != 0) {
					rddtRowCounter = 1
					rddtFolderCounter++
				} else {
					rddtRowCounter++
				}
				file.initOneRddtFile(i, k, rddtFolderCounter)
			}
		}
	}
	return nil
}

//具体编码生成某个校验文件
func (file File) initOneRddtFile(startFolderNum, k, rddtNum int) error {
	filePath := StructSliceFileName("temp", false, rddtNum, file.FileFullName, k, startFolderNum)
	rddtFileObj, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)

	if err != nil {
		glog.Error("新建冗余分块文件失败 " + filePath)
		panic(err)
	}
	defer rddtFileObj.Close()

	buffer := make([]byte, file.SliceSize)

	for i := 0; i < SysConfig.RowNum; i++ {
		folderPosi := startFolderNum + k*i
		if folderPosi >= SysConfig.DataNum {
			folderPosi = folderPosi - SysConfig.DataNum
		} else if folderPosi < 0 {
			folderPosi = SysConfig.DataNum + folderPosi
		}
		filePath := StructSliceFileName("temp", true, folderPosi, file.FileFullName, folderPosi, i)
		sourceFile, err := os.Open(filePath)
		if err != nil {
			glog.Error("生成冗余文件 " + filePath)
			panic(err)
		}
		defer sourceFile.Close()
		if i == 0 {
			_, err = sourceFile.Read(buffer)
			if err != nil && err != io.EOF {
				glog.Error("buffer读取文件失败 " + strconv.Itoa(i))
				panic(err)
			}
		} else {
			tempBytes := make([]byte, file.SliceSize)
			_, err = sourceFile.Read(tempBytes)
			if err != nil && err != io.EOF {
				glog.Error("tempBuffer读取文件失败" + strconv.Itoa(i))
				panic(err)
			}
			for byteCounter := 0; byteCounter < len(buffer); byteCounter++ {
				buffer[byteCounter] = buffer[byteCounter] ^ tempBytes[byteCounter]
			}
		}
		//循环到最后才写入文件
		if i == SysConfig.RowNum-1 {
			if _, err := rddtFileObj.Write(buffer[:file.SliceSize]); err != nil {
				glog.Error("写入冗余分块文件失败 " + filePath)
				panic(err)
			}
		}
	}
	return nil
}

//DeleteSlices 处理用户删除文件的请求
func (file File) DeleteSlices() {
	for i := 0; i < SysConfig.DataNum; i++ {
		if file.Size <= 1000 {
			//只需删除第一个
			deleteOneFile(file, true, DataNodes[i], i, 0)
		} else {
			for j := 0; j < SysConfig.RowNum; j++ {
				deleteOneFile(file, true, DataNodes[i], i, j)
			}
		}
	}

	if file.Size <= 1000 {
		for i := 0; i < SysConfig.RddtNum; i++ {
			//只需删除第一个
			deleteOneFile(file, false, RddtNodes[i], i, 0)
		}
	} else {
		nodeCounter := 0
		fileCounter := 0
		rddtFileCounter := 0
		for fCounter := 0; fCounter < SysConfig.FaultNum; fCounter++ {
			k := (int)((fCounter + 2) / 2 * (int)(math.Pow(-1, (float64)(fCounter+2))))
			for fileCounter < SysConfig.DataNum {
				glog.Infof("Rddt Node Num : %d \t k : %d \t fileCounter : %d \t nodeCounter : %d\n", nodeCounter, k, fileCounter, nodeCounter)
				//执行删除
				deleteOneFile(file, false, RddtNodes[nodeCounter], k, fileCounter)
				fileCounter++
				rddtFileCounter++
				if rddtFileCounter%(SysConfig.SliceNum/SysConfig.DataNum) == 0 {
					nodeCounter++
					rddtFileCounter = 0
				}
				if fileCounter != SysConfig.DataNum {
					continue
				}
				fileCounter = 0
				break
			}
		}
	}
}

//执行删除存储节点中某个文件的操作
func deleteOneFile(file File, isData bool, nodeID uint, posiX, posiY int) {
	var fileName string
	if isData {
		fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	} else {
		fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	}
	//设定请求发送url
	url := "http://" + AllNodes[nodeID].Address.String() + ":" + strconv.Itoa(AllNodes[nodeID].Port) + "/delete"
	glog.Info("[DELETE] URL " + url)

	//设定发送至存储节点的删除请求
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	//设定请求header
	req.Header.Set("fileName", fileName)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	glog.Info(resp.Status)
	glog.Info(string(respbody))
}

//DeleteAllTempFiles 删除中心节点中某个文件的全部临时文件
func (file File) DeleteAllTempFiles() error {
	file.deleteTempDataFiles()
	file.deleteTempRddtFiles()
	//删除原始文件副本
	if !FileExisted("temp/" + file.FileFullName) {
		glog.Warningf("[File to Delete NOT EXIST] temp/" + file.FileFullName)
	} else {
		err := os.Remove("temp/" + file.FileFullName)
		if err != nil {
			glog.Errorln(err)
		} else {
			glog.Infof("[File to Delete] temp/" + file.FileFullName)
		}
	}
	return nil
}

func (file File) deleteTempDataFiles() error {
	for i := 0; i < SysConfig.DataNum; i++ {
		if file.Size <= 1000 {
			//删除副本
			if !FileExisted("temp/Data." + strconv.Itoa(i) + "/" + file.FileFullName + "." + strconv.Itoa(i) + strconv.Itoa(0)) {
				glog.Warningf("[File to Delete NOT EXIST] temp/Data.%d/%s.%d%d ", i, file.FileFullName, i, 0)
			} else {
				err := os.Remove("temp/Data." + strconv.Itoa(i) + "/" + file.FileFullName + "." + strconv.Itoa(i) + strconv.Itoa(0))
				if err != nil {
					glog.Errorln(err)
				} else {
					glog.Infof("[File to Delete] temp/Data.%d/%s.%d%d ", i, file.FileFullName, i, 0)
				}
			}
		} else {
			for j := 0; j < SysConfig.RowNum; j++ {
				if !FileExisted("temp/Data." + strconv.Itoa(i) + "/" + file.FileFullName + "." + strconv.Itoa(i) + strconv.Itoa(j)) {
					glog.Warningf("[File to Delete NOT EXIST] temp/Data.%d/%s.%d%d ", i, file.FileFullName, i, j)
				} else {
					err := os.Remove("temp/Data." + strconv.Itoa(i) + "/" + file.FileFullName + "." + strconv.Itoa(i) + strconv.Itoa(j))
					if err != nil {
						glog.Errorln(err)
					} else {
						glog.Infof("[File to Delete] temp/Data.%d/%s.%d%d ", i, file.FileFullName, i, j)
					}
				}
			}
		}
	}
	return nil
}

func (file File) deleteTempRddtFiles() error {
	if file.Size <= 1000 {
		for i := 0; i < SysConfig.RddtNum; i++ {
			if !FileExisted("temp/Rddt." + strconv.Itoa(i) + "/" + file.FileFullName + "." + strconv.Itoa(i) + strconv.Itoa(0)) {
				glog.Warningf("[File to Delete NOT EXIST] temp/Rddt.%d/%s.%d%d ", i, file.FileFullName, i, 0)
			} else {
				err := os.Remove("temp/Rddt." + strconv.Itoa(i) + "/" + file.FileFullName + "." + strconv.Itoa(i) + strconv.Itoa(0))
				if err != nil {
					glog.Errorln(err)
				} else {
					glog.Infof("[File to Delete] temp/RDTT.%d/%s.%d%d ", i, file.FileFullName, i, 0)
				}
			}
		}
	} else {
		nodeCounter := 0
		fileCounter := 0
		rddtFileCounter := 0
		for i := 0; i < SysConfig.FaultNum; i++ {
			k := (int)((i + 2) / 2 * (int)(math.Pow(-1, (float64)(i+2))))
			for fileCounter < SysConfig.DataNum {
				if !FileExisted("temp/Rddt." + strconv.Itoa(nodeCounter) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(fileCounter)) {
					glog.Warningf("[File to Delete NOT EXIST] temp/Rddt.%d/%s.%d%d ", nodeCounter, file.FileFullName, k, fileCounter)
				} else {
					err := os.Remove("temp/Rddt." + strconv.Itoa(nodeCounter) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(fileCounter))
					if err != nil {
						glog.Errorln(err)
					} else {
						glog.Infof("[File to Delete] temp/RDTT.%d/%s.%d%d ", nodeCounter, file.FileFullName, k, fileCounter)
					}
				}
				fileCounter++
				rddtFileCounter++
				if rddtFileCounter%(SysConfig.SliceNum/SysConfig.DataNum) == 0 {
					nodeCounter++
					rddtFileCounter = 0
				}
				if fileCounter != SysConfig.DataNum {
					continue
				}
				fileCounter = 0
				break
			}
		}
	}

	return nil
}

//CollectFiles 从存储节点收集全部分块文件到中心节点
func (file File) CollectFiles() {
	for i := 0; i < SysConfig.DataNum; i++ {
		if file.Size <= 1000 {
			getOneFile(file, true, DataNodes[i], i, 0, 0)
		} else {
			for j := 0; j < SysConfig.RowNum; j++ {
				glog.Infof("Collect Files Data Node Status : %t, ID : %d", AllNodes[DataNodes[i]].Status, DataNodes[i])
				if AllNodes[DataNodes[i]].Status == true {
					getOneFile(file, true, DataNodes[i], i, j, 0)
				}
			}
		}
	}

	if file.Size <= 1000 {
		for i := 0; i < SysConfig.RddtNum; i++ {
			//只需要第一个
			getOneFile(file, false, RddtNodes[i], i, 0, i)
		}
	} else {
		nodeCounter := 0
		fileCounter := 0
		rddtFileCounter := 0
		for fCounter := 0; fCounter < SysConfig.FaultNum; fCounter++ {
			k := (int)((fCounter + 2) / 2 * (int)(math.Pow(-1, (float64)(fCounter+2))))
			for fileCounter < SysConfig.DataNum {
				glog.Infof("从节点获取冗余文件 Rddt Node Num : %d \t k : %d \t fileCounter : %d \t nodeCounter : %d\n", nodeCounter, k, fileCounter, nodeCounter)
				if AllNodes[RddtNodes[nodeCounter]].Status == true {
					getOneFile(file, false, RddtNodes[nodeCounter], k, fileCounter, nodeCounter)
				}
				fileCounter++
				rddtFileCounter++
				if rddtFileCounter%(SysConfig.SliceNum/SysConfig.DataNum) == 0 {
					nodeCounter++
					rddtFileCounter = 0
				}
				if fileCounter != SysConfig.DataNum {
					continue
				}
				fileCounter = 0
				break
			}
		}
	}
}

func getOneFile(file File, isData bool, nodeID uint, posiX, posiY, nodeCounter int) {
	var filePath string
	var fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	if isData {
		filePath = StructSliceFileName("temp", true, posiX, file.FileFullName, posiX, posiY)
	} else {
		filePath = StructSliceFileName("temp", false, nodeCounter, file.FileFullName, posiX, posiY)
	}

	glog.Infof("从节点收集文件 filePath : %s, fileName : %s", filePath, fileName)

	url := "http://" + AllNodes[nodeID].Address.String() + ":" + strconv.Itoa(AllNodes[nodeID].Port) + "/download/" + fileName
	glog.Infof("获取文件的URL : %s", url)
	res, _ := http.Get(url)
	fileGet, _ := os.Create(filePath)
	defer fileGet.Close()
	io.Copy(fileGet, res.Body)
	glog.Info(res.Status)
}

//LostHandle 系统恢复入口
//todo : 修改恢复策略
func LostHandle() {
	glog.Infoln("开始执行 tool.LostHandle() ")
	for _, file := range AllFiles {
		// go func() {
		glog.Infof("开始恢复文件 : %s", file.FileFullName)
		// Collect Distributed Files
		file.CollectFiles()
		var recFinish = true
		for index := range LostNodes {
			var result bool
			glog.Infof("需要检测节点 ID : %d", LostNodes[index])
			result = AllNodes[LostNodes[index]].DetectNode(file)
			recFinish = recFinish && result
		}
		for !recFinish {
			recFinish = true
			for index := range LostNodes {
				var result bool
				glog.Infof("需要检测节点 ID : %d", LostNodes[index])
				result = AllNodes[LostNodes[index]].DetectNode(file)
				recFinish = recFinish && result
			}
		}
		file.InitRddtFiles()
		file.GetFile("temp/")
		file.SendToNode()
		// }()
	}
	//todo : 清除所有丢失节点？
}

//DetectDataFile 检测数据分块是否全部恢复
func (file File) DetectDataFile(node Node, faultCount, targetRow int) bool {
	var NodeDataNum = node.getIndexInDataNodes()
	var k = (int)((faultCount + 2) / 2 * (int)(math.Pow(-1, (float64)(faultCount+2))))
	var startNodeNum = (node.getIndexInDataNodes() - targetRow*k + len(DataNodes)) % len(DataNodes)
	var rddtNodeNum = (faultCount*len(DataNodes) + startNodeNum) / SysConfig.RowNum
	glog.Info("Detecting Data File : " + "temp/Data." + strconv.Itoa(NodeDataNum) + "/" + file.FileFullName + "." + strconv.Itoa(NodeDataNum) + strconv.Itoa(targetRow))
	if FileExisted("temp/Data." + strconv.Itoa(NodeDataNum) + "/" + file.FileFullName + "." + strconv.Itoa(NodeDataNum) + strconv.Itoa(targetRow)) {
		return true
	}
	if !file.DetectRddtFile(AllNodes[RddtNodes[rddtNodeNum]], k, startNodeNum) {
		return false
	}
	if !file.DetectKLine(NodeDataNum, targetRow, rddtNodeNum, k, false) {
		return false
	}
	file.RestoreDataFile(NodeDataNum, rddtNodeNum, k, targetRow)
	return true
}

//DetectRddtFile 检测所有校验分块是否全部恢复
func (file File) DetectRddtFile(node Node, k, dataNodeNum int) bool {
	if FileExisted("temp/Rddt." + strconv.Itoa(node.getIndexInRddtNodes()) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(dataNodeNum)) {
		return true
	}
	if file.DetectKLine(dataNodeNum, 0, node.getIndexInRddtNodes(), k, true) {
		file.initOneRddtFile(dataNodeNum, k, node.getIndexInRddtNodes())
		return true
	}
	return false
}

//DetectKLine 检测某条码链是否可用于恢复
func (file File) DetectKLine(dataNodeNum, targetRow, rddtNodeNum, k int, isForRddt bool) bool {
	var result = true
	var startIndex = (dataNodeNum - k*targetRow + len(DataNodes)) % len(DataNodes)
	var tempResult = true
	for rowCount := 0; rowCount < SysConfig.RowNum; rowCount++ {
		if (targetRow == rowCount) && !isForRddt {
			continue
		}

		tempResult = FileExisted("temp/Data." + strconv.Itoa(startIndex+rowCount*k+len(DataNodes)%len(DataNodes)) + "/" + file.FileFullName + "." + strconv.Itoa(startIndex+rowCount*k+len(DataNodes)%len(DataNodes)) + strconv.Itoa(rowCount))
		result = result && tempResult
	}
	if isForRddt {
		return result
	}
	tempResult = FileExisted("temp/Rddt." + strconv.Itoa(rddtNodeNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startIndex))
	result = result && tempResult
	return result
}

//RestoreDataFile 解码恢复单个数据分块
func (file File) RestoreDataFile(dataNodeNum, rddtNodeNum, k, targetRow int) {
	var startIndex = (dataNodeNum - k*targetRow + len(DataNodes)) % len(DataNodes)
	buffer := make([]byte, file.SliceSize)
	filePath := StructSliceFileName("temp", false, rddtNodeNum, file.FileFullName, k, startIndex)

	rddtFile, err := os.Open(filePath)
	if err != nil {
		glog.Error("打开Rddt源文件失败 " + filePath)
		panic(err)
	}
	defer rddtFile.Close()

	_, err = rddtFile.Read(buffer)
	if err != nil && err != io.EOF {
		glog.Error("RDDT 读取文件失败" + filePath)
		panic(err)
	}

	for rowCount := 0; rowCount < SysConfig.RowNum; rowCount++ {
		if targetRow == rowCount {
			continue
		}

		dataPosi := (startIndex + rowCount*k + len(DataNodes)) % len(DataNodes)
		filePath := StructSliceFileName("temp", true, dataPosi, file.FileFullName, dataPosi, rowCount)
		tempBytes := make([]byte, file.SliceSize)

		dataFile, err := os.Open(filePath)
		_, err = dataFile.Read(tempBytes)
		if err != nil && err != io.EOF {
			glog.Error("tempBuffer 读取 Data 文件失败" + filePath)
			panic(err)
		}
		defer dataFile.Close()

		glog.Infof("恢复中 len(buffer) : %d, len(tempBuffer) : %d, slicesize : %d", len(buffer), len(tempBytes), file.SliceSize)

		for byteCounter := 0; byteCounter < len(buffer); byteCounter++ {
			buffer[byteCounter] = buffer[byteCounter] ^ tempBytes[byteCounter]
		}
	}

	filePath = StructSliceFileName("temp", true, dataNodeNum, file.FileFullName, dataNodeNum, targetRow)

	targetFilePath, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		glog.Error("新建恢复数据冗余分块文件失败 " + filePath)
		panic(err)
	}
	defer targetFilePath.Close()

	if _, err := targetFilePath.Write(buffer[:file.SliceSize]); err != nil {
		glog.Error("写入冗余分块文件失败 " + "./temp/Data." + strconv.Itoa(dataNodeNum) + "/" + file.FileFullName + "." + strconv.Itoa(dataNodeNum) + strconv.Itoa(targetRow))
		panic(err)
	}

}
