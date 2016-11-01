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

	//初始化命令行参数
	flag.IntVar(&fault, "fault", 2, "系统容错数量")
	flag.IntVar(&row, "row", 2, "文件阵列行数")
	//初始化命令行参数
	//调试时log_dir="./log"
	flag.Parse()
	//退出时调用，确保日志写入文件
	defer glog.Flush()
	//初始化系统配置
	//SysConfig.InitConfig(fault, row)
	glog.Info("Server start here")

	//testFileHandle()

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/index", indexPageHandler)
	http.HandleFunc("/file", filePageHandler)
	http.HandleFunc("/node", nodePageHandler)

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
	if r.URL.Path == "/" {
		if !sysConfigured() {
			fmt.Println("not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
		} else {
			fmt.Println("configured, redirect to file.html")
			http.Redirect(w, r, "/file", http.StatusFound)
		}
	} else {
		t, err := template.ParseFiles("view/404.html")
		if (err != nil) {
			log.Println(err)
		}
		t.Execute(w, nil)
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
	fmt.Println("mathod: ", r.Method)
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
			SysConfig.InitConfig(faultNum, rowNum)
			testFileHandle()
		}
	}

	fmt.Println(len(File.AllFiles))

	t, err := template.ParseFiles("view/file.html")
	if (err != nil) {
		fmt.Println(err)
	}
	t.Execute(w, nil)
}

func nodePageHandler(w http.ResponseWriter, r *http.Request) {

}

func testFileHandle() {
	//SysConfig & File 包测试
	var file01 File.File
	file01.Init("/Users/vaaaas/Desktop/READING/PDF.pdf")
	name, ext := file01.SliceFileName()
	fmt.Println(name + " , " + ext)
	file01.InitDataFiles()
	file01.InitRddtFiles()
	file01.GetFile("/Users/vaaaas/Desktop/WRITING/")
}
