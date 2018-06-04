package nodehandler

import (
	"net"
	"time"

	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/Vaaaas/MatrixFS/util"
	"github.com/golang/glog"
)

// IDCounter ID 线程安全计数器
var IDCounter *util.SafeID

// AllNodes key：节点ID Value：节点对象
var AllNodes *util.SafeMap

// DataNodes 数据节点ID列表
var DataNodes []uint

// RddtNodes 校验节点ID列表
var RddtNodes []uint

// LostNodes 丢失节点ID列表
var LostNodes []uint

// EmptyNodes 空节点ID列表
var EmptyNodes []uint

// Node 节点结构体
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

//NodeStatusDetect 节点状态监控
func NodeStatusDetect() {
	for {
		now := time.Now().UnixNano() / 1000000
		allNodesListTemp := AllNodes.Items()
		for key, value := range allNodesListTemp {
			converted, _ := value.(Node)
			key, _ := key.(uint)
			if now-converted.LastTime > 4000 {
				converted.Status = false
				AllNodes.Set(key, converted)
				onDeleted(&converted)
			}
		}
		time.Sleep(4 * time.Second)
	}
}

//onDeleted 节点丢失处理
func onDeleted(node *Node) {
	var isEmpty = false
	for _, value := range EmptyNodes {
		if value == node.ID {
			glog.Info("已在空节点列表中找到丢失的空节点")
			isEmpty = true
		}
	}

	if isEmpty {
		//If Empty Node Lost, delete from all & Empty Slices
		AllNodes.Delete(node.ID)
		index := util.GetIndexInAll(len(EmptyNodes), func(i int) bool {
			return EmptyNodes[i] == node.ID
		})
		EmptyNodes = append(EmptyNodes[:index], EmptyNodes[index+1:]...)
		glog.Warning("空节点丢失，已删除空节点")
	} else {
		var lostExist = false
		for _, value := range LostNodes {
			if value == node.ID {
				lostExist = true
			}
		}
		if !lostExist {
			LostNodes = append(LostNodes, node.ID)
			util.SysConfig.Status = false
			glog.Warningf("新的丢失节点, SysConfigure 变为 false, 丢失节点ID : %d", node.ID)
		}
	}
}

// AppendNode Node的成员 将空节点加入数据或校验节点列表（如果全部满，则仍为空节点）
func (node Node) AppendNode() bool {
	if util.SysConfig.DataNum-len(DataNodes) > 0 {
		DataNodes = append(DataNodes, node.ID)
		index := util.GetIndexInAll(len(EmptyNodes), func(i int) bool {
			return EmptyNodes[i] == node.ID
		})
		EmptyNodes = append(EmptyNodes[:index], EmptyNodes[index+1:]...)
		return true
	} else if util.SysConfig.RddtNum-len(RddtNodes) > 0 {
		RddtNodes = append(RddtNodes, node.ID)
		index := util.GetIndexInAll(len(EmptyNodes), func(i int) bool {
			return EmptyNodes[i] == node.ID
		})
		EmptyNodes = append(EmptyNodes[:index], EmptyNodes[index+1:]...)
		return true
	} else {
		return false
	}
}

// GetIndexInDataNodes Node的成员 获取节点在所有数据节点中的index（位置）
func (node Node) GetIndexInDataNodes() int {
	for index, nodeID := range DataNodes {
		if node.ID == nodeID {
			return index
		}
	}
	return 0
}

// IsDataNode Node的成员 判断该节点是否为数据节点
func IsDataNode(ID uint) bool {
	for _, nodeID := range DataNodes {
		if ID == nodeID {
			return true
		}
	}
	return false
}

// NodeConfigured 判断节点是否已经配置
func NodeConfigured() bool {
	return AllNodes.Count() != 0
}

// EmptyNodeToLostNode 根据空节点ID和丢失节点ID执行空节点转化为丢失节点
func EmptyNodeToLostNode(empID, lostID uint) {
	// node : 空节点对象
	node := AllNodes.Get(empID).(Node)
	// 生成url
	url := "http://" + node.Address.String() + ":" + strconv.Itoa(node.Port) + "/resetid"
	// 向空节点发送重设ID请求
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	req.Header.Set("NewID", strconv.Itoa((int)(lostID)))
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
	// 转化完成，得到新节点信息
	node.ID = lostID
	node.Status = false
	AllNodes.Set(lostID, node)
	AllNodes.Delete(empID)
}
