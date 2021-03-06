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

	"github.com/Vaaaas/MatrixFS/nodehandler"
	"github.com/Vaaaas/MatrixFS/util"
	"github.com/golang/glog"
)

//复制文件，可用于生成数据副本和校验副本
func (file File) copyFile(isData bool, col int, sourceFile *os.File) error {
	//构造副本文件名
	filePath := structSliceFileName("./temp", isData, col, file.FileFullName, col, 0)
	//打开副本文件
	outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		glog.Error("新建副本文件失败 " + strconv.Itoa(col) + "/" + file.FileFullName + "." + strconv.Itoa(col) + strconv.Itoa(0))
		panic(err)
	}
	defer outFile.Close()
	buffer := make([]byte, file.sliceSize)
	//将原始文件读入buffer
	_, err = sourceFile.Read(buffer)
	if err != nil && err != io.EOF {
		glog.Errorf("buffer读取文件失败 col=%d", col, 0)
		panic(err)
	}
	//将buffer写入副本文件
	if _, err := outFile.Write(buffer[:file.sliceSize]); err != nil {
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
		for i := 0; i < util.SysConfig.DataNum; i++ {
			//为每个数据节点拷贝一份副本
			if file.copyFile(true, i, sourceFile) != nil {
				glog.Errorf("生成副本文件失败 i=%d", i)
				panic(err)
			}
		}
	} else {
		//原始文件大于1000 Byte
		for i := 0; i < util.SysConfig.DataNum; i++ {
			for j := 0; j < util.SysConfig.RowNum; j++ {
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
	filePath := structSliceFileName("./temp", true, col, file.FileFullName, col, row)
	//建立分块文件
	outFile, err := os.Create(filePath)
	if err != nil {
		glog.Error("新建数据分块文件失败 " + "./temp/Data." + strconv.Itoa(col) + "/" + file.FileFullName + "." + strconv.Itoa(col) + strconv.Itoa(row))
		panic(err)
	}
	defer outFile.Close()
	//初始化数据buffer 变量
	buffer := make([]byte, file.sliceSize)
	//判断哪一个分块是原始数据文件的结尾，那么该分块仍需要读取文件，剩下的分块就只需要填充
	fillSliceCount := (int)(file.fillSize/file.sliceSize) + 1
	if file.fillLast && col*util.SysConfig.RowNum+row == util.SysConfig.SliceNum-fillSliceCount {
		//需要补充 && 第一个补充分块混合原文件和0
		//先读取原文件中剩余的数据
		n, err := sourceFile.Read(buffer)
		if err != nil && err != io.EOF {
			glog.Errorf("buffer读取文件失败 col=%d row=%d", col, row)
			panic(err)
		}
		//剩余部分用0补齐
		for j := n; (int64)(j) < file.sliceSize; j++ {
			buffer[j] = (byte)(0)
		}
	} else if file.fillLast && col*util.SysConfig.RowNum+row > util.SysConfig.SliceNum-fillSliceCount {
		//需要补充 && (混合点以后 或 第一个即为全0点)
		for j := 0; (int64)(j) < file.fillSize; j++ {
			buffer[j] = (byte)(0)
			if _, err := outFile.Write(buffer[:file.sliceSize]); err != nil {
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
	if _, err := outFile.Write(buffer[:file.sliceSize]); err != nil {
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
		for i := 0; i < util.SysConfig.RddtNum; i++ {
			//为每个校验节点拷贝一份副本
			if file.copyFile(false, i, sourceFile) != nil {
				glog.Errorf("生成副本文件失败 i=%d", i)
				panic(err)
			}
		}
	} else {
		rddtFolderCounter := 0
		rddtRowCounter := 0
		for faultCount := 0; faultCount < util.SysConfig.FaultNum; faultCount++ {
			k := (int)((faultCount + 2) / 2 * (int)(math.Pow(-1, (float64)(faultCount+2))))
			for i := 0; i < util.SysConfig.DataNum; i++ {
				if rddtRowCounter%util.SysConfig.RowNum == 0 && (i != 0 || faultCount != 0) {
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
	filePath := structSliceFileName("./temp", false, rddtNodePos, file.FileFullName, k, startNodeNum)
	rddtFileObj, err := os.Create(filePath)
	if err != nil {
		glog.Error("新建冗余分块文件失败 " + filePath)
		panic(err)
	}
	defer rddtFileObj.Close()
	buffer := make([]byte, file.sliceSize)
	for i := 0; i < util.SysConfig.RowNum; i++ {
		folderPosi := startNodeNum + k*i
		if folderPosi >= util.SysConfig.DataNum {
			folderPosi = folderPosi - util.SysConfig.DataNum
		} else if folderPosi < 0 {
			folderPosi = util.SysConfig.DataNum + folderPosi
		}
		filePath := structSliceFileName("./temp", true, folderPosi, file.FileFullName, folderPosi, i)
		sourceFile, err := os.Open(filePath)
		if err != nil {
			glog.Error("生成冗余文件 " + filePath)
			panic(err)
		}
		if i == 0 {
			_, err = sourceFile.Read(buffer)
			if err != nil && err != io.EOF {
				glog.Error("buffer读取文件失败 " + strconv.Itoa(i))
				panic(err)
			}
		} else {
			tempBytes := make([]byte, file.sliceSize)
			_, err = sourceFile.Read(tempBytes)
			if err != nil && err != io.EOF {
				glog.Error("tempBuffer读取文件失败" + strconv.Itoa(i))
				panic(err)
			}
			for byteCounter := 0; byteCounter < len(buffer); byteCounter++ {
				buffer[byteCounter] = buffer[byteCounter] ^ tempBytes[byteCounter]
			}
		}
		sourceFile.Close()
		//循环到最后才写入文件
		if i == util.SysConfig.RowNum-1 {
			if _, err := rddtFileObj.Write(buffer[:file.sliceSize]); err != nil {
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
	for i := 0; i < util.SysConfig.DataNum; i++ {
		//当有节点丢失且当前分块需要发往丢失节点时，直接跳过
		node := nodehandler.AllNodes.Get(nodehandler.DataNodes[i]).(nodehandler.Node)
		if util.SysConfig.Status == false && node.Status == false {
			continue
		}

		if file.Size <= 1000 {
			node := nodehandler.AllNodes.Get(nodehandler.DataNodes[i]).(nodehandler.Node)
			if node.Status == true {
				postOneFile(file, true, nodehandler.DataNodes[i], i, 0, 0)
			}
		} else {
			for j := 0; j < util.SysConfig.RowNum; j++ {
				postOneFile(file, true, nodehandler.DataNodes[i], i, j, 0)
			}
		}
	}
	//发送校验分块
	if file.Size <= 1000 {
		for i := 0; i < util.SysConfig.RddtNum; i++ {
			node := nodehandler.AllNodes.Get(nodehandler.RddtNodes[i]).(nodehandler.Node)
			if node.Status == true {
				postOneFile(file, false, nodehandler.RddtNodes[i], i, 0, 0)
			}
		}
	} else {
		nodeCounter := 0
		fileCounter := 0
		rddtFileCounter := 0
		for xx := 0; xx < util.SysConfig.FaultNum; xx++ {
			k := (int)((xx + 2) / 2 * (int)(math.Pow(-1, (float64)(xx+2))))
			for fileCounter < util.SysConfig.DataNum {
				node := nodehandler.AllNodes.Get(nodehandler.RddtNodes[nodeCounter]).(nodehandler.Node)
				if node.Status == true {
					postOneFile(file, false, nodehandler.RddtNodes[nodeCounter], k, fileCounter, nodeCounter)
				}
				fileCounter++
				rddtFileCounter++
				if rddtFileCounter%(util.SysConfig.SliceNum/util.SysConfig.DataNum) == 0 {
					nodeCounter++
					rddtFileCounter = 0
				}
				if fileCounter != util.SysConfig.DataNum {
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
		filePath = structSliceFileName("./temp", true, posiX, file.FileFullName, posiX, posiY)
	} else {
		filePath = structSliceFileName("./temp", false, nodeCounter, file.FileFullName, posiX, posiY)
	}
	//glog.Infof("[发送文件至节点] %s",filePath)
	//编写请求body
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filePath)
	if err != nil {
		glog.Errorf("error writing to buffer + %s", err)
		panic(err)
	}
	//打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		glog.Errorln("error opening file ",err)
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
	node := nodehandler.AllNodes.Get(nodeID).(nodehandler.Node)
	url := "http://" + node.Address.String() + ":" + strconv.Itoa(node.Port) + "/upload"
	//glog.Warningf("[发送文件至节点] source : %s  URL : %s",filePath,url)
	resp, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	defer resp.Body.Close()
	//获取response
	_, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		glog.Errorf("[PostOneFile] ERROR Status Code : %d", resp.StatusCode)
	}
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
}

//GetFile 将收集到的数据分块写入中心节点中的临时副本文件，用于发送给用户
func (file File) GetFile(targetFolder string) error {
	glog.Infof("[分块写入待提取副本] %s", targetFolder+file.FileFullName)
	target, err := os.Create(targetFolder + file.FileFullName)
	if err != nil {
		glog.Error("无法新建目标文件")
		panic(err)
	}
	defer target.Close()
	if file.Size <= 1000 {
		//直接读取（0，0）处的副本
		buffer := make([]byte, file.sliceSize)
		filePath := structSliceFileName("./temp", true, 0, file.FileFullName, 0, 0)
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
		if file.Size%file.sliceSize != 0 {
			realSliceNum = file.Size/file.sliceSize + 1
		} else {
			realSliceNum = file.Size / file.sliceSize
		}
		glog.Infoln("[Compile Slices]")
		buffer := make([]byte, file.sliceSize)
		for i := 0; (int64)(i) < realSliceNum; i++ {
			//数据阵列中的列数，也就是DataNodes中的Index
			dataPosition := i / util.SysConfig.RowNum
			//数据阵列中的行数，也就是RddtNodes中的Index
			rowPosition := i % util.SysConfig.RowNum
			//读取分块
			filePath := structSliceFileName("./temp", true, dataPosition, file.FileFullName, dataPosition, rowPosition)
			dataFile, err := os.Open(filePath)
			if err != nil {
				glog.Error("读取数据分块文件失败 " + filePath)
				panic(err)
			}
			n, err := dataFile.Read(buffer)
			dataFile.Close()
			if err != nil && err != io.EOF {
				glog.Error("buffer读取文件失败 " + strconv.Itoa(i))
				panic(err)
			}
			//判断哪一个分块是原始数据文件的结尾，那么该分块仍需要读取文件，剩下的分块就只需要填充
			fillSliceCount := (int)(file.fillSize/file.sliceSize) + 1
			if file.fillLast && dataPosition*util.SysConfig.RowNum+rowPosition == util.SysConfig.SliceNum-fillSliceCount {
				//需要补充 && 第一个补充分块混合原文件和0
				bufferNeeded := file.Size % file.sliceSize
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
