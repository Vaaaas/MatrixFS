package filehandler

import (
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/Vaaaas/MatrixFS/nodehandler"
	"github.com/Vaaaas/MatrixFS/util"
	"github.com/golang/glog"
)

// DeleteSlices 处理用户删除文件的请求
func (file File) DeleteSlices() {
	for i := 0; i < util.SysConfig.DataNum; i++ {
		if file.Size <= 1000 {
			// 只需删除第一个
			deleteOneFile(file, true, nodehandler.DataNodes[i], i, 0)
		} else {
			for j := 0; j < util.SysConfig.RowNum; j++ {
				deleteOneFile(file, true, nodehandler.DataNodes[i], i, j)
			}
		}
	}

	if file.Size <= 1000 {
		for i := 0; i < util.SysConfig.RddtNum; i++ {
			// 只需删除第一个
			deleteOneFile(file, false, nodehandler.RddtNodes[i], i, 0)
		}
	} else {
		nodeCounter := 0
		fileCounter := 0
		rddtFileCounter := 0
		for fCounter := 0; fCounter < util.SysConfig.FaultNum; fCounter++ {
			k := (int)((fCounter + 2) / 2 * (int)(math.Pow(-1, (float64)(fCounter+2))))
			for fileCounter < util.SysConfig.DataNum {
				// 执行删除
				deleteOneFile(file, false, nodehandler.RddtNodes[nodeCounter], k, fileCounter)
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

//执行删除存储节点中某个文件的操作
func deleteOneFile(file File, isData bool, nodeID uint, posiX, posiY int) {
	var fileName string
	if isData {
		fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	} else {
		fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	}
	//设定请求发送url
	node := nodehandler.AllNodes.Get(nodeID).(nodehandler.Node)
	url := "http://" + node.Address.String() + ":" + strconv.Itoa(node.Port) + "/delete"
	// 设定发送至存储节点的删除请求
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	// 设定请求header
	req.Header.Set("fileName", fileName)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
}

// DeleteAllTempFiles 删除中心节点中某个文件的全部临时文件
func (file File) DeleteAllTempFiles() error {
	file.deleteTempDataFiles()
	file.deleteTempRddtFiles()
	// 删除原始文件副本
	if !fileExistedInCenter("temp/" + file.FileFullName) {
		glog.Warningf("[File to Delete NOT EXIST] temp/" + file.FileFullName)
	} else {
		err := os.Remove("temp/" + file.FileFullName)
		if err != nil {
			glog.Errorln(err)
		}
	}
	return nil
}

func (file File) deleteTempDataFiles() error {
	for i := 0; i < util.SysConfig.DataNum; i++ {
		if file.Size <= 1000 {
			// 删除副本
			filePath := structSliceFileName("./temp", true, i, file.FileFullName, i, 0)
			if !fileExistedInCenter(filePath) {
				glog.Warningf("[File to Delete NOT EXIST] temp/Data.%d/%s.%d%d ", i, file.FileFullName, i, 0)
			} else {
				err := os.Remove(filePath)
				if err != nil {
					glog.Errorln(err)
				}
			}
		} else {
			for j := 0; j < util.SysConfig.RowNum; j++ {
				filePath := structSliceFileName("./temp", true, i, file.FileFullName, i, j)
				if !fileExistedInCenter(filePath) {
					break
				} else {
					err := os.Remove(filePath)
					if err != nil {
						glog.Errorln(err)
					}
				}
			}
		}
	}
	return nil
}

func (file File) deleteTempRddtFiles() error {
	if file.Size <= 1000 {
		for i := 0; i < util.SysConfig.RddtNum; i++ {
			filePath := structSliceFileName("./temp", false, i, file.FileFullName, i, 0)
			if !fileExistedInCenter(filePath) {
				glog.Warningf("[File to Delete NOT EXIST] temp/Rddt.%d/%s.%d%d ", i, file.FileFullName, i, 0)
			} else {
				err := os.Remove(filePath)
				if err != nil {
					glog.Errorln(err)
				}
			}
		}
	} else {
		nodeCounter := 0
		fileCounter := 0
		rddtFileCounter := 0

		for i := 0; i < util.SysConfig.FaultNum; i++ {
			k := (int)((i + 2) / 2 * (int)(math.Pow(-1, (float64)(i+2))))
			for fileCounter < util.SysConfig.DataNum {
				filePath := structSliceFileName("./temp", false, nodeCounter, file.FileFullName, k, fileCounter)
				if !fileExistedInCenter(filePath) {
					//glog.Warningf("[File to Delete NOT EXIST] temp/RDDT.%d/%s.%d%d ", nodeCounter, file.FileFullName, k, fileCounter)
				} else {
					err := os.Remove(filePath)
					if err != nil {
						glog.Errorln(err)
					}
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
	return nil
}
