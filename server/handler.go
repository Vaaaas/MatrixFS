package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
	"sync"

	"github.com/Vaaaas/MatrixFS/filehandler"
	"github.com/Vaaaas/MatrixFS/nodehandler"
	"github.com/Vaaaas/MatrixFS/util"
	"github.com/golang/glog"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		if !util.SysConfigured() {
			glog.Warningln("URL: " + r.URL.Path + " not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
		} else {
			if !nodehandler.NodeConfigured() {
				glog.Warningln("URL: " + r.URL.Path + " Node not configured, redirect to node.html")
				http.Redirect(w, r, "/node", http.StatusFound)
			} else {
				glog.Infoln("URL: " + r.URL.Path + "configured, redirect to file.html")
				http.Redirect(w, r, "/file", http.StatusFound)
			}
		}
	} else {
		//TODO: 什么情况下到这里
		glog.Errorln("[/] " + r.URL.Path)
		F0fTpl.Execute(w, nil)
	}
}

func IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	if util.SysConfigured() {
		if !nodehandler.NodeConfigured() {
			glog.Warningln("URL: " + r.URL.Path + " Node not configured, redirect to node.html")
			http.Redirect(w, r, "/node", http.StatusFound)
		} else {
			glog.Infoln("URL: " + r.URL.Path + " configured, redirect to file.html")
			http.Redirect(w, r, "/file", http.StatusFound)
		}
	} else {
		IndexTpl.Execute(w, nil)
	}
}

func NodeEnterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "GET" {
		if !util.SysConfigured() {
			glog.Warningln("URL: " + r.URL.Path + "not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
			return
		}
	} else {
		if !util.SysConfigured() {
			faultNum, err := strconv.Atoi(r.Form["faultNumber"][0])
			if err != nil {
				glog.Error("faultNumber参数转换为int失败")
			}
			rowNum, err := strconv.Atoi(r.Form["rowNumber"][0])
			if err != nil {
				glog.Error("rowNumber参数转换为int失败")
			}
			//TODO: 初始化系统设定后, 上传文件前仍然需要确认存储节点?
			util.InitConfig(faultNum, rowNum)
		}
	}
	allNodesListTemp := nodehandler.AllNodes.Items()
	var resultMap = make(map[uint]nodehandler.Node)
	for key, value := range allNodesListTemp {
		converted, ok := value.(nodehandler.Node)
		key, _ := key.(uint)
		if ok {
			resultMap[key] = converted
		}
	}
	data := struct {
		Nodes        map[uint]nodehandler.Node
		SystemStatus bool
	}{
		Nodes:        resultMap,
		SystemStatus: util.SysConfig.Status,
	}
	NodeTpl.Execute(w, data)
}

func FilePageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if !util.SysConfigured() {
		glog.Warningln("URL: " + r.URL.Path + "not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
		return
	}
	allFileListTemp := filehandler.AllFiles.Items()
	var resultList []filehandler.File
	for _, value := range allFileListTemp {
		result := value.(*filehandler.File)
		resultList = append(resultList, *result)
	}

	FileTpl.Execute(w, resultList)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		glog.Warningln("[/UPLOAD] GET " + r.URL.Path)
		F0fTpl.Execute(w, nil)
	} else {
		glog.Infoln("[UPLOAD] " + r.URL.Path)
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadInput")
		if err != nil {
			glog.Error(err)
			panic(err)
		}

		if handler.Filename == "" {
			http.Redirect(w, r, "/file", http.StatusFound)
			glog.Warningln("empty file")
			return
		}
		//在临时文件夹创建原始文件的副本
		f, err := os.OpenFile("temp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			glog.Error(err)
			panic(err)
		}
		//复制原始文件
		io.Copy(f, file)
		glog.Infoln("[处理上传文件]生成副本完成")
		//开始生成分块阵列
		file01 := fileHandle("temp/" + handler.Filename)
		f.Close()
		file.Close()
		//删除所有临时文件
		file01.DeleteAllTempFiles()
		glog.Infof("File upload & init finished %s, Content-Type: %s", handler.Filename, r.Header.Get("Content-Type"))
		info:= "上传文件完成"
		InfoTpl.Execute(w, info)
	}
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if !util.SysConfigured() {
		glog.Warningln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else {
		fileName := r.Form["fileName"][0]
		targetFile := filehandler.AllFiles.Get(fileName).(*filehandler.File)
		glog.Infoln("[/DOWNLOAD]开始收集数据分块 FullName is :" + fileName)
		targetFile.CollectFiles()
		glog.Infoln("[DOWNLOAD]收集数据分块完成")
		glog.Infoln("[DOWNLOAD]开始将数据分块写入副本")
		targetFile.GetFile("temp/")
		glog.Infoln("[DOWNLOAD]数据分块写入副本完成")
		w.Header().Set("Content-Disposition", "attachment; filename="+targetFile.FileFullName)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		http.ServeFile(w, r, "temp/"+targetFile.FileFullName)
		targetFile.DeleteAllTempFiles()
		glog.Infoln("[Download] " + targetFile.FileFullName + " Finished, Content-Type: " + r.Header.Get("Content-Type"))
	}
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if !util.SysConfigured() {
		glog.Warningln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else {
		glog.Infoln("[DELETE]" + r.Form["fileName"][0])
		fileName := r.Form["fileName"][0]
		targetFile := filehandler.AllFiles.Get(fileName).(*filehandler.File)
		targetFile.DeleteSlices()
		targetFile.DeleteAllTempFiles()
		filehandler.AllFiles.Delete(targetFile.FileFullName)
		glog.Infoln("[Download] " + targetFile.FileFullName + " Finished")
		http.Redirect(w, r, "/file", http.StatusFound)
	}

}

func GreetHandler(w http.ResponseWriter, r *http.Request) {
	//在Master中建立空Node变量
	var node nodehandler.Node
	var existed = false
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&node)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if node.ID != 0 {
		existed = true
	}
	if existed {
		var volume = node.Volume
		node = nodehandler.AllNodes.Get(node.ID).(nodehandler.Node)
		node.Volume = volume
		node.LastTime = time.Now().UnixNano() / 1000000
		nodehandler.AllNodes.Set(node.ID, node)
	} else {
		nodehandler.IDCounter.PlusSafeID()
		node.ID = nodehandler.IDCounter.GetSafeID()
		glog.Infof("Hello %d\n", node.ID)
		nodehandler.EmptyNodes = append(nodehandler.EmptyNodes, node.ID)
		node.LastTime = time.Now().UnixNano() / 1000000
		nodehandler.AllNodes.Set(node.ID, node)
		if !node.AppendNode() {
			glog.Infof("Still in Empty Slice %+v", node)
		} else {
			glog.Infof("[Removed from empty] %+v", node)
		}
	}
	w.Header().Set("ID", strconv.Itoa((int)(node.ID)))
	w.WriteHeader(http.StatusOK)
}

func RestoreHandler(w http.ResponseWriter, r *http.Request) {
	if !util.SysConfigured() {
		glog.Warningln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else if util.SysConfig.Status {
		glog.Infoln("系统正常运行，无需修复.")
		http.Redirect(w, r, "/node", http.StatusFound)
	} else if len(nodehandler.LostNodes) > util.SysConfig.FaultNum {
		glog.Warningf("丢失节点数 : %d", len(nodehandler.LostNodes))
		info:="丢失节点数超过可容错数."
		InfoTpl.Execute(w, info)
	} else if len(nodehandler.EmptyNodes) < len(nodehandler.LostNodes) {
		glog.Warningf("丢失节点数 : %d 大于最大容错数", len(nodehandler.LostNodes))
		info:="没有足够的空节点用于恢复."
		InfoTpl.Execute(w, info)
	} else {
		glog.Infoln("开始将空节点转换至丢失节点")
		//处理丢失节点，将空节点转化为丢失节点，为空节点设置新ID
		for i := 0; i < len(nodehandler.LostNodes); i++ {
			prevLostID := nodehandler.LostNodes[i]
			empID := nodehandler.EmptyNodes[i]
			nodehandler.EmptyNodeToLostNode(empID, prevLostID)
		}
		glog.Infoln("将空节点转换至丢失节点完成")

		glog.Infoln("开始解码恢复 循环所有文件，为每个文件开一个线程用于恢复")
		var waitGroup sync.WaitGroup
		allFileListTemp := filehandler.AllFiles.Items()
		for _, value := range allFileListTemp {
			converted, _ := value.(*filehandler.File)
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				//执行对单个文件的恢复
				glog.Infof("[解码恢复] 文件名 %s, 文件大小 %d", converted.FileFullName, converted.Size)
				converted.LostHandle()
			}()
		}

		//阻塞，等待全部文件恢复完成
		waitGroup.Wait()

		//恢复完成，丢失节点状态设为正常
		for i := 0; i < len(nodehandler.LostNodes); i++ {
			node := nodehandler.AllNodes.Get(nodehandler.LostNodes[i]).(nodehandler.Node)
			node.Status = true
			nodehandler.AllNodes.Set(nodehandler.LostNodes[i], node)
		}

		//删除已转化的空节点
		for i := 0; i < len(nodehandler.LostNodes); i++ {
			//glog.Warningf("[Delete removed empty nodes] i = %d",i)
			nodehandler.AllNodes.Delete(nodehandler.EmptyNodes[0])
			if len(nodehandler.EmptyNodes) > 0 {
				nodehandler.EmptyNodes = append(nodehandler.EmptyNodes[:0], nodehandler.EmptyNodes[1:]...)
			}
		}
		//清空失效节点列表
		nodehandler.LostNodes = []uint{}
		//系统状态设为正常
		util.SysConfig.Status = true
		//前端显示信息提示页面
		info:="系统恢复完成"
		InfoTpl.Execute(w, info)
	}
}

func fileHandle(source string) *filehandler.File {
	glog.Infoln("Start FileHandler")
	var file01 filehandler.File
	//分析原始文件属性
	file01.Init(source)
	//分割文件名
	name, ext := file01.SliceFileName()
	glog.Infof("File %s.%s init finished", name, ext)
	//生成数据分块阵列
	glog.Infoln("[处理上传文件]开始生成数据阵列")
	file01.InitDataFiles()
	glog.Infoln("[处理上传文件]生成数据阵列完成")
	//编码生成校验分块阵列
	glog.Infoln("[处理上传文件]开始生成校验阵列")
	file01.InitRddtFiles()
	glog.Infoln("[处理上传文件]生成校验阵列完成")
	//将分块文件发送至存储节点
	glog.Infoln("[处理上传文件]开始发送分块阵列至节点")
	file01.SendToNode()
	glog.Infoln("[处理上传文件]发送分块阵列至节点完成")
	return &file01
}
