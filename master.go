package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/Vaaaas/MatrixFS/filehandler"
	"github.com/Vaaaas/MatrixFS/nodehandler"
	"github.com/Vaaaas/MatrixFS/server"
	"github.com/Vaaaas/MatrixFS/util"
	"github.com/golang/glog"
)

func main() {
	//log_dir="./log"
	flag.Parse()
	//退出时调用 将日志写入文件
	defer glog.Flush()
	glog.Info("Master 服务器启动")
	//建立临时文件存储文件夹
	err := os.MkdirAll("./temp", 0766)
	if err != nil {
		glog.Errorln(err)
	}
	//建立日志存储文件夹
	err = os.MkdirAll("./log", 0766)
	if err != nil {
		glog.Errorln(err)
	}

	nodehandler.IDCounter = util.NewSafeID()
	//初始化节点Map
	nodehandler.AllNodes = util.NewSafeMap()
	//初始化文件Map
	filehandler.AllFiles = util.NewSafeMap()

	//页面处理方法
	http.HandleFunc("/", server.RootHandler)
	http.HandleFunc("/index", server.IndexPageHandler)
	http.HandleFunc("/file", server.FilePageHandler)
	http.HandleFunc("/node", server.NodeEnterHandler)

	//功能处理方法
	http.HandleFunc("/greet", server.GreetHandler)
	http.HandleFunc("/upload", server.UploadHandler)
	http.HandleFunc("/delete", server.DeleteHandler)
	http.HandleFunc("/download", server.DownloadHandler)
	http.HandleFunc("/restore", server.RestoreHandler)

	//定时遍历所有节点，比较最后访问时间
	go func() {
		for {
			now := time.Now().UnixNano() / 1000000
			allNodesListTemp := nodehandler.AllNodes.Items()
			for key, value := range allNodesListTemp {
				converted, _ := value.(nodehandler.Node)
				key, _ := key.(uint)
				if now-converted.LastTime > 6000 {
					converted.Status = false
					nodehandler.AllNodes.Set(key, converted)
					onDeleted(&converted)
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		glog.Errorln(err)
	}
}

//onDeleted 节点丢失处理
func onDeleted(node *nodehandler.Node) {
	var isEmpty = false
	for _, value := range nodehandler.EmptyNodes {
		if value == node.ID {
			glog.Info("已在空节点列表中找到丢失的空节点")
			isEmpty = true
		}
	}

	if isEmpty {
		//If Empty Node Lost, delete from all & Empty Slices
		nodehandler.AllNodes.Delete(node.ID)
		index := util.GetIndexInAll(len(nodehandler.EmptyNodes), func(i int) bool {
			return nodehandler.EmptyNodes[i] == node.ID
		})
		nodehandler.EmptyNodes = append(nodehandler.EmptyNodes[:index], nodehandler.EmptyNodes[index+1:]...)
		glog.Warning("空节点丢失，已删除空节点")
	} else {
		var lostExist = false
		for _, value := range nodehandler.LostNodes {
			if value == node.ID {
				lostExist = true
			}
		}
		if !lostExist {
			nodehandler.LostNodes = append(nodehandler.LostNodes, node.ID)
			util.SysConfig.Status = false
			glog.Warningf("新的丢失节点, SysConfigure 变为 false, 丢失节点ID : %d", node.ID)
		}
	}
}
