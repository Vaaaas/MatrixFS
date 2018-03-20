package nodeHandler

import (
	"net"

	"github.com/Vaaaas/MatrixFS/sysTool"
	"github.com/golang/glog"
)

//IDCounter ID 计数器
var IDCounter uint

//AllNodes key：节点ID Value：节点对象

var AllNodes = make(map[uint]Node)

//DataNodes 数据节点ID列表
var DataNodes []uint

//RddtNodes 校验节点ID列表
var RddtNodes []uint

//LostNodes 丢失节点ID列表
var LostNodes []uint

//EmptyNodes 空节点ID列表
var EmptyNodes []uint

//Node 节点结构体
type Node struct {
	ID       uint    `json:"ID"`
	Address  net.IP  `json:"Address"`
	Port     int     `json:"Port"`
	Volume   float64 `json:"Volume"`
	Status   bool    `json:"Status"`
	LastTime int64   `json:"Lasttime"`
}

/*
B 用于表示1个单位
*/
const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

//AppendNode Node的成员 将空节点加入数据或校验节点列表（如果全部满，则仍为空节点）
func (node Node) AppendNode() bool {
	if CheckDataNodeNum() > 0 {
		EmptyToData(node.ID)
		index := sysTool.GetIndexInAll(len(EmptyNodes), func(i int) bool {
			return EmptyNodes[i] == node.ID
		})
		EmptyNodes = append(EmptyNodes[:index], EmptyNodes[index+1:]...)
		return true
	} else if CheckRddtNodeNum() > 0 {
		EmptyToRddt(node.ID)
		index := sysTool.GetIndexInAll(len(EmptyNodes), func(i int) bool {
			return EmptyNodes[i] == node.ID
		})
		EmptyNodes = append(EmptyNodes[:index], EmptyNodes[index+1:]...)
		return true
	} else {
		return false
	}
}

//GetIndexInDataNodes Node的成员 获取节点在所有数据节点中的index（位置）
func (node Node) GetIndexInDataNodes() int {
	for index, nodeID := range DataNodes {
		if node.ID == nodeID {
			glog.Infof("Index in All Data Nodes is : %d", index)
			return index
		}
	}
	return 0
}

//GetIndexInRddtNodes Node的成员 获取节点在所有校验节点中的index（位置）
func (node Node) GetIndexInRddtNodes() int {
	for index, nodeID := range RddtNodes {
		if node.ID == nodeID {
			glog.Infof("Index in All Rddt Nodes is : %d", index)
			return index
		}
	}
	return 0
}

//IsDataNode Node的成员 判断该节点是否为数据节点
func (node Node) IsDataNode() bool {
	for _, nodeID := range DataNodes {
		if node.ID == nodeID {
			return true
		}
	}
	return false
}

//IsRddtNode Node的成员 判断该节点是否为校验节点
func (node Node) IsRddtNode() bool {
	for _, nodeID := range RddtNodes {
		if node.ID == nodeID {
			return true
		}
	}
	return false
}

//NodeConfigured 判断节点是否已经配置
func NodeConfigured() bool {
	//TODO: 自动设定节点
	return len(AllNodes) != 0
}

//AddToLost 将节点ID加入失效节点列表
func AddToLost(nodeID uint) {
	LostNodes = append(LostNodes, nodeID)
}

//CheckDataNodeNum 确定系统还需要几个数据节点
func CheckDataNodeNum() int {
	return sysTool.SysConfig.DataNum - len(DataNodes)
}

//CheckRddtNodeNum 确定系统还需要几个校验节点
func CheckRddtNodeNum() int {
	return sysTool.SysConfig.RddtNum - len(RddtNodes)
}

//EmptyToData 将空节点ID加入数据节点ID列表
func EmptyToData(nodeID uint) {
	DataNodes = append(DataNodes, nodeID)
}

//EmptyToRddt 将空节点ID加入校验节点ID列表
func EmptyToRddt(nodeID uint) {
	RddtNodes = append(RddtNodes, nodeID)
}

//GetIndexInRddt 在全部校验节点中寻找某个节点ID在该列表中的位置
func GetIndexInRddt(targetID uint) int {
	for index, rddtNodeID := range RddtNodes {
		if rddtNodeID == targetID {
			return index
		}
	}
	return 0
}

//TODO: 去掉对fileHandler的依赖
// func (node Node) DetectNode(file fileHandler.File) (finished, isDataNode bool) {
// 	if node.IsDataNode() {
// 		glog.Infof("开始检测节点ID : %d, 文件名 : %s", node.ID, file.FileFullName)
// 		//var allExist = false
// 		for row := 0; row <= sysTool.SysConfig.RowNum/2+1; row++ {
// 			//TODO:DetectNode
// 			//file.Old_DetectDataFile()
// 		}

// 		return true, true
// 	} else {
// 		return true, false
// 	}
// }

//TODO: 去掉对fileHandler的依赖
//Old_DetectNode 检测节点中的文件是否全部恢复
// func (node Node) Old_DetectNode(file fileHandler.File) bool {
// 	glog.Infof("开始检测节点ID : %d, 文件名 : %s", node.ID, file.FileFullName)
// 	if node.IsDataNode() {
// 		glog.Info("被检测节点为 [DATA] Node")
// 		var allExist = false
// 		//TODO: 外层循环错误数？
// 		for faultCount := 0; faultCount < sysTool.SysConfig.FaultNum; faultCount++ {
// 			if allExist {
// 				break
// 			}
// 			allExist = true
// 			for rowCount := 0; rowCount < sysTool.SysConfig.RowNum; rowCount++ {
// 				var result = file.Old_DetectDataFile(node, faultCount, rowCount)
// 				allExist = allExist && result
// 			}
// 		}
// 		glog.Infof("数据节点（ID : %d, File Name : %s）恢复完成", node.ID, file.FileFullName)
// 		return allExist
// 	} else if node.IsRddtNode() {
// 		glog.Info("被检测节点为 [RDDT] Node")
// 		var allExist = false
// 		var rddtNum = GetIndexInRddt(node.ID)
// 		var nodeCount = rddtNum * sysTool.SysConfig.RowNum % len(DataNodes)
// 		var fCount = rddtNum * sysTool.SysConfig.RowNum / len(DataNodes)
// 		k := (int)((fCount + 2) / 2 * (int)(math.Pow(-1, (float64)(fCount+2))))
// 		for rowCount := 0; rowCount < sysTool.SysConfig.RowNum; rowCount++ {
// 			allExist = true
// 			var result = file.DetectRddtFile(node, k, nodeCount)
// 			allExist = allExist && result
// 			if nodeCount == len(DataNodes)-1 {
// 				nodeCount = 0
// 				fCount++
// 				k = (int)((rowCount + 2) / 2 * (int)(math.Pow(-1, (float64)(rowCount+2))))
// 			} else {
// 				nodeCount++
// 			}
// 		}
// 		glog.Infof("冗余节点（ID : %d, File Name : %s）恢复完成", node.ID, file.FileFullName)
// 		return allExist
// 	}
// 	return false
// }
