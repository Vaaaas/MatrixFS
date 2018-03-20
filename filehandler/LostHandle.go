package filehandler

import (
	"io"
	"math"
	"os"
	"strconv"

	"github.com/Vaaaas/MatrixFS/nodeHandler"
	"github.com/Vaaaas/MatrixFS/sysTool"
	"github.com/golang/glog"
)

//TODO: 修改恢复策略

//LostHandle 执行解码恢复功能
func (file *File) LostHandle() bool {
	glog.Infoln("开始执行 LostHandle() ")
	glog.Infof("开始恢复文件 : %s", file.FileFullName)

	var dataNodes []uint
	//找出所有数据节点
	for col := 0; col < len(nodeHandler.LostNodes); col++ {
		glog.Infof("需要检测节点 ID : %d", nodeHandler.LostNodes[col])
		if nodeHandler.AllNodes[nodeHandler.LostNodes[col]].IsDataNode() {
			dataNodes = append(dataNodes, nodeHandler.LostNodes[col])
		}
	}
	if len(dataNodes) != 0 {
		//有数据节点丢失
		var recFinish = true
		for row := 0; row < sysTool.SysConfig.RowNum/2+1; row++ {
			var rowResult = true
			for col := 0; col < len(dataNodes); col++ {
				//检测并恢复单个文件
				colResult := file.DetectDataFile(nodeHandler.AllNodes[dataNodes[col]], row)
				rowResult = rowResult && colResult
			}
			recFinish = recFinish && rowResult
		}
		if !recFinish {
			recFinish = true
			for !recFinish {
				for row := 0; row < sysTool.SysConfig.RowNum/2+1; row++ {
					var rowResult = true
					for col := 0; col < len(dataNodes); col++ {
						//检测并恢复单个文件
						colResult := file.DetectDataFile(nodeHandler.AllNodes[dataNodes[col]], row)
						rowResult = rowResult && colResult
					}
					recFinish = recFinish && rowResult
				}
			}
		}
	} else {
		//丢失的全部是校验节点，直接生成校验分块
		//TODO: (按需从节点提取文件)

	}
	return true
}

//DetectDataFile 检测单个数据文件，如果不在中心节点，则依次检测对应校验文件，码链上的所有数据分块，以判断是否可恢复，可以则恢复
func (file File) DetectDataFile(node nodeHandler.Node, targetRow int) bool {
	var dataNodePos = node.GetIndexInDataNodes()
	filePath := StructSliceFileName("temp", true, dataNodePos, file.FileFullName, dataNodePos, targetRow)
	if FileExistedInCenter(filePath) {
		return true
	}
	for fCount := 0; fCount < sysTool.SysConfig.FaultNum; fCount++ {
		k := (int)((fCount + 2) / 2 * (int)(math.Pow(-1, (float64)(fCount+2))))
		var startDataNodePos = (node.GetIndexInDataNodes() - targetRow*k + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
		var rddtNodePos = (fCount*len(nodeHandler.DataNodes) + startDataNodePos) / sysTool.SysConfig.RowNum
		glog.Info("Detecting Data File : " + "temp/Data." + strconv.Itoa(dataNodePos) + "/" + file.FileFullName + "." + strconv.Itoa(dataNodePos) + strconv.Itoa(targetRow))
		//首先判断本地是否有该文件
		if !file.DetectRddtFile(rddtNodePos, k, startDataNodePos) {
			//当前k对应的校验文件不在中心节点
			//尝试从存储节点下载该校验分块
			if !getOneFile(file, false, nodeHandler.RddtNodes[rddtNodePos], k, startDataNodePos, rddtNodePos) {
				//未取得校验文件
				if fCount == sysTool.SysConfig.FaultNum-1 {
					//直到最后一种斜率也不行
					return false
				}
				continue
			}
		}
		if !file.DetectKLine(dataNodePos, targetRow, rddtNodePos, k) {
			if fCount == sysTool.SysConfig.FaultNum-1 {
				//直到最后一种斜率也不行
				return false
			}
			continue
		}
		file.RestoreDataFile(dataNodePos, rddtNodePos, k, targetRow)
		//TODO: 同时生成对称Data分块
		//收集对称数据分块对应的校验文件和码链
		pairTargetRow:=sysTool.SysConfig.RowNum-targetRow-1
		if pairTargetRow!=targetRow{
			var pairRddtNodePos int
			var pairStartDataNodePos = (node.GetIndexInDataNodes() - pairTargetRow*(-k) + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
			if(k>0){
				pairRddtNodePos=((fCount+1)*len(nodeHandler.DataNodes) + pairStartDataNodePos) / sysTool.SysConfig.RowNum
			}else{
				pairRddtNodePos=((fCount-1)*len(nodeHandler.DataNodes) + pairStartDataNodePos) / sysTool.SysConfig.RowNum
			}
			if getOneFile(file, false, nodeHandler.RddtNodes[pairRddtNodePos], -k, pairStartDataNodePos, pairRddtNodePos){
				if file.DetectKLine(pairStartDataNodePos,pairTargetRow,pairRddtNodePos,-k){
					file.RestoreDataFile(pairStartDataNodePos, pairRddtNodePos, -k, pairTargetRow)
				}
			}
		}
		return true
	}
	return false
}

//DetectRddtFile 检测所有校验分块是否全部恢复
func (file File) DetectRddtFile(rddtNodePos, k, dataNodePos int) bool {
	filePath := StructSliceFileName("temp", false, rddtNodePos, file.FileFullName, k, dataNodePos)
	if !FileExistedInCenter(filePath) {
		if !FileExistInNode(file, false, nodeHandler.AllNodes[nodeHandler.RddtNodes[rddtNodePos]].ID, k, dataNodePos, rddtNodePos) {
			return false
		}
	}
	return true
}

//DetectKLine 检测一条码链上的所有数据文件
func (file File) DetectKLine(dataNodePos, targetRow, rddtNodePos, k int) bool {
	//TODO:DetectKLine
	var result = true
	var startIndex = (dataNodePos - k*targetRow + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
	for rowCount := 0; rowCount < sysTool.SysConfig.RowNum; rowCount++ {
		if targetRow == rowCount {
			continue
		}
		filePath := StructSliceFileName("temp", true, (startIndex+rowCount*k+len(nodeHandler.DataNodes))%len(nodeHandler.DataNodes), file.FileFullName, startIndex+rowCount*k+len(nodeHandler.DataNodes)%len(nodeHandler.DataNodes), rowCount)
		//先看中心节点有没有
		resultCenter := FileExistedInCenter(filePath)
		if !resultCenter {
			//中心节点没有
			resultNode := FileExistInNode(file, true,nodeHandler.DataNodes[(startIndex+rowCount*k+len(nodeHandler.DataNodes))%len(nodeHandler.DataNodes)], dataNodePos, targetRow,rddtNodePos)
			if !resultNode{
				//也没能从存储节点提取，那这条码链就至少缺一个文件
				return false
			}
			result = result && resultNode
			continue
		}
		result = result && resultCenter
	}
	return result
}

//TODO: 是否需要修改

//RestoreDataFile 解码恢复单个数据分块
func (file File) RestoreDataFile(dataNodePos, rddtNodePos, k, targetRow int) {
	var startIndex = (dataNodePos - k*targetRow + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
	buffer := make([]byte, file.SliceSize)
	filePath := StructSliceFileName("temp", false, rddtNodePos, file.FileFullName, k, startIndex)
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
		dataPos := (startIndex + rowCount*k + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
		filePath := StructSliceFileName("temp", true, dataPos, file.FileFullName, dataPos, rowCount)
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
	filePath = StructSliceFileName("temp", true, dataNodePos, file.FileFullName, dataNodePos, targetRow)
	targetFilePath, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		glog.Error("新建恢复数据冗余分块文件失败 " + filePath)
		panic(err)
	}
	defer targetFilePath.Close()
	if _, err := targetFilePath.Write(buffer[:file.SliceSize]); err != nil {
		glog.Error("写入冗余分块文件失败 " + "./temp/Data." + strconv.Itoa(dataNodePos) + "/" + file.FileFullName + "." + strconv.Itoa(dataNodePos) + strconv.Itoa(targetRow))
		panic(err)
	}
}

//Old_LostHandle 系统恢复入口
//func Old_LostHandle() {
//glog.Infoln("开始执行 sysTool.Old_LostHandle() ")
//	glog.Infof("开始恢复文件 : %s", file.FileFullName)
//
//	//收集剩余分块
//	//TODO: 按需要收集分块
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
//TODO: 清除所有丢失节点？
//}

//Old_DetectDataFile 检测数据分块是否全部恢复
// func (file File) Old_DetectDataFile(node nodeHandler.Node, faultCount, targetRow int) bool {
// 	var NodeDataNum = node.GetIndexInDataNodes()
// 	var k = (int)((faultCount + 2) / 2 * (int)(math.Pow(-1, (float64)(faultCount+2))))
// 	var startNodeNum = (node.GetIndexInDataNodes() - targetRow*k + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
// 	var rddtNodeNum = (faultCount*len(nodeHandler.DataNodes) + startNodeNum) / sysTool.SysConfig.RowNum
// 	glog.Info("Detecting Data File : " + "temp/Data." + strconv.Itoa(NodeDataNum) + "/" + file.FileFullName + "." + strconv.Itoa(NodeDataNum) + strconv.Itoa(targetRow))
// 	if FileExistedInCenter("temp/Data." + strconv.Itoa(NodeDataNum) + "/" + file.FileFullName + "." + strconv.Itoa(NodeDataNum) + strconv.Itoa(targetRow)) {
// 		return true
// 	}
// 	if !file.DetectRddtFile(nodeHandler.AllNodes[nodeHandler.RddtNodes[rddtNodeNum]], k, startNodeNum) {
// 		return false
// 	}
// 	//此时可以确定校验分块存在，不需要在检测码链时再检测一次
// 	if !file.DetectKLine(NodeDataNum, targetRow, rddtNodeNum, k, false) {
// 		return false
// 	}
// 	file.RestoreDataFile(NodeDataNum, rddtNodeNum, k, targetRow)
// 	return true
// }

//OldDetectKLine 检测某条码链是否可用于恢复
//func (file File) OldDetectKLine(dataNodeNum, targetRow, rddtNodeNum, k int, isForRddt bool) bool {
//	var result = true
//	var startIndex = (dataNodeNum - k*targetRow + len(nodeHandler.DataNodes)) % len(nodeHandler.DataNodes)
//	var tempResult = true
//	for rowCount := 0; rowCount < sysTool.SysConfig.RowNum; rowCount++ {
//		if (targetRow == rowCount) && !isForRddt {
//			continue
//		}
//
//		tempResult = FileExistedInCenter("temp/Data." + strconv.Itoa(startIndex+rowCount*k+len(nodeHandler.DataNodes)%len(nodeHandler.DataNodes)) + "/" + file.FileFullName + "." + strconv.Itoa(startIndex+rowCount*k+len(nodeHandler.DataNodes)%len(nodeHandler.DataNodes)) + strconv.Itoa(rowCount))
//		result = result && tempResult
//	}
//	if isForRddt {
//		return result
//	}
//	tempResult = FileExistedInCenter("temp/Rddt." + strconv.Itoa(rddtNodeNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startIndex))
//	result = result && tempResult
//	return result
//}
