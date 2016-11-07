package main

import (
	"github.com/golang/glog"
	"flag"
	"Vaaaas/MatrixFS/SysConfig"
	"Vaaaas/MatrixFS/File"
	"fmt"
	"net/http"
	"log"
	"html/template"
	"strconv"
	"os"
	"io"
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

	//Pages
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/index", indexPageHandler)
	http.HandleFunc("/file", filePageHandler)
	http.HandleFunc("/node", nodePageHandler)

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/delete", deleteHandler)

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Panic(err)
	}
}

func sysConfigured() bool {
	return SysConfig.SysConfig.FaultNum != 0 && SysConfig.SysConfig.RowNum != 0
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	//todo : Also need to judge if the Nodes are configured, if not redirect to Node manage page
	if r.URL.Path == "/" {
		if !sysConfigured() {
			fmt.Println("not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
		} else {
			fmt.Println("configured, redirect to file.html")
			http.Redirect(w, r, "/file", http.StatusFound)
		}
	} else {
		if r.URL.Path == "/favicon.ico" {
			fmt.Println("[/favicon.ico] " + r.URL.Path)
			http.ServeFile(w, r, "favicon.ico")
		} else {
			fmt.Println("[/] " + r.URL.Path)
			t, err := template.ParseFiles("view/404.html")
			if (err != nil) {
				log.Println(err)
			}
			t.Execute(w, nil)
		}
	}
}

func indexPageHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("view/index.html")
	if (err != nil) {
		log.Println(err)
	}
	t.Execute(w, nil)
}

func filePageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("method: ", r.Method)
	if r.Method == "GET" {
		if !sysConfigured() {
			fmt.Println("not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
			return
		}
	} else {
		if !sysConfigured() {
			fmt.Println(r.Form["faultNumber"][0])
			fmt.Println(r.Form["rowNumber"][0])
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

	fmt.Println(len(File.AllFiles))

	t, err := template.ParseFiles("view/file.html")
	if (err != nil) {
		fmt.Println(err)
	}
	t.Execute(w, File.AllFiles)
}

func nodePageHandler(w http.ResponseWriter, r *http.Request) {

}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method: ", r.Method)
	if r.Method == "GET" {
		fmt.Println("[/UPLOAD] " + r.URL.Path)
		t, err := template.ParseFiles("view/404.html")
		if (err != nil) {
			log.Println(err)
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
		f, err := os.OpenFile("temp/" + handler.Filename, os.O_WRONLY | os.O_CREATE, 0666)
		if err != nil {
			glog.Error(err)
			panic(err)
		}
		defer f.Close()
		io.Copy(f, file)
		fileHandle("temp/" + handler.Filename)
		glog.Infof("File upload & init finished, redirect to file page : %s", handler.Filename)
		http.Redirect(w, r, "/file", http.StatusFound)
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[DOWNLOAD] " + r.URL.Path)
	r.ParseForm()
	fmt.Println("method: ", r.Method)

	if !sysConfigured() {
		fmt.Println("not configured, redirect to index.html")
		http.Redirect(w, r, "/index", http.StatusFound)
	}else{
		fmt.Println(r.Form["fileName"][0])
		fileName := r.Form["fileName"][0]
		targetFile := findFileInAll(fileName)
		fmt.Println(targetFile.FileFullName)
		targetFile.GetFile("/temp/")
		fmt.Println(targetFile.FileFullName + " Finished")
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[DELETE] " + r.URL.Path)
}

func fileHandle(source string) {
	//SysConfig & File pkg test
	fmt.Println("Start FileHandler")
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
