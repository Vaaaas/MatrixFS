package fileHandler

import (
	"io"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/golang/glog"
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
				glog.Infof("Collect Files Data Node Status : %t, ID : %d", nodeHandler.AllNodes[nodeHandler.DataNodes[i]].Status, nodeHandler.DataNodes[i])
				if nodeHandler.AllNodes[nodeHandler.DataNodes[i]].Status == true {
					getOneFile(file, true, nodeHandler.DataNodes[i], i, j, 0)
				}
			}
		}
	}

	if file.Size <= 1000 {
		for i := 0; i < sysTool.SysConfig.RddtNum; i++ {
			//只需要第一个
			getOneFile(file, false, nodeHandler.RddtNodes[i], i, 0, i)
		}
	} else {
		nodeCounter := 0
		fileCounter := 0
		rddtFileCounter := 0
		for fCounter := 0; fCounter < sysTool.SysConfig.FaultNum; fCounter++ {
			k := (int)((fCounter + 2) / 2 * (int)(math.Pow(-1, (float64)(fCounter+2))))
			for fileCounter < sysTool.SysConfig.DataNum {
				glog.Infof("从节点获取冗余文件 Rddt Node Num : %d \t k : %d \t fileCounter : %d \t nodeCounter : %d\n", nodeCounter, k, fileCounter, nodeCounter)
				if nodeHandler.AllNodes[nodeHandler.RddtNodes[nodeCounter]].Status == true {
					getOneFile(file, false, nodeHandler.RddtNodes[nodeCounter], k, fileCounter, nodeCounter)
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

func getOneFile(file File, isData bool, nodeID uint, posiX, posiY, nodeCounter int) {
	var filePath string
	var fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	if isData {
		filePath = StructSliceFileName("temp", true, posiX, file.FileFullName, posiX, posiY)
	} else {
		filePath = StructSliceFileName("temp", false, nodeCounter, file.FileFullName, posiX, posiY)
	}

	glog.Infof("从节点收集文件 filePath : %s, fileName : %s", filePath, fileName)

	url := "http://" + nodeHandler.AllNodes[nodeID].Address.String() + ":" + strconv.Itoa(nodeHandler.AllNodes[nodeID].Port) + "/download/" + fileName
	glog.Infof("获取文件的URL : %s", url)
	res, _ := http.Get(url)
	fileGet, _ := os.Create(filePath)
	defer fileGet.Close()
	io.Copy(fileGet, res.Body)
	glog.Info(res.Status)
}
