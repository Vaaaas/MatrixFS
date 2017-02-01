package Tool

import (
	"net"
	"syscall"
	"unsafe"
	"github.com/golang/glog"
)

var IDCounter uint

var AllNodes = make(map[uint]Node)
var DataNodes []uint
var RddtNodes []uint
var LostNodes []uint
var EmptyNodes []uint

type Node struct {
	ID       uint    `json:"ID"`
	Address  net.IP  `json:Address`
	Port     int     `json:Port`
	Volume   float64 `json:Volume`
	Status   bool    `json:Status`
	LastTime int64   `json:Lasttime`
	//Status:
	//false	 -> 丢失或
	//true	 -> 正常
}

const (
	B = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func NodeConfigured() bool {
	//todo : auto configure node
	return len(AllNodes) != 0
}

// disk usage of path/disk
func DiskUsage(path string) float64 {
	//Win:
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")
	lpFreeBytesAvailable := int64(0)
	lpTotalNumberOfBytes := int64(0)
	lpTotalNumberOfFreeBytes := int64(0)
	_, _, err := c.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)))
	if err != nil {
		glog.Error(err)
		//panic(err)
	}
	return (float64)(lpFreeBytesAvailable / 1024 / 1024.0 / 1000)
}

func (node Node) AppendNode() bool {
	if CheckDataNodeNum() > 0 {
		EmptyToData(node.ID)
		index := GetFileIndexInAll(len(EmptyNodes), func(i int) bool {
			return EmptyNodes[i] == node.ID
		})
		EmptyNodes = append(EmptyNodes[:index], EmptyNodes[index + 1:]...)
		return true
	} else if CheckRddtNodeNum() > 0 {
		EmptyToRddt(node.ID)
		index := GetFileIndexInAll(len(EmptyNodes), func(i int) bool {
			return EmptyNodes[i] == node.ID
		})
		EmptyNodes = append(EmptyNodes[:index], EmptyNodes[index + 1:]...)
		return true
	} else {
		return false
	}
}

func AddToLost(nodeID uint) {
	LostNodes = append(LostNodes, nodeID)
}

func (node Node) getIndexInDataNodes() (int) {
	for index, nodeID := range DataNodes {
		if node.ID == nodeID {
			glog.Infof("Index in All Data Nodes is : %d", index)
			return index
		}
	}
	return 0
}

func (node Node) getIndexInRddtNodes() (int) {
	for index, nodeID := range RddtNodes {
		if node.ID == nodeID {
			glog.Infof("Index in All DRddt Nodes is : %d", index)
			return index
		}
	}
	return 0
}

func (node Node) DetectNode(file File) bool {
	if node.isDataNode() {
		glog.Info("Detecting a [DATA] Node")
		var allExist = false
		for faultCount := 0; faultCount < SysConfig.FaultNum; faultCount++ {
			if (allExist) {
				break;
			}
			allExist = true;
			for rowCount := 0; rowCount < SysConfig.RowNum; rowCount++ {
				var result = file.DetectDataFile(node, faultCount, rowCount);
				allExist = allExist && result;
			}
		}
		//return allExist;
	} else if node.isRddtNode() {
		glog.Info("Detecting a [RDDT] Node")

	}
	return false
}

func (node Node) isDataNode() bool{
	for _, nodeID := range DataNodes {
		if node.ID == nodeID {
			return true
		}
	}
	return false
}

func (node Node) isRddtNode() bool{
	for _, nodeID := range RddtNodes {
		if node.ID == nodeID {
			return true
		}
	}
	return false
}

func CheckDataNodeNum() int {
	return SysConfig.DataNum - len(DataNodes)
}

func CheckRddtNodeNum() int {
	return SysConfig.RddtNum - len(RddtNodes)
}

func EmptyToData(nodeID uint) {
	DataNodes = append(DataNodes, nodeID)
}

func EmptyToRddt(nodeID uint) {
	RddtNodes = append(RddtNodes, nodeID)
}

func GetIndexInData(targetID uint) int {
	for index, dataNodeID := range DataNodes {
		if dataNodeID == targetID {
			return index
		}
	}
	return 0
}

func GetIndexInRddt(targetID uint) int {
	for index, rddtNodeID := range RddtNodes {
		if rddtNodeID == targetID {
			return index
		}
	}
	return 0
}