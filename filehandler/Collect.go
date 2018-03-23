package filehandler

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/Vaaaas/MatrixFS/nodehandler"
	"github.com/Vaaaas/MatrixFS/util"
)

//CollectFiles 从存储节点收集全部分块文件到中心节点
func (file File) CollectFiles() {
	for i := 0; i < util.SysConfig.DataNum; i++ {
		if file.Size <= 1000 {
			getOneFile(file, true, nodehandler.DataNodes[i], i, 0, 0)
		} else {
			for j := 0; j < util.SysConfig.RowNum; j++ {
				node := nodehandler.AllNodes.Get(nodehandler.DataNodes[i]).(nodehandler.Node)
				if node.Status == true {
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
