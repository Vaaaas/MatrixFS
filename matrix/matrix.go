package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/Vaaaas/MatrixFS/nodeHandler"
	"github.com/Vaaaas/MatrixFS/sysTool"
	"github.com/golang/glog"
)

func main() {
	//when debug, log_dir="./log"
	flag.Parse()
	//Trigger on exit, write log into files
	defer glog.Flush()
	//Init system config(Moved to indexPageHandler)
	//SysConfig.InitConfig(fault, row)
	glog.Info("Master 服务器启动")

	err := os.MkdirAll("./temp", 0766)
	if err != nil {
		glog.Errorln(err)
	}
	err = os.MkdirAll("./log", 0766)
	if err != nil {
		glog.Errorln(err)
	}
	nodeHandler.IDCounter = 0

	//Pages
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/index", indexPageHandler)
	http.HandleFunc("/file", filePageHandler)
	http.HandleFunc("/node", nodeEnterHandler)

	http.HandleFunc("/greet", greetHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/restore", restoreHandler)

	http.HandleFunc("/download", downloadHandler)

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))

	go func() {
		for {
			now := time.Now().UnixNano() / 1000000
			for key, value := range nodeHandler.AllNodes {
				//glog.Infof("Now : %d, value.Lasttime : %d, delta : %d, delta-6000 : %d", now, value.LastTime, now - value.LastTime, now - value.LastTime - 6000)
				if now-value.LastTime > 6000 {
					node := value
					node.Status = false
					nodeHandler.AllNodes[key] = node
					OnDeleted(&node)
				}
			}
			time.Sleep(4 * time.Second)
		}
	}()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		glog.Errorln(err)
	}
}

//OnDeleted 节点丢失处理
func OnDeleted(node *nodeHandler.Node) {
	//glog.Info("OnDeleted")
	var isEmpty = false
	for _, value := range nodeHandler.EmptyNodes {
		if value == node.ID {
			glog.Info("Empty Node Found")
			isEmpty = true
		}
	}

	if isEmpty {
		//If Empty Node Lost, delete from all & Empty Slices
		delete(nodeHandler.AllNodes, node.ID)
		index := sysTool.GetIndexInAll(len(nodeHandler.EmptyNodes), func(i int) bool {
			return nodeHandler.EmptyNodes[i] == node.ID
		})
		nodeHandler.EmptyNodes = append(nodeHandler.EmptyNodes[:index], nodeHandler.EmptyNodes[index+1:]...)
		glog.Info("已删除空节点")
	} else {
		var lostExist = false
		for _, value := range nodeHandler.LostNodes {
			if value == node.ID {
				//glog.Info("该节点为已丢失节点 : %d", node.ID)
				lostExist = true
			}
		}
		if !lostExist {
			nodeHandler.AddToLost(node.ID)
			sysTool.SysConfig.Status = false
			glog.Infof("新的丢失节点, SysConfigure 变为 false, 丢失节点ID : %d", node.ID)
		}
	}
}
