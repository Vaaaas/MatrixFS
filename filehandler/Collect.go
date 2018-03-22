package filehandler

import (
	"io"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/Vaaaas/MatrixFS/nodeHandler"
	"github.com/Vaaaas/MatrixFS/sysTool"
)

//CollectFiles 从存储节点收集全部分块文件到中心节点
func (file File) CollectFiles() {
	for i := 0; i < sysTool.SysConfig.DataNum; i++ {
		if file.Size <= 1000 {
			getOneFile(file, true, nodeHandler.DataNodes[i], i, 0, 0)
		} else {
			for j := 0; j < sysTool.SysConfig.RowNum; j++ {
				node := nodeHandler.AllNodes.Get(nodeHandler.DataNodes[i]).(nodeHandler.Node)
				//glog.Infof("Collect Files Data Node Status : %t, ID : %d", node.Status, nodeHandler.DataNodes[i])
				if node.Status == true {
					getOneFile(file, true, nodeHandler.DataNodes[i], i, j, 0)
				}
			}
		}
	}
	//if needRddt{
	//	if file.Size <= 1000 {
	//		for i := 0; i < sysTool.SysConfig.RddtNum; i++ {
	//			//只需要第一个
	//			getOneFile(file, false, nodeHandler.RddtNodes[i], i, 0, i)
	//		}
	//	} else {
	//		nodeCounter := 0
	//		fileCounter := 0
	//		rddtFileCounter := 0
	//		for fCounter := 0; fCounter < sysTool.SysConfig.FaultNum; fCounter++ {
	//			k := (int)((fCounter + 2) / 2 * (int)(math.Pow(-1, (float64)(fCounter+2))))
	//			for fileCounter < sysTool.SysConfig.DataNum {
	//				//glog.Infof("从节点获取冗余文件 Rddt Node Num : %d \t k : %d \t fileCounter : %d \t nodeCounter : %d\n", nodeCounter, k, fileCounter, nodeCounter)
	//				node := nodeHandler.AllNodes.Get(nodeHandler.RddtNodes[nodeCounter]).(nodeHandler.Node)
	//				if node.Status == true {
	//					getOneFile(file, false, nodeHandler.RddtNodes[nodeCounter], k, fileCounter, nodeCounter)
	//				}
	//				fileCounter++
	//				rddtFileCounter++
	//				if rddtFileCounter%(sysTool.SysConfig.SliceNum/sysTool.SysConfig.DataNum) == 0 {
	//					nodeCounter++
	//					rddtFileCounter = 0
	//				}
	//				if fileCounter != sysTool.SysConfig.DataNum {
	//					continue
	//				}
	//				fileCounter = 0
	//				break
	//			}
	//		}
	//	}
	//}
}

func getOneFile(file File, isData bool, nodeID uint, posiX, posiY, rddtNodePos int) bool {
	var filePath string
	var fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	if isData {
		filePath = StructSliceFileName("./temp", true, posiX, file.FileFullName, posiX, posiY)
	} else {
		filePath = StructSliceFileName("./temp", false, rddtNodePos, file.FileFullName, posiX, posiY)
	}
	//glog.Infof("从节点收集文件 filePath : %s, fileName : %s", filePath, fileName)
	node := nodeHandler.AllNodes.Get(nodeID).(nodeHandler.Node)
	url := "http://" + node.Address.String() + ":" + strconv.Itoa(node.Port) + "/download/" + fileName
	//glog.Infof("获取文件的URL : %s", url)
	res, _ := http.Get(url)
	if res.StatusCode != 200 {
		return false
	}
	fileGet, _ := os.Create(filePath)
	defer fileGet.Close()
	io.Copy(fileGet, res.Body)
	//glog.Info(res.Status)
	return true
}
