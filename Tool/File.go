package Tool

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

//todo : remove to map
var AllFiles = []File{}

func FindFileInAll(name string) *File {
	for _, tempFile := range AllFiles {
		if tempFile.FileFullName == name {
			return &tempFile
		}
	}
	return nil
}

//使用时需要通过File(包名).File(结构体名)来访问
type File struct {
	FileFullName string
	Size         int64
	FillLast     bool
	FillSize     int64
	SliceSize    int64
}

func FileExisted(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (file *File) Init(source string) error {
	fileInfo, err := os.Stat(source)
	if err != nil {
		glog.Error(err)
		panic(err)
	}
	if fileInfo.IsDir() {
		glog.Error("初始化失败：该路径指向的是文件夹")
		return errors.New("初始化失败：该路径指向的是文件夹")
	} else {
		file.FileFullName = fileInfo.Name()
		file.Size = fileInfo.Size()
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
		AllFiles = append(AllFiles, *file)

		glog.Infof("All Files: %+v  ,len: %d", AllFiles[len(AllFiles)-1], len(AllFiles))

		return nil
	}
}

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
		} else {
			glog.Error("分割文件名失败：该文件不含扩展名")
			return "", ""
		}
	} else {
		glog.Error("分割文件名失败：未定义文件名")
		return "", ""
	}
}

func GetFileIndexInAll(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func (file File) InitDataFiles() error {
	source := "./temp/" + file.FileFullName

	sourceFile, err := os.Open(source)
	if err != nil {
		glog.Error("分割时打开源文件失败")
		panic(err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			panic(err)
		}
	}()

	for i := 0; i < SysConfig.DataNum; i++ {
		for j := 0; j < SysConfig.RowNum; j++ {
			if file.initOneDataFile(i, j, sourceFile) != nil {
				glog.Errorf("生成单个数据文件文件失败 i=%d j=%d", i, j)
				panic(err)
			}
		}
	}
	return nil
}

func (file File) initOneDataFile(col int, row int, sourceFile *os.File) error {
	outFile, err := os.Create("./temp/Data." + strconv.Itoa(col) + "/" + file.FileFullName + "." + strconv.Itoa(col) + strconv.Itoa(row))
	if err != nil {
		glog.Error("新建数据分块文件失败 " + "./temp/Data." + strconv.Itoa(col) + "/" + file.FileFullName + "." + strconv.Itoa(col) + strconv.Itoa(row))
		panic(err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			panic(err)
		}
	}()

	buffer := make([]byte, file.SliceSize)
	n, err := sourceFile.Read(buffer)
	if err != nil && err != io.EOF {
		glog.Errorf("buffer读取文件失败 col=%d row=%d", col, row)
		panic(err)
	}

	if file.FillLast && (int64)(n) != file.SliceSize {
		if (int64)(n) == 0 {
			tempBuffer := make([]byte, file.SliceSize)
			for j := 0; (int64)(j) < file.FillSize; j++ {
				tempBuffer[j] = (byte)(0)
				if _, err := outFile.Write(buffer[:file.SliceSize]); err != nil {
					glog.Errorf("写入数据分块失败 col=%d row=%d", col, row)
					panic(err)
				}
				return nil
			}
		} else {
			for j := n; (int64)(j) < file.SliceSize; j++ {
				buffer[j] = (byte)(0)
			}
		}
	}

	if _, err := outFile.Write(buffer[:file.SliceSize]); err != nil {
		glog.Errorf("写入数据分块失败 col=%d row=%d", col, row)
		panic(err)
	}
	return nil
}

//SendToNode will send one file to Nodes
func (file File) SendToNode() {
	for i := 0; i < SysConfig.DataNum; i++ {
		for j := 0; j < SysConfig.RowNum; j++ {
			postOneFile(file, true, DataNodes[i], i, j, 0)
		}
	}

	nodeCounter := 0
	fileCounter := 0
	rddtFileCounter := 0
	for xx := 0; xx < SysConfig.FaultNum; xx++ {
		k := (int)((xx + 2) / 2 * (int)(math.Pow(-1, (float64)(xx+2))))
		for fileCounter < SysConfig.DataNum {
			glog.Infof("Rddt Node Num : %d \t k : %d \t fileCounter : %d \t nodeCounter : %d\n", nodeCounter, k, fileCounter, nodeCounter)
			postOneFile(file, false, RddtNodes[nodeCounter], k, fileCounter, nodeCounter)
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

func postOneFile(file File, isData bool, nodeID uint, posiX, posiY, nodeCounter int) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	var filePath string
	if isData {
		filePath = "./temp/Data." + strconv.Itoa(posiX) + "/" + file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	} else {
		filePath = "./temp/Rddt." + strconv.Itoa(nodeCounter) + "/" + file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	}

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

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	url := "http://" + AllNodes[nodeID].Address.String() + ":" + strconv.Itoa(AllNodes[nodeID].Port) + "/upload"
	resp, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	glog.Info(resp.Status)
	glog.Info(string(resp_body))
}

func (file File) GetFile(targetFolder string) error {
	glog.Infof("[Get File] %s", targetFolder+file.FileFullName)
	target, err := os.Create(targetFolder + file.FileFullName)
	var realSliceNum int64
	if file.Size%file.SliceSize != 0 {
		realSliceNum = file.Size/file.SliceSize + 1
	} else {
		realSliceNum = file.Size / file.SliceSize
	}
	if err != nil {
		glog.Error("无法新建目标文件")
		panic(err)
	}
	defer func() {
		if err := target.Close(); err != nil {
			panic(err)
		}
	}()

	buffer := make([]byte, file.SliceSize)
	for i := 0; (int64)(i) < realSliceNum; i++ {
		dataPosition := i / SysConfig.RowNum
		rowPosition := i % SysConfig.RowNum

		dataFile, err := os.Open("./temp/Data." + strconv.Itoa(dataPosition) + "/" + file.FileFullName + "." + strconv.Itoa(dataPosition) + strconv.Itoa(rowPosition))
		if err != nil {
			glog.Error("读取数据分块文件失败 " + "./temp/Data." + strconv.Itoa(dataPosition) + "/" + file.FileFullName + "." + strconv.Itoa(dataPosition) + strconv.Itoa(rowPosition))
			panic(err)
		}
		defer func() {
			if err := dataFile.Close(); err != nil {
				panic(err)
			}
		}()
		n, err := dataFile.Read(buffer)
		if err != nil && err != io.EOF {
			glog.Error("buffer读取文件失败 " + strconv.Itoa(i))
			panic(err)
		}

		if file.FillLast && (int64)(i) == realSliceNum-1 {
			if (int64)(file.FillSize) != file.SliceSize {
				if _, err := target.Write(buffer[:((int64)(n) - file.FillSize%file.SliceSize)]); err != nil {
					glog.Error("写入数据分块失败 " + strconv.Itoa(i))
					panic(err)
				}
			}
		} else {
			if _, err := target.Write(buffer[:n]); err != nil {
				glog.Error("写入数据分块失败 " + strconv.Itoa(i))
				panic(err)
			}
		}
	}

	return nil
}

func (file File) InitRddtFiles() error {
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
	return nil
}

func (file File) initOneRddtFile(startFolderNum, k, rddtNum int) error {
	rddtFilePath, err := os.Create("./temp/Rddt." + strconv.Itoa(rddtNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startFolderNum))
	if err != nil {
		glog.Error("新建冗余分块文件失败 " + "./temp/Rddt." + strconv.Itoa(rddtNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startFolderNum))
		panic(err)
	}
	defer func() {
		if err := rddtFilePath.Close(); err != nil {
			panic(err)
		}
	}()

	buffer := make([]byte, file.SliceSize)

	for i := 0; i < SysConfig.RowNum; i++ {
		folderPosi := startFolderNum + k*i
		if folderPosi >= SysConfig.DataNum {
			folderPosi = folderPosi - SysConfig.DataNum
		} else if folderPosi < 0 {
			folderPosi = SysConfig.DataNum + folderPosi
		}

		sourceFile, err := os.Open("./temp/Data." + strconv.Itoa(folderPosi) + "/" + file.FileFullName + "." + strconv.Itoa(folderPosi) + strconv.Itoa(i))
		if err != nil {
			glog.Error("生成冗余文件 " + "./temp/Rddt." + strconv.Itoa(rddtNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startFolderNum) + " 时打开源文件失败 " + "./temp/Data." + strconv.Itoa(folderPosi) + "/" + file.FileFullName + "." + strconv.Itoa(folderPosi) + strconv.Itoa(i))
			panic(err)
		}
		defer func() {
			if err := sourceFile.Close(); err != nil {
				panic(err)
			}
		}()
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
		if i == SysConfig.RowNum-1 {
			if _, err := rddtFilePath.Write(buffer[:file.SliceSize]); err != nil {
				glog.Error("写入冗余分块文件失败 " + "./temp/Rddt." + strconv.Itoa(rddtNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startFolderNum))
				panic(err)
			}
		}
	}
	return nil
}

//todo : Delete all temp files when finished the whole upload operation

func (file File) DeleteSlices() {
	for i := 0; i < SysConfig.DataNum; i++ {
		for j := 0; j < SysConfig.RowNum; j++ {
			deleteOneFile(file, true, DataNodes[i], i, j)
		}
	}

	nodeCounter := 0
	fileCounter := 0
	rddtFileCounter := 0
	for xx := 0; xx < SysConfig.FaultNum; xx++ {
		k := (int)((xx + 2) / 2 * (int)(math.Pow(-1, (float64)(xx+2))))
		for fileCounter < SysConfig.DataNum {
			glog.Infof("Rddt Node Num : %d \t k : %d \t fileCounter : %d \t nodeCounter : %d\n", nodeCounter, k, fileCounter, nodeCounter)
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

func deleteOneFile(file File, isData bool, nodeID uint, posiX, posiY int) {
	var fileName string
	if isData {
		fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	} else {
		fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	}

	url := "http://" + AllNodes[nodeID].Address.String() + ":" + strconv.Itoa(AllNodes[nodeID].Port) + "/delete"
	glog.Info("[DELETE] URL " + url)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	req.Header.Set("fileName", fileName)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}

	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	glog.Info(resp.Status)
	glog.Info(string(resp_body))
}

func (file File) DeleteAllTempFiles() error {
	file.deleteDataFiles()
	file.deleteRddtFiles()
	if FileExisted("temp/" + file.FileFullName) {
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

func (file File) deleteDataFiles() error {
	for i := 0; i < SysConfig.DataNum; i++ {
		for j := 0; j < SysConfig.SliceNum/SysConfig.DataNum; j++ {
			if FileExisted("temp/Data." + strconv.Itoa(i) + "/" + file.FileFullName + "." + strconv.Itoa(i) + strconv.Itoa(j)) {
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

	return nil
}

func (file File) deleteRddtFiles() error {
	nodeCounter := 0
	fileCounter := 0
	rddtFileCounter := 0
	for i := 0; i < SysConfig.FaultNum; i++ {
		k := (int)((i + 2) / 2 * (int)(math.Pow(-1, (float64)(i+2))))
		for fileCounter < SysConfig.DataNum {
			if FileExisted("temp/Rddt." + strconv.Itoa(nodeCounter) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(fileCounter)) {
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

	return nil
}

func (file File) CollectFiles() {
	for i := 0; i < SysConfig.DataNum; i++ {
		for j := 0; j < SysConfig.RowNum; j++ {
			getOneFile(file, true, DataNodes[i], i, j, 0)
		}
	}

	nodeCounter := 0
	fileCounter := 0
	rddtFileCounter := 0
	for xx := 0; xx < SysConfig.FaultNum; xx++ {
		k := (int)((xx + 2) / 2 * (int)(math.Pow(-1, (float64)(xx+2))))
		for fileCounter < SysConfig.DataNum {
			glog.Infof("Rddt Node Num : %d \t k : %d \t fileCounter : %d \t nodeCounter : %d\n", nodeCounter, k, fileCounter, nodeCounter)
			getOneFile(file, false, RddtNodes[nodeCounter], k, fileCounter, nodeCounter)
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

func getOneFile(file File, isData bool, nodeID uint, posiX, posiY, nodeCounter int) {
	var filePath string
	var fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	if isData {
		filePath = "./temp/Data." + strconv.Itoa(posiX) + "/" + fileName
	} else {
		filePath = "./temp/Rddt." + strconv.Itoa(nodeCounter) + "/" + fileName
	}

	url := "http://" + AllNodes[nodeID].Address.String() + ":" + strconv.Itoa(AllNodes[nodeID].Port) + "/download/" + fileName
	res, _ := http.Get(url)
	fileGet, _ := os.Create(filePath)
	io.Copy(fileGet, res.Body)

	glog.Info(res.Status)
}

func LostHandle() {
	for _, file := range AllFiles {
		go func() {
			// Collect Distributed Files
			file.CollectFiles()
			var recFinish = false
			for !recFinish {
				recFinish = true
				//todo : Attention to this expression
				for index := range LostNodes {
					var result bool
					result = AllNodes[LostNodes[index]].DetectNode(file)
					recFinish = recFinish && result
				}
			}
			file.InitRddtFiles()
			file.GetFile("temp/")
			file.SendToNode()
		}()
	}
	//todo : Clear All Lost Nodes

}

func (file File) DetectDataFile(node Node, faultCount, targetRow int) bool {
	var NodeDataNum = node.getIndexInDataNodes()
	var k = (int)((faultCount + 2) / 2 * (int)(math.Pow(-1, (float64)(faultCount+2))))
	var startNodeNum = (node.getIndexInDataNodes() - targetRow*k + len(DataNodes)) % len(DataNodes)
	var rddtNodeNum = (faultCount*len(DataNodes) + startNodeNum) / SysConfig.RowNum
	glog.Info("Detecting Data File : " + "temp/Data." + strconv.Itoa(NodeDataNum) + "/" + file.FileFullName + "." + strconv.Itoa(NodeDataNum) + strconv.Itoa(targetRow))
	if FileExisted("temp/Data." + strconv.Itoa(NodeDataNum) + "/" + file.FileFullName + "." + strconv.Itoa(NodeDataNum) + strconv.Itoa(targetRow)) {
		if file.DetectRddtFile(AllNodes[RddtNodes[rddtNodeNum]], k, startNodeNum) {
			if file.DetectKLine(NodeDataNum, targetRow, rddtNodeNum, k, false) {
				file.RestoreDataFile(NodeDataNum, rddtNodeNum, k, targetRow)
			} else {
				return false
			}
		} else {
			return false
		}
	} else {
		return true
	}
	return false
}

func (file File) DetectRddtFile(node Node, k, dataNodeNum int) bool {
	if FileExisted("temp/Rddt." + strconv.Itoa(node.getIndexInRddtNodes()) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(dataNodeNum)) {
		return true
	}
	if file.DetectKLine(dataNodeNum, 0, node.getIndexInRddtNodes(), k, true) {
		file.initOneRddtFile(dataNodeNum, k, node.getIndexInRddtNodes())
		return true
	} else {
		return false
	}
}

func (file File) DetectKLine(dataNodeNum, targetRow, rddtNodeNum, k int, isForRddt bool) bool {
	var result = true
	var startIndex = (dataNodeNum - k*targetRow + len(DataNodes)) % len(DataNodes)
	var tempResult = true
	for rowCount := 0; rowCount < SysConfig.RowNum; rowCount++ {
		if (targetRow == rowCount) && !isForRddt {
			continue
		}

		tempResult = FileExisted("temp/Data." + strconv.Itoa(startIndex+rowCount*k+len(DataNodes)%len(DataNodes)) + "/" + file.FileFullName + "." + strconv.Itoa((startIndex + rowCount*k + len(DataNodes)%len(DataNodes))) + strconv.Itoa(rowCount))
		result = result && tempResult
	}
	if isForRddt {
		return result
	}
	tempResult = FileExisted("temp/Rddt." + strconv.Itoa(rddtNodeNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startIndex))
	result = result && tempResult
	return result
}

func (file File) RestoreDataFile(dataNodeNum, rddtNodeNum, k, targetRow int) {
	var startIndex = (dataNodeNum - k*targetRow + len(DataNodes)) % len(DataNodes)
	buffer := make([]byte, file.SliceSize)
	rddtFile, err := os.Open("./temp/Rddt." + strconv.Itoa(rddtNodeNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startIndex))
	if err != nil {
		glog.Error("打开源文件失败 " + "./temp/Rddt." + strconv.Itoa(rddtNodeNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startIndex))
		panic(err)
	}
	defer func() {
		if err := rddtFile.Close(); err != nil {
			panic(err)
		}
	}()
	_, err = rddtFile.Read(buffer)
	if err != nil && err != io.EOF {
		glog.Error("RDDT 读取文件失败")
		panic(err)
	}

	for rowCount := 0; rowCount < SysConfig.RowNum; rowCount++ {
		if targetRow == rowCount {
			continue
		}
		tempBytes := make([]byte, file.SliceSize)
		dataPosi := (startIndex + rowCount*k + len(DataNodes)) % len(DataNodes)
		dataFile, err := os.Open("temp/Data." + strconv.Itoa(dataPosi) + "/" + file.FileFullName + "." + strconv.Itoa(dataPosi) + strconv.Itoa(rowCount))
		_, err = dataFile.Read(tempBytes)
		if err != nil && err != io.EOF {
			glog.Error("tempBuffer读取 Data 文件失败")
			panic(err)
		}
		for byteCounter := 0; byteCounter < len(buffer); byteCounter++ {
			buffer[byteCounter] = buffer[byteCounter] ^ tempBytes[byteCounter]
		}
	}

	targetFilePath, err := os.Create("./temp/Data." + strconv.Itoa(dataNodeNum) + "/" + file.FileFullName + "." + strconv.Itoa(dataNodeNum) + strconv.Itoa(targetRow))
	if err != nil {
		glog.Error("新建恢复数据冗余分块文件失败 " + "./temp/Data." + strconv.Itoa(dataNodeNum) + "/" + file.FileFullName + "." + strconv.Itoa(dataNodeNum) + strconv.Itoa(targetRow))
		panic(err)
	}
	defer func() {
		if err := targetFilePath.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := targetFilePath.Write(buffer[:file.SliceSize]); err != nil {
		glog.Error("写入冗余分块文件失败 " + "./temp/Data." + strconv.Itoa(dataNodeNum) + "/" + file.FileFullName + "." + strconv.Itoa(dataNodeNum) + strconv.Itoa(targetRow))
		panic(err)
	}

}
