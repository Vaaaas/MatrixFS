package main

import (
	"github.com/golang/glog"
	"github.com/Vaaaas/MatrixFS/NodeStruct"
	"github.com/Vaaaas/MatrixFS/SysConfig"
	"github.com/Vaaaas/MatrixFS/File"
	"flag"
	"net/http"
	"os"
	"encoding/json"
	"fmt"
	"strconv"
	"html/template"
	"io"
	"io/ioutil"
	"bytes"
	"mime/multipart"
	"math"
)

var AllNodes = make(map[uint]NodeStruct.Node)
var DataNodes []uint
var RddtNodes []uint
var LostNodes []uint
var EmptyNodes []uint

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
	NodeStruct.IDCounter = 0

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

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {

		if !SysConfig.SysConfigured() {
			glog.Infoln("URL: " + r.URL.Path + " not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
		} else {
			if !NodeConfigured() {
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
	if SysConfig.SysConfigured() {
		if !NodeConfigured() {
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
		if !SysConfig.SysConfigured() {
			glog.Infoln("URL: " + r.URL.Path + "not configured, redirect to index.html")
			http.Redirect(w, r, "/index", http.StatusFound)
			return
		}
	} else {
		if !SysConfig.SysConfigured() {
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

	glog.Infof("Length of AllNodes : %d", len(AllNodes))
	t, err := template.ParseFiles("view/node.html")
	if (err != nil) {
		glog.Errorln(err)
	}
	t.Execute(w, AllNodes)
}

func filePageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	glog.Infoln("[File Page] method: ", r.Method)
	if !SysConfig.SysConfigured() {
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

	if !SysConfig.SysConfigured() {
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

	if !SysConfig.SysConfigured() {
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

	// todo : send slices to Nodes
	SendToNode(file01)
}

func greetHandler(w http.ResponseWriter, r *http.Request) {
	//在Master中建立空Node变量
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
	NodeStruct.IDCounter++
	node.ID = NodeStruct.IDCounter
	fmt.Printf("Hello %d\n", node.ID)

	//先将新节点加入空节点slice中
	EmptyNodes = append(EmptyNodes, node.ID)

	//先将新节点加入空节点slice中
	AllNodes[node.ID] = node

	//根据情况，将新节点加入data／rddt slice中
	if !appendNode(&node) {
		glog.Infof("Still in Empty Slice %+v", node)
	} else {
		glog.Infof("%+v", node)
	}
	w.Header().Set("ID", strconv.Itoa((int)(node.ID)))
	w.WriteHeader(http.StatusOK)
}

func NodeConfigured() bool {
	//todo : auto configure node

	return len(AllNodes) != 0
}

func findFileInAll(name string) *File.File {
	for _, tempFile := range File.AllFiles {
		if tempFile.FileFullName == name {
			return &tempFile
		}
	}
	return nil
}

func checkDataNodeNum() int {
	return SysConfig.SysConfig.DataNum - len(DataNodes)
}

func checkRddtNodeNum() int {
	return SysConfig.SysConfig.RddtNum - len(RddtNodes)
}

func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func appendNode(node *NodeStruct.Node) bool {
	if checkDataNodeNum() > 0 {
		emptyToData(node)
		index := SliceIndex(len(EmptyNodes), func(i int) bool {
			return EmptyNodes[i] == node.ID
		})
		EmptyNodes = append(EmptyNodes[:index], EmptyNodes[index + 1:]...)
		return true;
	} else if checkRddtNodeNum() > 0 {
		emptyToRddt(node)
		index := SliceIndex(len(EmptyNodes), func(i int) bool {
			return EmptyNodes[i] == node.ID
		})
		EmptyNodes = append(EmptyNodes[:index], EmptyNodes[index + 1:]...)
		return true;
	} else {
		return false;
	}
}

func emptyToData(node *NodeStruct.Node) {
	DataNodes = append(DataNodes, node.ID)
}

func emptyToRddt(node *NodeStruct.Node) {
	RddtNodes = append(RddtNodes, node.ID)
}

func SendToNode(file File.File) {
	for i := 0; i < SysConfig.SysConfig.DataNum; i++ {
		for j := 0; j < SysConfig.SysConfig.RowNum; j++ {
			postOneFile(file, true, DataNodes[i], i, j, 0)
		}
	}

	nodeCounter := 0
	fileCounter := 0
	rddtFileCounter := 0
	for xx := 0; xx < SysConfig.SysConfig.FaultNum; xx++ {
		k := (int)((xx + 2) / 2 * (int)(math.Pow(-1, (float64)(xx + 2))))
		for fileCounter < SysConfig.SysConfig.DataNum {
			glog.Infof("Rddt Node Num : %d \t k : %d \t fileCounter : %d \t nodeCounter : %d\n", nodeCounter, k, fileCounter, nodeCounter)
			postOneFile(file, false, RddtNodes[nodeCounter], k, fileCounter, nodeCounter)
			fileCounter++;
			rddtFileCounter++;
			if (rddtFileCounter % (SysConfig.SysConfig.SliceNum / SysConfig.SysConfig.DataNum) == 0) {
				nodeCounter++;
				rddtFileCounter = 0;
			}
			if (fileCounter != SysConfig.SysConfig.DataNum) {
				continue;
			}
			fileCounter = 0;
			break;
		}
	}
}

func postOneFile(file File.File, isData bool, nodeID uint, posiX, posiY, nodeCounter int) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	var filePath string
	if isData {
		filePath = "./temp/DATA." + strconv.Itoa(posiX) + "/" + file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	} else {
		filePath = "./temp/RDDT." + strconv.Itoa(nodeCounter) + "/" + file.FileFullName + "." + strconv.Itoa((int)(posiX)) + strconv.Itoa(posiY)
	}

	//关键的一步操作
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filePath)
	if err != nil {
		glog.Errorf("error writing to buffer + %s", err)
		panic(err)
	}

	//打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		glog.Errorln("error opening file + %s", err)
		panic(err)
	}

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	url := "http://" + AllNodes[nodeID].Address.String() + ":" + strconv.Itoa(AllNodes[nodeID].Port) + "/upload"
	resp, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorln(err)
		panic(err)
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
}