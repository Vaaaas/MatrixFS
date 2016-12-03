package main

import (
	"github.com/golang/glog"
	"flag"
	"github.com/Vaaaas/MatrixFS/SysConfig"
	"github.com/Vaaaas/MatrixFS/File"
	"net/http"
	"html/template"
	"strconv"
	"os"
	"io"
	"encoding/json"
	"fmt"
	"github.com/Vaaaas/MatrixFS/NodeStruct"
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
	NodeStruct.IDCounter=0

	//Pages
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/index", indexPageHandler)
	http.HandleFunc("/file", filePageHandler)
	http.HandleFunc("/node", nodeHandler)

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/greet", greetHandler)

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		glog.Errorln(err)
	}
}

func sysConfigured() bool {
	return SysConfig.SysConfig.FaultNum != 0 && SysConfig.SysConfig.RowNum != 0
}

func nodeConfigured() bool {
	return len(NodeStruct.AllNodes) != 0
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {

		if !sysConfigured() {
			glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
		} else {
			if !nodeConfigured() {
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
			if (err != nil) {
				glog.Errorln(err)
			}
			t.Execute(w, nil)
		}
	}
}

func indexPageHandler(w http.ResponseWriter, r *http.Request) {
	if sysConfigured() {
		if !nodeConfigured() {
			glog.Infoln("URL: " + r.URL.Path + " Node not configured, redirect to nnode.html")
			http.Redirect(w, r, "/node", http.StatusFound)
		} else {
			glog.Infoln("URL: " + r.URL.Path + " configured, redirect to file.html")
			http.Redirect(w, r, "/file", http.StatusFound)
		}
	} else {
		t, err := template.ParseFiles("view/index.html")
		if (err != nil) {
			glog.Errorln(err)
		}
		t.Execute(w, nil)
	}
}

func nodeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	glog.Infoln("[File Page] method: ", r.Method)
	if r.Method == "GET" {
		if !sysConfigured() {
			glog.Infoln("URL: " + r.URL.Path + "not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
			return
		}
	} else {
		if !sysConfigured() {
			glog.Infoln("[Configure-Fault]" + r.Form["faultNumber"][0])
			glog.Infoln("[Configure-Row]" + r.Form["rowNumber"][0])
			faultNum, err := strconv.Atoi(r.Form["faultNumber"][0])
			if (err != nil) {
				glog.Error("faultNumber参数转换为int失败")
			}
			rowNum, err := strconv.Atoi(r.Form["rowNumber"][0])
			if (err != nil) {
				glog.Error("rowNumber参数转换为int失败")
			}
			//todo : After init System config, still need to confirm the nodes before upload file
			SysConfig.InitConfig(faultNum, rowNum)
		}
	}

	glog.Infof("Length of AllNodes : %d", len(NodeStruct.AllNodes))
	fmt.Println(NodeStruct.AllNodes)
	t, err := template.ParseFiles("view/node.html")
	if (err != nil) {
		glog.Errorln(err)
	}
	t.Execute(w, NodeStruct.AllNodes)
}

func filePageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	glog.Infoln("[File Page] method: ", r.Method)
	if !sysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + "not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
		return
	} else {
		glog.Infof("Length of AllFiles : %d", len(File.AllFiles))
		t, err := template.ParseFiles("view/file.html")
		if (err != nil) {
			glog.Errorln(err)
		}
		t.Execute(w, File.AllFiles)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[UPLOAD] method: ", r.Method)
	if r.Method == "GET" {
		glog.Infoln("[/UPLOAD] " + r.URL.Path)
		t, err := template.ParseFiles("view/404.html")
		if (err != nil) {
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
		defer file.Close()
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
		defer f.Close()
		io.Copy(f, file)
		fileHandle("temp/" + handler.Filename)
		glog.Infof("File upload & init finished, redirect to file page : %s, Content-Type: %s", handler.Filename, r.Header.Get("Content-Type"))
		http.Redirect(w, r, "/file", http.StatusFound)
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[DOWNLOAD] " + r.URL.Path)
	r.ParseForm()
	glog.Infoln("[DOWNLOAD] method: ", r.Method)

	if !sysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else {
		glog.Infoln("[Form-File_Name]" + r.Form["fileName"][0])
		fileName := r.Form["fileName"][0]
		targetFile := findFileInAll(fileName)
		targetFile.GetFile("temp/")
		w.Header().Set("Content-Disposition", "attachment; filename=" + fileName)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		http.ServeFile(w, r, "temp/" + fileName)
		glog.Infoln("[Download] " + targetFile.FileFullName + " Finished, Content-Type: " + r.Header.Get("Content-Type"))
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("[DELETE] " + r.URL.Path)
	r.ParseForm()
	glog.Infoln("[DELETE] method: ", r.Method)

	if !sysConfigured() {
		glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	} else {
		glog.Infoln("[Form-File_Name]" + r.Form["fileName"][0])
		fileName := r.Form["fileName"][0]
		targetFile := findFileInAll(fileName)
		targetFile.DeleteAllTempFiles()
		glog.Infoln("[Download] " + targetFile.FileFullName + " Finished")
		http.Redirect(w, r, "/file", http.StatusFound)
	}

}

func fileHandle(source string) {
	glog.Infoln("Start FileHandler")
	var file01 File.File
	file01.Init(source)
	name, ext := file01.SliceFileName()
	glog.Infof("File %s init finished", name + ext)
	file01.InitDataFiles()
	file01.InitRddtFiles()
}

func findFileInAll(name string) (file File.File) {
	for _, tempFile := range File.AllFiles {
		if tempFile.FileFullName == name {
			return tempFile
		}
	}
	return
}

func greetHandler(w http.ResponseWriter, r *http.Request) {
	var node NodeStruct.Node
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&node)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	fmt.Println(node)
	NodeStruct.IDCounter++
	node.ID=NodeStruct.IDCounter
	fmt.Printf("Hello %d", node.ID)

	NodeStruct.AllNodes = append(NodeStruct.AllNodes, node)
}