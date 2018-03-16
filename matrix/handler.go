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

	"github.com/Vaaaas/MatrixFS/tool"
	"github.com/golang/glog"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {

		if !tool.SysConfigured() {
			glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
		} else {
			if !tool.NodeConfigured() {
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
	if tool.SysConfigured() {
		if !tool.NodeConfigured() {
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
		if !tool.SysConfigured() {
			glog.Infoln("URL: " + r.URL.Path + "not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
			return
		}
	} else {
		if !tool.SysConfigured() {
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
			//todo : 初始化系统设定后, 上传文件前仍然需要确认存储节点
			tool.InitConfig(faultNum, rowNum)
		}
	}
	data := struct {
		Nodes        map[uint]tool.Node
		SystemStatus bool
	}{
		Nodes:        tool.AllNodes,
		SystemStatus: tool.SysConfig.Status,
	}

	glog.Infof("System Status : %s, Length of AllNodes : %d", strconv.FormatBool(tool.SysConfig.Status), len(tool.AllNodes))
	t, err := template.ParseFiles("view/node.html")
	if err != nil {
		glog.Errorln(err)
	}
	t.Execute(w, data)
}

func filePageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	glog.Infoln("[File Page] method: ", r.Method)
	if !tool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + "not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
		return
	}
	//glog.Infof("Length of AllFiles : %d", len(tool.AllFiles))
	t, err := template.ParseFiles("view/file.html")
	if err != nil {
		glog.Errorln(err)
	}
	t.Execute(w, tool.AllFiles)
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
		glog.Infof("File upload & init finished, redirect to file page : %s, Content-Type: %s", handler.Filename, r.Header.Get("Content-Type"))
		http.Redirect(w, r, "/file", http.StatusFound)
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[DOWNLOAD] " + r.URL.Path)
	r.ParseForm()
	glog.Infoln("[DOWNLOAD] method: ", r.Method)

	if !tool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else {
		glog.Infoln("[Form-File_Name]" + r.Form["fileName"][0])
		fileName := r.Form["fileName"][0]
		targetFile := tool.FindFileInAll(fileName)
		targetFile.CollectFiles()
		if tool.SysConfig.Status == false {
			var recFinish = true
			for index := range tool.LostNodes {
				var result bool
				glog.Infof("需要检测节点 ID : %d", tool.LostNodes[index])
				result = tool.AllNodes[tool.LostNodes[index]].DetectNode(*targetFile)
				recFinish = recFinish && result
			}
			for !recFinish {
				recFinish = true
				for index := range tool.LostNodes {
					var result bool
					glog.Infof("需要检测节点 ID : %d", tool.LostNodes[index])
					result = tool.AllNodes[tool.LostNodes[index]].DetectNode(*targetFile)
					recFinish = recFinish && result
				}
			}
		}
		targetFile.GetFile("temp/")
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		http.ServeFile(w, r, "temp/"+fileName)
		targetFile.DeleteAllTempFiles()
		glog.Infoln("[Download] " + targetFile.FileFullName + " Finished, Content-Type: " + r.Header.Get("Content-Type"))
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[DELETE] " + r.URL.Path)
	r.ParseForm()
	glog.Infoln("[DELETE] method: ", r.Method)

	if !tool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else {
		glog.Infoln("[Form-File_Name]" + r.Form["fileName"][0])
		fileName := r.Form["fileName"][0]
		targetFile := tool.FindFileInAll(fileName)
		targetFile.DeleteSlices()
		targetFile.DeleteAllTempFiles()
		index := tool.GetFileIndexInAll(len(tool.AllFiles), func(i int) bool {
			return tool.AllFiles[i].FileFullName == targetFile.FileFullName
		})
		tool.AllFiles = append(tool.AllFiles[:index], tool.AllFiles[index+1:]...)
		glog.Infoln("[Download] " + targetFile.FileFullName + " Finished")
		http.Redirect(w, r, "/file", http.StatusFound)
	}

}

func fileHandle(source string) *tool.File {
	glog.Infoln("Start FileHandler")
	var file01 tool.File
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
	var node tool.Node
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
		node = tool.AllNodes[node.ID]
		node.Volume = volume
		//node.Status = true
		//glog.Infof("Before Time : %d", node.LastTime)
		node.LastTime = time.Now().UnixNano() / 1000000
		tool.AllNodes[node.ID] = node
		//glog.Infof("Refresh Time : %d", tool.AllNodes[node.ID].LastTime)
	} else {
		tool.IDCounter++
		node.ID = tool.IDCounter
		glog.Infof("Hello %d\n", node.ID)

		tool.EmptyNodes = append(tool.EmptyNodes, node.ID)
		node.LastTime = time.Now().UnixNano() / 1000000

		tool.AllNodes[node.ID] = node

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
	if !tool.SysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else if tool.SysConfig.Status {
		glog.Infoln("系统正常运行，无需修复.")
		http.Redirect(w, r, "/node", http.StatusFound)
	} else if len(tool.LostNodes) > tool.SysConfig.FaultNum {
		glog.Warningf("丢失节点数 : %d", len(tool.LostNodes))
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
	} else if len(tool.EmptyNodes) < len(tool.LostNodes) {
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

		for i := 0; i < len(tool.LostNodes); i++ {
			prevLostID := tool.LostNodes[i]
			empID := tool.EmptyNodes[i]

			url := "http://" + tool.AllNodes[empID].Address.String() + ":" + strconv.Itoa(tool.AllNodes[empID].Port) + "/resetid"
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
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				glog.Errorln(err)
				panic(err)
			}
			glog.Info(resp.Status)
			glog.Info(string(respBody))

			glog.Infof("空节点 ID : %d, 丢失节点ID : %d", empID, prevLostID)
			newNode := tool.AllNodes[empID]
			newNode.ID = prevLostID
			newNode.Status = false
			tool.AllNodes[prevLostID] = newNode
			glog.Infof("用于恢复的节点ID : %d", newNode.ID)
		}
		tool.LostHandle()

		for i := 0; i < len(tool.LostNodes); i++ {
			node := tool.AllNodes[tool.LostNodes[i]]
			node.Status = true
			tool.AllNodes[tool.LostNodes[i]] = node
		}

		//Delete Reset EmptyNodes
		for i := 0; i < len(tool.LostNodes); i++ {
			delete(tool.AllNodes, tool.EmptyNodes[0])
			tool.EmptyNodes = append(tool.EmptyNodes[:0], tool.EmptyNodes[1:]...)
		}
		//Delete All LostNodes
		tool.LostNodes = []uint{}
		tool.SysConfig.Status = true
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
