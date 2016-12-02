package NodeStruct

import (
	"syscall"
	"net"
)

var AllNodes []Node
var LostNodes []int
var IDCounter uint

type Node struct {
	ID      uint `json:"ID"`
	Address net.IP `json:Address`
	port    int `json:Port`
	Volume  float64 `json:Volume`
	Status  bool `json:Status`

	//Status:
	//false	 -> 丢失或
	//true	 -> 正常
	//
}

type DiskStatus struct {
	All  float64 `json:"all"`
	Used float64 `json:"used"`
	Free float64 `json:"free"`
}

const (
	B = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

// disk usage of path/disk
func diskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = float64(fs.Blocks) * float64(fs.Bsize)
	disk.Free = float64(fs.Bfree) * float64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}

func DiskFreeSize(path string) (float64) {
	disk := diskUsage(path)
	return disk.Free / float64(GB)
}