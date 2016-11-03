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
)

func main() {
	var fault, row int

	//Init command line args
	flag.IntVar(&fault, "fault", 2, "系统容错数量")
	flag.IntVar(&row, "row", 2, "文件阵列行数")
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
			testFileHandle("/Users/vaaaas/Desktop/READING/PDF.pdf")
			testFileHandle("/Users/vaaaas/Desktop/READING/MP3.mp3")
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
	fmt.Println("[UPLOAD] " + r.URL.Path)
	fmt.Println(r.Form["uploadInput"][0])
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[DOWNLOAD] " + r.URL.Path)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[DELETE] " + r.URL.Path)
}

func testFileHandle(source string) {
	//SysConfig & File pkg test
	fmt.Println("Start test FileHandler")
	var file01 File.File
	file01.Init(source)
	name, ext := file01.SliceFileName()
	fmt.Println(name + " , " + ext)
	file01.InitDataFiles()
	file01.InitRddtFiles()
	file01.GetFile("/Users/vaaaas/Desktop/WRITING/")
}
