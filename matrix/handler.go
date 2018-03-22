package main

import (
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"sync"

	"github.com/Vaaaas/MatrixFS/filehandler"
	"github.com/Vaaaas/MatrixFS/nodeHandler"
	"github.com/Vaaaas/MatrixFS/sysTool"
	"github.com/golang/glog"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {

		if !sysTool.SysConfigured() {
			glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
		} else {
			if !nodeHandler.NodeConfigured() {
				glog.Infoln("URL: " + r.URL.Path + " Node not configured, redirect to node.html")
				http.Redirect(w, r, "/node", http.StatusFound)
			} else {
				glog.Infoln("URL: " + r.URL.Path + "configured, redirect to file.html")
				http.Redirect(w, r, "/file", http.StatusFound)
			}
		}
	} else {
		if r.URL.Path == "/favicon.ico" {
			glog.Infoln("[/favicon.ico] " + r.URL.Path)
			http.ServeFile(w, r, "favicon.ico")
		} else {
			glog.Infoln("[/] " + r.URL.Path)
			t, err := template.ParseFiles("view/404.html")
			if err != nil {
				glog.Errorln(err)
			}
			t.Execute(w, nil)
		}
	}
}

func indexPageHandler(w http.ResponseWriter, r *http.Request) {
	if sysTool.SysConfigured() {
		if !nodeHandler.NodeConfigured() {
			glog.Infoln("URL: " + r.URL.Path + " Node not configured, redirect to node.html")
			http.Redirect(w, r, "/node", http.StatusFound)
		} else {
			glog.Infoln("URL: " + r.URL.Path + " configured, redirect to file.html")
			http.Redirect(w, r, "/file", http.StatusFound)
		}
	} else {
		t, err := template.ParseFiles("view/index.html")
		if err != nil {
			glog.Errorln(err)
		}
		t.Execute(w, nil)
	}
}

func nodeEnterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	glog.Infoln("[File Page] method: ", r.Method)
	if r.Method == "GET" {
		if !sysTool.SysConfigured() {
			glog.Infoln("URL: " + r.URL.Path + "not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
			return
		}
	} else {
		if !sysTool.SysConfigured() {
			glog.Infoln("[Configure-Fault]" + r.Form["faultNumber"][0])
			glog.Infoln("[Configure-Row]" + r.Form["rowNumber"][0])
			faultNum, err := strconv.Atoi(r.Form["faultNumber"][0])
			if err != nil {
				glog.Error("faultNumber参数转换为int失败")
			}
			rowNum, err := strconv.Atoi(r.Form["rowNumber"][0])
			if err != nil {
				glog.Error("rowNumber参数转换为int失败")
			}
			//TODO: 初始化系统设定后, 上传文件前仍然需要确认存储节点?
			sysTool.InitConfig(faultNum, rowNum)
		}
	}
	allNodesListTemp:=nodeHandler.AllNodes.Items()
	var resultMap =make(map[uint]nodeHandler.Node)
	for key, value := range allNodesListTemp {
		converted, ok := value.(nodeHandler.Node)
		key,_:=key.(uint)
		if ok {
			resultMap[key] = converted
		}
	}
	data := struct {
		Nodes        map[uint]nodeHandler.Node
		SystemStatus bool
	}{
		Nodes:        resultMap,
		SystemStatus: sysTool.SysConfig.Status,
	}

	glog.Infof("System Status : %s, Length of AllNodes : %d", strconv.FormatBool(sysTool.SysConfig.Status), nodeHandler.AllNodes.Count())
	t, err := template.ParseFiles("view/node.html")
	if err != nil {
		glog.Errorln(err)
	}
	t.Execute(w, data)
}

func filePageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	glog.Infoln("[File Page] method: ", r.Method)
	if !sysTool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + "not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
		return
	}
	glog.Infof("Length of AllFiles : %d", filehandler.AllFiles.Count())
	allFileListTemp:=filehandler.AllFiles.Items()
	var resultList []*filehandler.File
	for _, value := range allFileListTemp {
		result := value.(*filehandler.File)
		resultList = append(resultList,result)
		glog.Infoln("[Show File Page]File name Value: "+result.FileFullName)
	}
	t, err := template.ParseFiles("view/file.html")
	if err != nil {
		glog.Errorln(err)
	}
	t.Execute(w, resultList)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[UPLOAD] method: ", r.Method)
	if r.Method == "GET" {
		glog.Infoln("[/UPLOAD] " + r.URL.Path)
		t, err := template.ParseFiles("view/404.html")
		if err != nil {
			glog.Errorln(err)
		}
		t.Execute(w, nil)
	} else {
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
		//开始生成分块阵列
		file01 := fileHandle("temp/" + handler.Filename)
		f.Close()
		file.Close()
		//删除所有临时文件
		file01.DeleteAllTempFiles()
		glog.Infof("File upload & init finished %s, Content-Type: %s", handler.Filename, r.Header.Get("Content-Type"))
		http.Redirect(w, r, "/file", http.StatusFound)
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[DOWNLOAD] " + r.URL.Path)
	r.ParseForm()
	glog.Infoln("[DOWNLOAD] method: ", r.Method)

	if !sysTool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else {
		glog.Infoln("[Form-File_Name]" + r.Form["fileName"][0])
		fileName := r.Form["fileName"][0]
		targetFile := filehandler.AllFiles.Get(fileName).(*filehandler.File)
		glog.Infoln("[Target File Struct]FullName is :")
		targetFile.CollectFiles()
		if sysTool.SysConfig.Status == false {
			//TODO: 降级读
			// var recFinish = true
			// for index := range nodeHandler.LostNodes {
			// 	var result bool
			// 	glog.Infof("需要检测节点 ID : %d", nodeHandler.LostNodes[index])
			// 	result = nodeHandler.safeMap[nodeHandler.LostNodes[index]].Old_DetectNode(*targetFile)
			// 	recFinish = recFinish && result
			// }
			// for !recFinish {
			// 	recFinish = true
			// 	for index := range nodeHandler.LostNodes {
			// 		var result bool
			// 		glog.Infof("需要检测节点 ID : %d", nodeHandler.LostNodes[index])
			// 		result = nodeHandler.safeMap[nodeHandler.LostNodes[index]].Old_DetectNode(*targetFile)
			// 		recFinish = recFinish && result
			// 	}
			// }
		}
		targetFile.GetFile("temp/")
		w.Header().Set("Content-Disposition", "attachment; filename="+targetFile.FileFullName)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		http.ServeFile(w, r, "temp/"+targetFile.FileFullName)
		targetFile.DeleteAllTempFiles()
		glog.Infoln("[Download] " + targetFile.FileFullName + " Finished, Content-Type: " + r.Header.Get("Content-Type"))
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[DELETE] " + r.URL.Path)
	r.ParseForm()
	glog.Infoln("[DELETE] method: ", r.Method)

	if !sysTool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else {
		glog.Infoln("[Form-File_Name]" + r.Form["fileName"][0])
		fileName := r.Form["fileName"][0]
		targetFile := filehandler.AllFiles.Get(fileName).(*filehandler.File)
		targetFile.DeleteSlices()
		targetFile.DeleteAllTempFiles()
		filehandler.AllFiles.Delete(targetFile.FileFullName)
		glog.Infoln("[Download] " + targetFile.FileFullName + " Finished")
		http.Redirect(w, r, "/file", http.StatusFound)
	}

}

func fileHandle(source string) *filehandler.File {
	glog.Infoln("Start FileHandler")
	var file01 filehandler.File
	//分析原始文件属性
	file01.Init(source)
	//分割文件名
	name, ext := file01.SliceFileName()
	glog.Infof("File %s init finished", name+ext)
	//生成数据分块阵列
	file01.InitDataFiles()
	//编码生成校验分块阵列
	file01.InitRddtFiles()

	//将分块文件发送至存储节点
	file01.SendToNode()
	return &file01
}

func greetHandler(w http.ResponseWriter, r *http.Request) {
	//在Master中建立空Node变量
	var node nodeHandler.Node
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

	//glog.Infoln("Existence Status : " + strconv.FormatBool(existed))

	if existed {
		//glog.Infof("Node [%d] already Existed", node.ID)
		var volume = node.Volume
		node = nodeHandler.AllNodes.Get(node.ID).(nodeHandler.Node)
		node.Volume = volume
		//node.Status = true
		//glog.Infof("Before Time : %d", node.LastTime)
		node.LastTime = time.Now().UnixNano() / 1000000
		nodeHandler.AllNodes.Set(node.ID,node)
		//glog.Infof("Refresh Time : %d", sysTool.safeMap[node.ID].LastTime)
	} else {
		nodeHandler.IDCounter.PlusSafeID()
		node.ID = nodeHandler.IDCounter.GetSafeID()

		glog.Infof("Hello %d\n", node.ID)

		nodeHandler.EmptyNodes = append(nodeHandler.EmptyNodes, node.ID)
		node.LastTime = time.Now().UnixNano() / 1000000

		nodeHandler.AllNodes.Set(node.ID,node)

		if !node.AppendNode() {
			glog.Infof("Still in Empty Slice %+v", node)
		} else {
			glog.Infof("[Removed from empty]%+v", node)
		}
	}

	w.Header().Set("ID", strconv.Itoa((int)(node.ID)))
	w.WriteHeader(http.StatusOK)
}

func restoreHandler(w http.ResponseWriter, r *http.Request) {
	if !sysTool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else if sysTool.SysConfig.Status {
		glog.Infoln("系统正常运行，无需修复.")
		http.Redirect(w, r, "/node", http.StatusFound)
	} else if len(nodeHandler.LostNodes) > sysTool.SysConfig.FaultNum {
		glog.Warningf("丢失节点数 : %d", len(nodeHandler.LostNodes))
		t, err := template.ParseFiles("view/info.html")
		data := struct {
			info string
		}{
			info: "丢失节点数超过可容错数.",
		}
		if err != nil {
			glog.Errorln(err)
		}
		t.Execute(w, data)
	} else if len(nodeHandler.EmptyNodes) < len(nodeHandler.LostNodes) {
		t, err := template.ParseFiles("view/info.html")
		data := struct {
			info string
		}{
			info: "没有足够的空节点用于恢复.",
		}
		if err != nil {
			glog.Errorln(err)
		}
		t.Execute(w, data)
	} else {
		//以前：将所有文件收集至中心节点
		glog.Infoln("将空节点转换至丢失节点")

		//处理丢失节点，将空节点转化为丢失节点，为空节点设置新ID
		for i := 0; i < len(nodeHandler.LostNodes); i++ {
			prevLostID := nodeHandler.LostNodes[i]
			empID := nodeHandler.EmptyNodes[i]

			//node : 空节点对象
			node := nodeHandler.AllNodes.Get(empID).(nodeHandler.Node)
			//生成url
			url := "http://" + node.Address.String() + ":" + strconv.Itoa(node.Port) + "/resetid"
			glog.Info("[Reset ID] URL " + url)

			//向空节点发送重设ID请求
			req, err := http.NewRequest("POST", url, nil)
			if err != nil {
				glog.Errorln(err)
				panic(err)
			}
			req.Header.Set("NewID", strconv.Itoa((int)(prevLostID)))
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				glog.Errorln(err)
				panic(err)
			}
			defer resp.Body.Close()
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				glog.Errorln(err)
				panic(err)
			}
			glog.Info(resp.Status)
			glog.Info(string(respBody))
			glog.Infof("空节点 ID : %d, 丢失节点ID : %d", empID, prevLostID)
			//转化完成，得到新节点信息
			node.ID = prevLostID
			node.Status = false
			nodeHandler.AllNodes.Set(prevLostID,node)
			glog.Infof("用于恢复的节点ID : %d", node.ID)
		}

		//TODO: 开始解码恢复 循环所有文件，为每个文件开一个线程用于恢复
		var waitGroup sync.WaitGroup

		allFileListTemp:=filehandler.AllFiles.Items()
		for _, value := range allFileListTemp {
			converted, ok := value.(filehandler.File)
			if ok {
				waitGroup.Add(1)
				go func() {
					defer waitGroup.Done()
					//执行对单个文件的恢复
					converted.LostHandle()
				}()
			}
		}

		//阻塞，等待全部文件恢复完成
		waitGroup.Wait()

		//恢复完成，丢失节点状态设为正常
		for i := 0; i < len(nodeHandler.LostNodes); i++ {
			node := nodeHandler.AllNodes.Get(nodeHandler.LostNodes[i]).(nodeHandler.Node)
			node.Status = true
			nodeHandler.AllNodes.Set(nodeHandler.LostNodes[i],node)
		}

		//删除已转化的空节点
		for i := 0; i < len(nodeHandler.LostNodes); i++ {
			nodeHandler.AllNodes.Delete(nodeHandler.EmptyNodes[0])
			nodeHandler.EmptyNodes = append(nodeHandler.EmptyNodes[:0], nodeHandler.EmptyNodes[1:]...)
		}
		//清空失效节点列表
		nodeHandler.LostNodes = []uint{}
		//系统状态设为正常
		sysTool.SysConfig.Status = true
		//前段显示信息提示页面
		t, err := template.ParseFiles("view/info.html")
		data := struct {
			info string
		}{
			info: "Restore Finished!",
		}
		if err != nil {
			glog.Errorln(err)
		}
		t.Execute(w, data)
	}
}
