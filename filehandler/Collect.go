package filehandler

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/Vaaaas/MatrixFS/nodehandler"
	"github.com/Vaaaas/MatrixFS/util"
)

// CollectFiles 从存储节点收集全部分块文件到中心节点
func (file File) CollectFiles() {
	if !util.SysConfig.Status && file.Size > 1000 {
		var dataNodes []uint
		var rddtNodes []uint
		//找出所有数据节点
		for col := 0; col < len(nodehandler.LostNodes); col++ {
			if nodehandler.IsDataNode(nodehandler.LostNodes[col]) {
				dataNodes = append(dataNodes, nodehandler.LostNodes[col])
			} else {
				rddtNodes = append(rddtNodes, nodehandler.LostNodes[col])
			}
		}

		if len(dataNodes) != 0 {
			//有数据节点丢失
			var recFinish = true
			for row := 0; row < util.SysConfig.RowNum; row++ {
				var rowResult = true
				for col := 0; col < len(dataNodes); col++ {
					//检测并恢复单个文件
					node := nodehandler.AllNodes.Get(dataNodes[col]).(nodehandler.Node)
					colResult := file.detectDataFile(node, row)
					rowResult = rowResult && colResult
				}
				recFinish = recFinish && rowResult
			}
			for !recFinish {
				recFinish = true
				for row := 0; row < util.SysConfig.RowNum; row++ {
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
	for i := 0; i < util.SysConfig.DataNum; i++ {
		if file.Size <= 1000 {
			getOneFile(file, true, nodehandler.DataNodes[i], i, 0, 0)
		} else {
			for j := 0; j < util.SysConfig.RowNum; j++ {
				node := nodehandler.AllNodes.Get(nodehandler.DataNodes[i]).(nodehandler.Node)
				filePath := structSliceFileName("./temp", true, i, file.FileFullName, i, j)
				if node.Status == true && !fileExistedInCenter(filePath) {
					getOneFile(file, true, nodehandler.DataNodes[i], i, j, 0)
				}
			}
		}
	}
}

func getOneFile(file File, isData bool, nodeID uint, posiX, posiY, rddtNodePos int) bool {
	var filePath string
	var fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	if isData {
		filePath = structSliceFileName("./temp", true, posiX, file.FileFullName, posiX, posiY)
	} else {
		filePath = structSliceFileName("./temp", false, rddtNodePos, file.FileFullName, posiX, posiY)
	}
	//glog.Infof("从节点收集文件 filePath : %s, fileName : %s", filePath, fileName)
	node := nodehandler.AllNodes.Get(nodeID).(nodehandler.Node)
	url := "http://" + node.Address.String() + ":" + strconv.Itoa(node.Port) + "/download/" + fileName
	res, _ := http.Get(url)
	if res.StatusCode != 200 {
		return false
	}
	fileGet, _ := os.Create(filePath)
	defer fileGet.Close()
	io.Copy(fileGet, res.Body)
	return true
}
