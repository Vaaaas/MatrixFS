// +build !windows

//Package util 包含一些通用的工具函数
package main

import "syscall"

//DiskUsage 磁盘使用情况结构体
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
func (du *DiskUsage) Available() uint64 {
	return du.stat.Bavail * uint64(du.stat.Bsize)
}
