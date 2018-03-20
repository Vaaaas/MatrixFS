package filehandler

import (
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/golang/glog"
	"github.com/Vaaaas/MatrixFS/sysTool"
	"github.com/Vaaaas/MatrixFS/nodeHandler"
)

//DeleteSlices 处理用户删除文件的请求
func (file File) DeleteSlices() {
	for i := 0; i < sysTool.SysConfig.DataNum; i++ {
		if file.Size <= 1000 {
			//只需删除第一个
			deleteOneFile(file, true, nodeHandler.DataNodes[i], i, 0)
		} else {
			for j := 0; j < sysTool.SysConfig.RowNum; j++ {
				deleteOneFile(file, true, nodeHandler.DataNodes[i], i, j)
			}
		}
	}

	if file.Size <= 1000 {
		for i := 0; i < sysTool.SysConfig.RddtNum; i++ {
			//只需删除第一个
			deleteOneFile(file, false, nodeHandler.RddtNodes[i], i, 0)
		}
	} else {
		nodeCounter := 0
		fileCounter := 0
		rddtFileCounter := 0
		for fCounter := 0; fCounter < sysTool.SysConfig.FaultNum; fCounter++ {
			k := (int)((fCounter + 2) / 2 * (int)(math.Pow(-1, (float64)(fCounter+2))))
			for fileCounter < sysTool.SysConfig.DataNum {
				glog.Infof("Rddt Node Num : %d \t k : %d \t fileCounter : %d \t nodeCounter : %d\n", nodeCounter, k, fileCounter, nodeCounter)
				//执行删除
				deleteOneFile(file, false, nodeHandler.RddtNodes[nodeCounter], k, fileCounter)
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

//执行删除存储节点中某个文件的操作
func deleteOneFile(file File, isData bool, nodeID uint, posiX, posiY int) {
	var fileName string
	if isData {
		fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	} else {
		fileName = file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	}
	//设定请求发送url
	url := "http://" + nodeHandler.AllNodes[nodeID].Address.String() + ":" + strconv.Itoa(nodeHandler.AllNodes[nodeID].Port) + "/delete"
	glog.Info("[DELETE] URL " + url)

	//设定发送至存储节点的删除请求
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	//设定请求header
	req.Header.Set("fileName", fileName)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	glog.Info(resp.Status)
	glog.Info(string(respbody))
}

//DeleteAllTempFiles 删除中心节点中某个文件的全部临时文件
func (file File) DeleteAllTempFiles() error {
	file.deleteTempDataFiles()
	file.deleteTempRddtFiles()
	//删除原始文件副本
	if !FileExistedInCenter("temp/" + file.FileFullName) {
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

func (file File) deleteTempDataFiles() error {
	for i := 0; i < sysTool.SysConfig.DataNum; i++ {
		if file.Size <= 1000 {
			//删除副本
			filePath := StructSliceFileName("temp", true, i, file.FileFullName, i, 0)
			if !FileExistedInCenter(filePath) {
				glog.Warningf("[File to Delete NOT EXIST] temp/Data.%d/%s.%d%d ", i, file.FileFullName, i, 0)
			} else {
				err := os.Remove(filePath)
				if err != nil {
					glog.Errorln(err)
				} else {
					glog.Infof("[File to Delete] temp/Data.%d/%s.%d%d ", i, file.FileFullName, i, 0)
				}
			}
		} else {
			for j := 0; j < sysTool.SysConfig.RowNum; j++ {
				filePath := StructSliceFileName("temp", true, i, file.FileFullName, i, j)
				if !FileExistedInCenter(filePath) {
					glog.Warningf("[File to Delete NOT EXIST] temp/Data.%d/%s.%d%d ", i, file.FileFullName, i, j)
				} else {
					err := os.Remove(filePath)
					if err != nil {
						glog.Errorln(err)
					} else {
						glog.Infof("[File to Delete] temp/Data.%d/%s.%d%d ", i, file.FileFullName, i, j)
					}
				}
			}
		}
	}
	return nil
}

func (file File) deleteTempRddtFiles() error {
	if file.Size <= 1000 {
		for i := 0; i < sysTool.SysConfig.RddtNum; i++ {
			filePath := StructSliceFileName("temp", false, i, file.FileFullName, i, 0)
			if !FileExistedInCenter(filePath) {
				glog.Warningf("[File to Delete NOT EXIST] temp/Rddt.%d/%s.%d%d ", i, file.FileFullName, i, 0)
			} else {
				err := os.Remove(filePath)
				if err != nil {
					glog.Errorln(err)
				} else {
					glog.Infof("[File to Delete] temp/RDTT.%d/%s.%d%d ", i, file.FileFullName, i, 0)
				}
			}
		}
	} else {
		nodeCounter := 0
		fileCounter := 0
		rddtFileCounter := 0
		for i := 0; i < sysTool.SysConfig.FaultNum; i++ {
			k := (int)((i + 2) / 2 * (int)(math.Pow(-1, (float64)(i+2))))
			for fileCounter < sysTool.SysConfig.DataNum {
				filePath := StructSliceFileName("temp", false, i, file.FileFullName, k, fileCounter)
				if !FileExistedInCenter(filePath) {
					glog.Warningf("[File to Delete NOT EXIST] temp/Rddt.%d/%s.%d%d ", nodeCounter, file.FileFullName, k, fileCounter)
				} else {
					err := os.Remove(filePath)
					if err != nil {
						glog.Errorln(err)
					} else {
						glog.Infof("[File to Delete] temp/RDTT.%d/%s.%d%d ", nodeCounter, file.FileFullName, k, fileCounter)
					}
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
	return nil
}
