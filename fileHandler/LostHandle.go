package fileHandler

import (
	"io"
	"math"
	"os"
	"strconv"

	"github.com/golang/glog"
	"github.com/Vaaaas/MatrixFS/sysTool"
	"github.com/Vaaaas/MatrixFS/nodeHandler"
	"google.golang.org/appengine/file"
)

//Old_LostHandle 系统恢复入口
//todo : 修改恢复策略
func LostHandle() bool {
	//todo:LostHandle
	var rddtList []int
	oneFileFinished := true

	for _, file := range AllFiles {
		for index := range nodeHandler.LostNodes {
			finished, isData := nodeHandler.AllNodes[nodeHandler.LostNodes[index]].DetectNode(file)
			if !isData {
				rddtList = append(rddtList, index)
			}
			oneFileFinished = oneFileFinished && finished
		}
	}

	return true
}

func Old_LostHandle() {
	//glog.Infoln("开始执行 sysTool.Old_LostHandle() ")
	//	glog.Infof("开始恢复文件 : %s", file.FileFullName)
	//
	//	//收集剩余分块
	//	//todo : 按需要收集分块
	//	file.CollectFiles()
	//
	//	var recFinish = true
	//	for index := range nodeHandler.LostNodes {
	//		var result bool
	//		glog.Infof("需要检测节点 ID : %d", nodeHandler.LostNodes[index])
	//		result = nodeHandler.AllNodes[nodeHandler.LostNodes[index]].Old_DetectNode(file)
	//		recFinish = recFinish && result
	//	}
	//	for !recFinish {
	//		recFinish = true
	//		for index := range nodeHandler.LostNodes {
	//			var result bool
	//			glog.Infof("需要检测节点 ID : %d", nodeHandler.LostNodes[index])
	//			result = nodeHandler.AllNodes[nodeHandler.LostNodes[index]].Old_DetectNode(file)
	//			recFinish = recFinish && result
	//		}
	//	}
	//	file.InitRddtFiles()
	//	file.GetFile("temp/")
	//	file.SendToNode()
	//todo : 清除所有丢失节点？
}

func (file File) DetectDataFile(node nodeHandler.Node, targetRow int) bool {
	var NodeDataNum = node.GetIndexInDataNodes()
	for fCount := 0; fCount < sysTool.SysConfig.FaultNum; fCount++ {
		k := (int)((fCount + 2) / 2 * (int)(math.Pow(-1, (float64)(fCount+2))))
		var startNodeNum = (node.GetIndexInDataNodes() - targetRow*k + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
		var rddtNodeNum = (fCount*len(nodeHandler.DataNodes) + startNodeNum) / sysTool.SysConfig.RowNum
		glog.Info("Detecting Data File : " + "temp/Data." + strconv.Itoa(NodeDataNum) + "/" + file.FileFullName + "." + strconv.Itoa(NodeDataNum) + strconv.Itoa(targetRow))
		if FileExisted("temp/Data." + strconv.Itoa(NodeDataNum) + "/" + file.FileFullName + "." + strconv.Itoa(NodeDataNum) + strconv.Itoa(targetRow)) {
			return true
		}
		if !file.DetectRddtFile(nodeHandler.AllNodes[nodeHandler.RddtNodes[rddtNodeNum]], k, startNodeNum) {
			return false
		}
		if !file.Old_DetectKLine(NodeDataNum, targetRow, rddtNodeNum, k, false) {
			//todo :DetectDataFile
		}
		//todo:DetectDataFile
	}
	return true
}

//Old_DetectDataFile 检测数据分块是否全部恢复
func (file File) Old_DetectDataFile(node nodeHandler.Node, faultCount, targetRow int) bool {
	var NodeDataNum = node.GetIndexInDataNodes()
	var k = (int)((faultCount + 2) / 2 * (int)(math.Pow(-1, (float64)(faultCount+2))))
	var startNodeNum = (node.GetIndexInDataNodes() - targetRow*k + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
	var rddtNodeNum = (faultCount*len(nodeHandler.DataNodes) + startNodeNum) / sysTool.SysConfig.RowNum
	glog.Info("Detecting Data File : " + "temp/Data." + strconv.Itoa(NodeDataNum) + "/" + file.FileFullName + "." + strconv.Itoa(NodeDataNum) + strconv.Itoa(targetRow))
	if FileExisted("temp/Data." + strconv.Itoa(NodeDataNum) + "/" + file.FileFullName + "." + strconv.Itoa(NodeDataNum) + strconv.Itoa(targetRow)) {
		return true
	}
	if !file.DetectRddtFile(nodeHandler.AllNodes[nodeHandler.RddtNodes[rddtNodeNum]], k, startNodeNum) {
		return false
	}
	if !file.Old_DetectKLine(NodeDataNum, targetRow, rddtNodeNum, k, false) {
		return false
	}
	file.RestoreDataFile(NodeDataNum, rddtNodeNum, k, targetRow)
	return true
}

//DetectRddtFile 检测所有校验分块是否全部恢复
func (file File) DetectRddtFile(node nodeHandler.Node, k, dataNodeNum int) bool {
	if FileExisted("temp/Rddt." + strconv.Itoa(node.GetIndexInRddtNodes()) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(dataNodeNum)) {
		return true
	}
	if file.Old_DetectKLine(dataNodeNum, 0, node.GetIndexInRddtNodes(), k, true) {
		file.initOneRddtFile(dataNodeNum, k, node.GetIndexInRddtNodes())
		return true
	}
	return false
}

func (file File) DetectKLine(dataNodeNum, targetRow, rddtNodeNum, k int, isForRddt bool) bool {
	//todo:DetectKLine

	return true
}

//Old_DetectKLine 检测某条码链是否可用于恢复
func (file File) Old_DetectKLine(dataNodeNum, targetRow, rddtNodeNum, k int, isForRddt bool) bool {
	var result = true
	var startIndex = (dataNodeNum - k*targetRow + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
	var tempResult = true
	for rowCount := 0; rowCount < sysTool.SysConfig.RowNum; rowCount++ {
		if (targetRow == rowCount) && !isForRddt {
			continue
		}

		tempResult = FileExisted("temp/Data." + strconv.Itoa(startIndex+rowCount*k+len(nodeHandler.DataNodes)%len(nodeHandler.DataNodes)) + "/" + file.FileFullName + "." + strconv.Itoa(startIndex+rowCount*k+len(nodeHandler.DataNodes)%len(nodeHandler.DataNodes)) + strconv.Itoa(rowCount))
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
	var startIndex = (dataNodeNum - k*targetRow + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
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

	for rowCount := 0; rowCount < sysTool.SysConfig.RowNum; rowCount++ {
		if targetRow == rowCount {
			continue
		}

		dataPosi := (startIndex + rowCount*k + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
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
