package main

import (
	"net/http"
	"github.com/Vaaaas/MatrixFS/Tool"
	"strconv"
	"os"
	"encoding/json"
	"time"
	"github.com/golang/glog"
	"html/template"
	"io"
	"io/ioutil"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {

		if !Tool.SysConfigured() {
			glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
		} else {
			if !Tool.NodeConfigured() {
				glog.Infoln("URL: " + r.URL.Path + " Node not configured, redirect to nnode.html")
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
	if Tool.SysConfigured() {
		if !Tool.NodeConfigured() {
			glog.Infoln("URL: " + r.URL.Path + " Node not configured, redirect to nnode.html")
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

func nodeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	glog.Infoln("[File Page] method: ", r.Method)
	if r.Method == "GET" {
		if !Tool.SysConfigured() {
			glog.Infoln("URL: " + r.URL.Path + "not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
			return
		}
	} else {
		if !Tool.SysConfigured() {
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
			//todo : After init System config, still need to confirm the nodes before upload file
			Tool.InitConfig(faultNum, rowNum)
		}
	}
	data := struct {
		Nodes        map[uint]Tool.Node
		SystemStatus bool
	}{
		Nodes:        Tool.AllNodes,
		SystemStatus: Tool.SysConfig.Status,
	}

	glog.Infof("System Status : %s, Length of AllNodes : %d", strconv.FormatBool(Tool.SysConfig.Status), len(Tool.AllNodes))
	t, err := template.ParseFiles("view/node.html")
	if err != nil {
		glog.Errorln(err)
	}
	t.Execute(w, data)
}

func filePageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	glog.Infoln("[File Page] method: ", r.Method)
	if !Tool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + "not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
		return
	} else {
		//glog.Infof("Length of AllFiles : %d", len(Tool.AllFiles))
		t, err := template.ParseFiles("view/file.html")
		if err != nil {
			glog.Errorln(err)
		}
		t.Execute(w, Tool.AllFiles)
	}
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
		//fmt.Fprintf(w, "%v", handler.Header)
		if handler.Filename == "" {
			http.Redirect(w, r, "/file", http.StatusFound)
			glog.Warningln("empty file")
			return
		}
		f, err := os.OpenFile("temp/" + handler.Filename, os.O_WRONLY | os.O_CREATE, 0666)
		if err != nil {
			glog.Error(err)
			panic(err)
		}
		io.Copy(f, file)

		file01 := fileHandle("temp/" + handler.Filename)
		f.Close()
		file.Close()
		file01.DeleteAllTempFiles()
		glog.Infof("File upload & init finished, redirect to file page : %s, Content-Type: %s", handler.Filename, r.Header.Get("Content-Type"))
		http.Redirect(w, r, "/file", http.StatusFound)
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[DOWNLOAD] " + r.URL.Path)
	r.ParseForm()
	glog.Infoln("[DOWNLOAD] method: ", r.Method)

	if !Tool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else {
		glog.Infoln("[Form-File_Name]" + r.Form["fileName"][0])
		fileName := r.Form["fileName"][0]
		targetFile := Tool.FindFileInAll(fileName)
		targetFile.CollectFiles()
		if Tool.SysConfig.Status==false {
			var recFinish = true
			//todo : Attention to this expression
			for index := range Tool.LostNodes {
				var result bool
				glog.Infof("需要检测节点 ID : %d", Tool.LostNodes[index])
				result = Tool.AllNodes[Tool.LostNodes[index]].DetectNode(*targetFile)
				recFinish = recFinish && result
			}
			for !recFinish {
				recFinish = true
				//todo : Attention to this expression
				for index := range Tool.LostNodes {
					var result bool
					glog.Infof("需要检测节点 ID : %d", Tool.LostNodes[index])
					result = Tool.AllNodes[Tool.LostNodes[index]].DetectNode(*targetFile)
					recFinish = recFinish && result
				}
			}
		}
		targetFile.GetFile("temp/")
		w.Header().Set("Content-Disposition", "attachment; filename=" + fileName)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		http.ServeFile(w, r, "temp/" + fileName)
		targetFile.DeleteAllTempFiles()
		glog.Infoln("[Download] " + targetFile.FileFullName + " Finished, Content-Type: " + r.Header.Get("Content-Type"))
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[DELETE] " + r.URL.Path)
	r.ParseForm()
	glog.Infoln("[DELETE] method: ", r.Method)

	if !Tool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else {
		glog.Infoln("[Form-File_Name]" + r.Form["fileName"][0])
		fileName := r.Form["fileName"][0]
		targetFile := Tool.FindFileInAll(fileName)
		targetFile.DeleteSlices()
		targetFile.DeleteAllTempFiles()
		index := Tool.GetFileIndexInAll(len(Tool.AllFiles), func(i int) bool {
			return Tool.AllFiles[i].FileFullName == targetFile.FileFullName
		})
		Tool.AllFiles = append(Tool.AllFiles[:index], Tool.AllFiles[index + 1:]...)
		glog.Infoln("[Download] " + targetFile.FileFullName + " Finished")
		http.Redirect(w, r, "/file", http.StatusFound)
	}

}

func fileHandle(source string) *Tool.File {
	glog.Infoln("Start FileHandler")
	var file01 Tool.File
	//分析原始文件属性
	file01.Init(source)
	//分割文件名
	name, ext := file01.SliceFileName()
	glog.Infof("File %s init finished", name + ext)
	//生成数据分块阵列
	file01.InitDataFiles()
	//编码生成校验分块阵列
	file01.InitRddtFiles()

	// todo : send slices to Nodes
	file01.SendToNode()
	return &file01
}

func greetHandler(w http.ResponseWriter, r *http.Request) {
	//在Master中建立空Node变量
	var node Tool.Node
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
		node = Tool.AllNodes[node.ID]
		node.Volume = volume
		//node.Status = true
		//glog.Infof("Before Time : %d", node.LastTime)
		node.LastTime = time.Now().UnixNano() / 1000000
		Tool.AllNodes[node.ID] = node
		//glog.Infof("Refresh Time : %d", Tool.AllNodes[node.ID].LastTime)
	} else {
		Tool.IDCounter++
		node.ID = Tool.IDCounter
		glog.Infof("Hello %d\n", node.ID)

		Tool.EmptyNodes = append(Tool.EmptyNodes, node.ID)
		node.LastTime = time.Now().UnixNano() / 1000000

		Tool.AllNodes[node.ID] = node

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
	if !Tool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else if Tool.SysConfig.Status {
		glog.Infoln("系统正常运行，无需修复.")
		http.Redirect(w, r, "/node", http.StatusFound)
	} else if len(Tool.LostNodes) > Tool.SysConfig.FaultNum {
		glog.Warningf("丢失节点数 : %d", len(Tool.LostNodes))
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
	} else if len(Tool.EmptyNodes) < len(Tool.LostNodes) {
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
		//Need to Collect All files to Master Server
		glog.Infoln("将空节点转换至丢失节点")

		for i := 0; i < len(Tool.LostNodes); i++ {
			prevLostID := Tool.LostNodes[i]
			empID := Tool.EmptyNodes[i]

			url := "http://" + Tool.AllNodes[empID].Address.String() + ":" + strconv.Itoa(Tool.AllNodes[empID].Port) + "/resetid"
			glog.Info("[Reset ID] URL " + url)

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
			respbody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				glog.Errorln(err)
				panic(err)
			}
			glog.Info(resp.Status)
			glog.Info(string(respbody))

			glog.Infof("空节点 ID : %d, 丢失节点ID : %d", empID, prevLostID)
			newNode := Tool.AllNodes[empID]
			newNode.ID = prevLostID
			newNode.Status = false
			Tool.AllNodes[prevLostID] = newNode
			glog.Infof("用于恢复的节点ID : %d", newNode.ID)
		}
		Tool.LostHandle()

		for i := 0; i < len(Tool.LostNodes); i++ {
			node := Tool.AllNodes[Tool.LostNodes[i]]
			node.Status = true
			Tool.AllNodes[Tool.LostNodes[i]] = node
		}

		//Delete Reset EmptyNodes
		for i := 0; i < len(Tool.LostNodes); i++ {
			delete(Tool.AllNodes, Tool.EmptyNodes[0])
			Tool.EmptyNodes = append(Tool.EmptyNodes[:0], Tool.EmptyNodes[1:]...)
		}
		//Delete All LostNodes
		Tool.LostNodes = []uint{}
		Tool.SysConfig.Status = true
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
