package NodeStruct

import (
	"net"
	"syscall"
	"unsafe"

	"github.com/Vaaaas/MatrixFS/File"
	"github.com/golang/glog"
	"github.com/Vaaaas/MatrixFS/SysConfig"
)

var IDCounter uint

type Node struct {
	ID       uint    `json:"ID"`
	Address  net.IP  `json:Address`
	Port     int     `json:Port`
	Volume   float64 `json:Volume`
	Status   bool    `json:Status`
	Lasttime int64   `json:Lasttime`
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

// disk usage of path/disk
func DiskUsage(path string) (free float64) {
	//macOS:
	//fs := syscall.Statfs_t{}
	//err := syscall.Statfs(path, &fs)
	//if err != nil {
	//	return
	//}
	//free = float64(fs.Bfree) * float64(fs.Bsize)
	//return free / float64(GB)

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

func (node Node) seekInNodes(Nodes []uint) {
	for _, nodeID := range Nodes {
		if node.ID == nodeID {
			return true
		}
	}
}

func (node Node) DetectNode(file File.File, dataNodes []uint, rddtNodes []uint) {
	if node.seekInNodes(dataNodes) {
		var allExist = false
		for faultCount := 0; faultCount < SysConfig.SysConfig.FaultNum; faultCount++ {
			if (allExist){
				break;
			}
			allExist = true;
			for rowCount := 0; rowCount < SysConfig.SysConfig.RowNum; rowCount++ {
				//var result = file.DetectDataFile(this, faultCount, rowCount);
				//allExist = allExist && result;
			}
		}
		return allExist;
	} else if node.seekInNodes(rddtNodes) {
	}
}

func GetIndexInData(dataNodes []uint, targetID uint) int {
	for index, dataNodeID := range dataNodes {
		if dataNodeID == targetID {
			return index
		}
	}
	return 0
}

func GetIndexInRddt(rddtNodes []uint, targetID uint) int {
	for index, rddtNodeID := range rddtNodes {
		if rddtNodeID == targetID {
			return index
		}
	}
	return 0
}