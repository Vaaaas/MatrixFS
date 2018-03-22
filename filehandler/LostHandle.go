package filehandler

import (
	"io"
	"math"
	"os"
	"strconv"

	"github.com/Vaaaas/MatrixFS/nodehandler"
	"github.com/Vaaaas/MatrixFS/sysTool"
	"github.com/golang/glog"
)

//LostHandle 执行解码恢复功能
func (file *File) LostHandle() bool {
	glog.Infoln("开始执行 LostHandle() ")
	glog.Infof("开始恢复文件 : %s", file.FileFullName)

	if file.size <= 1000 {
		var dataNodePos int
		for i := 0; i < len(nodehandler.DataNodes); i++ {
			if nodehandler.AllNodes.Get(nodehandler.DataNodes[i]).(nodehandler.Node).Status {
				dataNodePos = i
				getOneFile(*file, true, nodehandler.DataNodes[i], i, 0, 0)
			}
		}
		source := structSliceFileName("./temp", true, dataNodePos, file.FileFullName, dataNodePos, 0)
		//打开原始文件
		sourceFile, err := os.Open(source)
		if err != nil {
			glog.Error("分割时打开源文件失败")
			panic(err)
		}
		defer sourceFile.Close()
		for i := 0; i < len(nodehandler.LostNodes); i++ {
			if nodehandler.IsDataNode(nodehandler.LostNodes[i]) {
				index := sysTool.GetIndexInAll(len(nodehandler.DataNodes), func(finder int) bool {
					return nodehandler.DataNodes[finder] == nodehandler.LostNodes[i]
				})
				file.copyFile(true, index, sourceFile)
				postOneFile(*file, true, nodehandler.DataNodes[index], index, 0, 0)
			} else {
				index := sysTool.GetIndexInAll(len(nodehandler.RddtNodes), func(finder int) bool {
					return nodehandler.RddtNodes[finder] == nodehandler.LostNodes[i]
				})
				file.copyFile(false, index, sourceFile)
				postOneFile(*file, false, nodehandler.RddtNodes[index], index, 0, 0)
			}
		}
	} else {
		var dataNodes []uint
		var rddtNodes []uint
		//找出所有数据节点
		for col := 0; col < len(nodehandler.LostNodes); col++ {
			glog.Infof("需要检测节点 ID : %d", nodehandler.LostNodes[col])
			if nodehandler.IsDataNode(nodehandler.LostNodes[col]) {
				dataNodes = append(dataNodes, nodehandler.LostNodes[col])
			} else {
				rddtNodes = append(rddtNodes, nodehandler.LostNodes[col])
			}
		}
		if len(dataNodes) != 0 {
			//有数据节点丢失
			var recFinish = true
			for row := 0; row < sysTool.SysConfig.RowNum/2+1; row++ {
				var rowResult = true
				for col := 0; col < len(dataNodes); col++ {
					//检测并恢复单个文件
					node := nodehandler.AllNodes.Get(dataNodes[col]).(nodehandler.Node)
					colResult := file.detectDataFile(node, row)
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
							node := nodehandler.AllNodes.Get(dataNodes[col]).(nodehandler.Node)
							colResult := file.detectDataFile(node, row)
							rowResult = rowResult && colResult
						}
						recFinish = recFinish && rowResult
					}
				}
			}
		}
		//生成丢失的校验分块
		for i := 0; i < len(rddtNodes); i++ {
			index := sysTool.GetIndexInAll(len(nodehandler.RddtNodes), func(finder int) bool {
				return nodehandler.RddtNodes[finder] == nodehandler.LostNodes[i]
			})
			for row := 0; row < sysTool.SysConfig.RowNum; row++ {
				posY := (index*sysTool.SysConfig.RowNum + row) % sysTool.SysConfig.DataNum
				faultCount := (index*sysTool.SysConfig.RowNum + row) / sysTool.SysConfig.DataNum
				k := (int)((faultCount + 2) / 2 * (int)(math.Pow(-1, (float64)(faultCount+2))))
				file.detectKLine(posY, row, index, k)
				file.initOneRddtFile(posY, k, index)
				postOneFile(*file, false, nodehandler.RddtNodes[index], k, posY, index)
			}
		}
	}
	return true
}

//detectDataFile 检测单个数据文件，如果不在中心节点，则依次检测对应校验文件，码链上的所有数据分块，以判断是否可恢复，可以则恢复
func (file File) detectDataFile(node nodehandler.Node, targetRow int) bool {
	var dataNodePos = node.GetIndexInDataNodes()
	filePath := structSliceFileName("./temp", true, dataNodePos, file.FileFullName, dataNodePos, targetRow)
	if fileExistedInCenter(filePath) {
		return true
	}
	for fCount := 0; fCount < sysTool.SysConfig.FaultNum; fCount++ {
		k := (int)((fCount + 2) / 2 * (int)(math.Pow(-1, (float64)(fCount+2))))
		var startDataNodePos = (node.GetIndexInDataNodes() - targetRow*k + len(nodehandler.DataNodes)) % len(nodehandler.DataNodes)
		var rddtNodePos = (fCount*len(nodehandler.DataNodes) + startDataNodePos) / sysTool.SysConfig.RowNum
		glog.Info("Detecting Data File : " + "temp/Data." + strconv.Itoa(dataNodePos) + "/" + file.FileFullName + "." + strconv.Itoa(dataNodePos) + strconv.Itoa(targetRow))
		//首先判断本地是否有该文件
		if !file.detectRddtFile(rddtNodePos, k, startDataNodePos) {
			//当前k对应的校验文件不在中心节点
			//尝试从存储节点下载该校验分块
			if !getOneFile(file, false, nodehandler.RddtNodes[rddtNodePos], k, startDataNodePos, rddtNodePos) {
				//未取得校验文件
				if fCount == sysTool.SysConfig.FaultNum-1 {
					//直到最后一种斜率也不行
					return false
				}
				continue
			}
		}
		if !file.detectKLine(dataNodePos, targetRow, rddtNodePos, k) {
			if fCount == sysTool.SysConfig.FaultNum-1 {
				//直到最后一种斜率也不行
				return false
			}
			continue
		}
		file.restoreDataFile(dataNodePos, rddtNodePos, k, targetRow)
		postOneFile(file, true, nodehandler.DataNodes[dataNodePos], dataNodePos, targetRow, 0)
		//收集对称数据分块对应的校验文件和码链
		pairTargetRow := sysTool.SysConfig.RowNum - targetRow - 1
		if pairTargetRow != targetRow {
			var pairRddtNodePos int
			var pairStartDataNodePos = (node.GetIndexInDataNodes() - pairTargetRow*(-k) + len(nodehandler.DataNodes)) % len(nodehandler.DataNodes)
			if k > 0 {
				pairRddtNodePos = ((fCount+1)*len(nodehandler.DataNodes) + pairStartDataNodePos) / sysTool.SysConfig.RowNum
			} else {
				pairRddtNodePos = ((fCount-1)*len(nodehandler.DataNodes) + pairStartDataNodePos) / sysTool.SysConfig.RowNum
			}
			if getOneFile(file, false, nodehandler.RddtNodes[pairRddtNodePos], -k, pairStartDataNodePos, pairRddtNodePos) {
				if file.detectKLine(pairStartDataNodePos, pairTargetRow, pairRddtNodePos, -k) {
					file.restoreDataFile(pairStartDataNodePos, pairRddtNodePos, -k, pairTargetRow)
					postOneFile(file, true, nodehandler.DataNodes[pairStartDataNodePos], pairStartDataNodePos, pairTargetRow, 0)
				}
			}
		}
		return true
	}
	return false
}

//detectRddtFile 检测所有校验分块是否全部恢复
func (file File) detectRddtFile(rddtNodePos, k, dataNodePos int) bool {
	filePath := structSliceFileName("./temp", false, rddtNodePos, file.FileFullName, k, dataNodePos)
	if !fileExistedInCenter(filePath) {
		if !fileExistInNode(file, false, nodehandler.RddtNodes[rddtNodePos], k, dataNodePos, rddtNodePos) {
			return false
		}
	}
	return true
}

//detectKLine 检测一条码链上的所有数据文件
func (file File) detectKLine(dataNodePos, targetRow, rddtNodePos, k int) bool {
	var result = true
	var startIndex = (dataNodePos - k*targetRow + len(nodehandler.DataNodes)) % len(nodehandler.DataNodes)
	for rowCount := 0; rowCount < sysTool.SysConfig.RowNum; rowCount++ {
		if targetRow == rowCount {
			continue
		}
		filePath := structSliceFileName("./temp", true, (startIndex+rowCount*k+len(nodehandler.DataNodes))%len(nodehandler.DataNodes), file.FileFullName, startIndex+rowCount*k+len(nodehandler.DataNodes)%len(nodehandler.DataNodes), rowCount)
		//先看中心节点有没有
		resultCenter := fileExistedInCenter(filePath)
		if !resultCenter {
			//中心节点没有
			resultNode := fileExistInNode(file, true, nodehandler.DataNodes[(startIndex+rowCount*k+len(nodehandler.DataNodes))%len(nodehandler.DataNodes)], dataNodePos, targetRow, rddtNodePos)
			if !resultNode {
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

//restoreDataFile 解码恢复单个数据分块
func (file File) restoreDataFile(dataNodePos, rddtNodePos, k, targetRow int) {
	var startIndex = (dataNodePos - k*targetRow + len(nodehandler.DataNodes)) % len(nodehandler.DataNodes)
	buffer := make([]byte, file.sliceSize)
	filePath := structSliceFileName("./temp", false, rddtNodePos, file.FileFullName, k, startIndex)
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
		dataPos := (startIndex + rowCount*k + len(nodehandler.DataNodes)) % len(nodehandler.DataNodes)
		filePath := structSliceFileName("./temp", true, dataPos, file.FileFullName, dataPos, rowCount)
		tempBytes := make([]byte, file.sliceSize)
		dataFile, err := os.Open(filePath)
		_, err = dataFile.Read(tempBytes)
		if err != nil && err != io.EOF {
			glog.Error("tempBuffer 读取 Data 文件失败" + filePath)
			panic(err)
		}
		defer dataFile.Close()
		glog.Infof("恢复中 len(buffer) : %d, len(tempBuffer) : %d, slicesize : %d", len(buffer), len(tempBytes), file.sliceSize)
		for byteCounter := 0; byteCounter < len(buffer); byteCounter++ {
			buffer[byteCounter] = buffer[byteCounter] ^ tempBytes[byteCounter]
		}
	}
	filePath = structSliceFileName("./temp", true, dataNodePos, file.FileFullName, dataNodePos, targetRow)
	targetFilePath, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		glog.Error("新建恢复数据冗余分块文件失败 " + filePath)
		panic(err)
	}
	defer targetFilePath.Close()
	if _, err := targetFilePath.Write(buffer[:file.sliceSize]); err != nil {
		glog.Error("写入冗余分块文件失败 " + "./temp/Data." + strconv.Itoa(dataNodePos) + "/" + file.FileFullName + "." + strconv.Itoa(dataNodePos) + strconv.Itoa(targetRow))
		panic(err)
	}
}
