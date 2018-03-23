// +build !windows
package util

import "syscall"

type DiskUsage struct {
	stat *syscall.Statfs_t
}

//NewDiskUsage 返回包含磁盘使用情况的对象
func NewDiskUsage(volumePath string) *DiskUsage {
	var stat syscall.Statfs_t
	syscall.Statfs(volumePath, &stat)
	return &DiskUsage{&stat}
}

//Available 磁盘对用户的可用空间
func (this *DiskUsage) Available() uint64 {
	return this.stat.Bavail * uint64(this.stat.Bsize)
}

