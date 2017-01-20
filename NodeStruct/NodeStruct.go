package NodeStruct

import (
	"syscall"
	"net"
	"unsafe"
	"github.com/golang/glog"
)

var IDCounter uint

type Node struct {
	ID       uint `json:"ID"`
	Address  net.IP `json:Address`
	Port     int `json:Port`
	Volume   float64 `json:Volume`
	Status   bool `json:Status`
	Lasttime int64 `json:Lasttime`

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