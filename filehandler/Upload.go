package filehandler

import (
	"bytes"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/Vaaaas/MatrixFS/nodeHandler"
	"github.com/Vaaaas/MatrixFS/sysTool"
	"github.com/golang/glog"
)

//复制文件，可用于生成数据副本和校验副本
func (file File) copyFile(isData bool, col int, sourceFile *os.File) error {
	//构造副本文件名
	filePath := StructSliceFileName("temp", isData, col, file.FileFullName, col, 0)
	//打开副本文件
	outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
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
		for i := 0; i < sysTool.SysConfig.DataNum; i++ {
			//为每个数据节点拷贝一份副本
			if file.copyFile(true, i, sourceFile) != nil {
				glog.Errorf("生成副本文件失败 i=%d", i)
				panic(err)
			}
		}
	} else {
		//原始文件大于1000 Byte
		for i := 0; i < sysTool.SysConfig.DataNum; i++ {
			for j := 0; j < sysTool.SysConfig.RowNum; j++ {
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

//生成一个数据分块
func (file File) initOneDataFile(col int, row int, sourceFile *os.File) error {
	//构造分块文件名
	filePath := StructSliceFileName("temp", true, col, file.FileFullName, col, row)
	//建立分块文件
	outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		glog.Error("新建数据分块文件失败 " + "./temp/Data." + strconv.Itoa(col) + "/" + file.FileFullName + "." + strconv.Itoa(col) + strconv.Itoa(row))
		panic(err)
	}
	defer outFile.Close()
	//初始化数据buffer 变量
	buffer := make([]byte, file.SliceSize)
	//判断哪一个分块是原始数据文件的结尾，那么该分块仍需要读取文件，剩下的分块就只需要填充
	fillSliceCount := (int)(file.FillSize/file.SliceSize) + 1
	if file.FillLast && row*sysTool.SysConfig.DataNum+col == sysTool.SysConfig.SliceNum-fillSliceCount {
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
	} else if file.FillLast && row*sysTool.SysConfig.DataNum+col >= sysTool.SysConfig.SliceNum-fillSliceCount {
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
		for i := 0; i < sysTool.SysConfig.DataNum; i++ {
			//为每个数据节点拷贝一份副本
			if file.copyFile(false, i, sourceFile) != nil {
				glog.Errorf("生成副本文件失败 i=%d", i)
				panic(err)
			}
		}
	} else {
		rddtFolderCounter := 0
		rddtRowCounter := 0
		for faultCount := 0; faultCount < sysTool.SysConfig.FaultNum; faultCount++ {
			k := (int)((faultCount + 2) / 2 * (int)(math.Pow(-1, (float64)(faultCount+2))))
			for i := 0; i < sysTool.SysConfig.DataNum; i++ {
				if rddtRowCounter%sysTool.SysConfig.RowNum == 0 && (i != 0 || faultCount != 0) {
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
func (file File) initOneRddtFile(startNodeNum, k, rddtNodePos int) error {
	filePath := StructSliceFileName("temp", false, rddtNodePos, file.FileFullName, k, startNodeNum)
	rddtFileObj, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		glog.Error("新建冗余分块文件失败 " + filePath)
		panic(err)
	}
	defer rddtFileObj.Close()
	buffer := make([]byte, file.SliceSize)
	for i := 0; i < sysTool.SysConfig.RowNum; i++ {
		folderPosi := startNodeNum + k*i
		if folderPosi >= sysTool.SysConfig.DataNum {
			folderPosi = folderPosi - sysTool.SysConfig.DataNum
		} else if folderPosi < 0 {
			folderPosi = sysTool.SysConfig.DataNum + folderPosi
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
		if i == sysTool.SysConfig.RowNum-1 {
			if _, err := rddtFileObj.Write(buffer[:file.SliceSize]); err != nil {
				glog.Error("写入冗余分块文件失败 " + filePath)
				panic(err)
			}
		}
	}
	return nil
}

//SendToNode 将某个文件发送至对应存储节点
func (file File) SendToNode() {
	//发送数据分块
	for i := 0; i < sysTool.SysConfig.DataNum; i++ {
		if file.Size <= 1000 {
			postOneFile(file, true, nodeHandler.DataNodes[i], i, 0, 0)
		} else {
			//当有节点丢失且当前分块需要发往丢失节点时，直接跳过
			if sysTool.SysConfig.Status == false && nodeHandler.AllNodes[nodeHandler.DataNodes[i]].Status == false {
				continue
			}
			for j := 0; j < sysTool.SysConfig.RowNum; j++ {
				postOneFile(file, true, nodeHandler.DataNodes[i], i, j, 0)
			}
		}
	}
	//发送校验分块
	if file.Size <= 1000 {
		for i := 0; i < sysTool.SysConfig.RddtNum; i++ {
			postOneFile(file, false, nodeHandler.RddtNodes[i], i, 0, 0)
		}
	} else {
		nodeCounter := 0
		fileCounter := 0
		rddtFileCounter := 0
		for xx := 0; xx < sysTool.SysConfig.FaultNum; xx++ {
			k := (int)((xx + 2) / 2 * (int)(math.Pow(-1, (float64)(xx+2))))
			for fileCounter < sysTool.SysConfig.DataNum {
				glog.Infof("Rddt Node Num : %d \t k : %d \t fileCounter : %d \t nodeCounter : %d\n", nodeCounter, k, fileCounter, nodeCounter)
				if sysTool.SysConfig.Status == true && nodeHandler.AllNodes[nodeHandler.RddtNodes[nodeCounter]].Status == true {
					postOneFile(file, false, nodeHandler.RddtNodes[nodeCounter], k, fileCounter, nodeCounter)
				}
				fileCounter++
				rddtFileCounter++
				if rddtFileCounter%(sysTool.SysConfig.SliceNum/sysTool.SysConfig.DataNum) == 0 {
					nodeCounter++
					rddtFileCounter = 0
				}
				if fileCounter != sysTool.SysConfig.DataNum {
					continue
				}
				fileCounter = 0
				break
			}
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
	url := "http://" + nodeHandler.AllNodes[nodeID].Address.String() + ":" + strconv.Itoa(nodeHandler.AllNodes[nodeID].Port) + "/upload"
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
			dataPosition := i / sysTool.SysConfig.RowNum
			rowPosition := i % sysTool.SysConfig.RowNum
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
			if file.FillLast && rowPosition*sysTool.SysConfig.DataNum+dataPosition == sysTool.SysConfig.SliceNum-fillSliceCount {
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
