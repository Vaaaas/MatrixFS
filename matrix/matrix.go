package main

import (
	"flag"
	"net/http"
	"os"
	"time"
	"github.com/golang/glog"
	"github.com/Vaaaas/MatrixFS/Tool"
)

func main() {
	//when debug, log_dir="./log"
	flag.Parse()
	//Trigger on exit, write log into files
	defer glog.Flush()
	//Init system config(Moved to indexPageHandler)
	//SysConfig.InitConfig(fault, row)
	glog.Info("Server start here")

	//testFileHandle()
	err := os.MkdirAll("./temp", 0766)
	if err != nil {
		glog.Errorln(err)
	}
	err = os.MkdirAll("./log", 0766)
	if err != nil {
		glog.Errorln(err)
	}
	Tool.IDCounter = 0

	//Pages
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/index", indexPageHandler)
	http.HandleFunc("/file", filePageHandler)
	http.HandleFunc("/node", nodeHandler)

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
			for key, value := range Tool.AllNodes {
				glog.Infof("Now : %d, value.Lasttime : %d, delta : %d, delta-6000 : %d", now, value.LastTime, now - value.LastTime, now - value.LastTime - 6000)
				if now - value.LastTime > 6000 {
					node := value
					node.Status = false
					Tool.AllNodes[key] = node
					OnDeleted(&node)
				} else {
					node := value
					node.Status = true
					Tool.AllNodes[key] = node
				}
			}
			time.Sleep(4 * time.Second)
		}
	}()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		glog.Errorln(err)
	}
}

func OnDeleted(node *Tool.Node) {
	glog.Info("OnDeleted")
	var isEmpty = false
	for _, value := range Tool.EmptyNodes {
		if value == node.ID {
			glog.Info("Empty Node Found")
			isEmpty = true
		}
	}

	if isEmpty {
		//If Empty Node Lost, delete from all & Empty Slices
		delete(Tool.AllNodes, node.ID)
		index := Tool.GetFileIndexInAll(len(Tool.EmptyNodes), func(i int) bool {
			return Tool.EmptyNodes[i] == node.ID
		})
		Tool.EmptyNodes = append(Tool.EmptyNodes[:index], Tool.EmptyNodes[index + 1:]...)
		glog.Info("Empty Node Deleted")
	} else {
		var lostExist = false
		for _, value := range Tool.LostNodes {
			if value == node.ID {
				glog.Info("Lost Node Found")
				lostExist = true
			}
		}
		if !lostExist {
			Tool.AddToLost(node.ID)
			Tool.SysConfig.Status = false
			glog.Info("New Lost Node, SysConfigure turned to false")
		}
	}
}