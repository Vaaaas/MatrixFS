package File

import (
	"strings"
	"os"
	"Vaaaas/MatrixFS/SysConfig"
	"github.com/golang/glog"
	"errors"
	"io"
	"strconv"
	"math"
	"fmt"
)

var AllFiles = []File{}

//使用时需要通过File(包名).File(结构体名)来访问
type File struct {
	FileFullName string
	Size         int64
	FillLast     bool
	FillSize     int64
	SliceSize    int64
}

func (file *File) Init(source string) error {
	fileInfo, err := os.Stat(source)
	if err != nil {
		glog.Error(err)
		panic(err)
	}
	if fileInfo.IsDir() {
		glog.Error("初始化失败：该路径指向的是文件夹")
		return errors.New("初始化失败：该路径指向的是文件夹")
	} else {
		file.FileFullName = fileInfo.Name()
		file.Size = fileInfo.Size()
		if ((file.Size % (int64)(SysConfig.SysConfig.SliceNum)) != 0) {
			file.FillLast = true
			file.SliceSize = file.Size / (int64)(SysConfig.SysConfig.SliceNum) + 1
			file.FillSize = file.SliceSize * (int64)(SysConfig.SysConfig.SliceNum) - file.Size
		} else {
			file.FillLast = false
			file.SliceSize = file.Size / (int64)(SysConfig.SysConfig.SliceNum)
			file.FillSize = 0
		}
		glog.Infof("%+v ", file)
		//todo : 在内存中的具体行为待观察
		AllFiles = append(AllFiles, *file)

		fmt.Printf("All Files: %+v  ,len: %d\n", AllFiles[len(AllFiles) - 1], len(AllFiles))

		return nil
	}
}

//func (file File) cpyFile(source string) (error) {
//	in, err := os.Open(source)
//	if err != nil {
//		return err
//	}
//	defer in.Close()
//	out, err := os.Create("./temp/" + file.FileFullName)
//	if err != nil {
//		glog.Error(err)
//		panic(err)
//	}
//	defer func() {
//		cerr := out.Close()
//		if err == nil {
//			err = cerr
//		}
//	}()
//	if _, err = io.Copy(out, in); err != nil {
//		glog.Error(err)
//		panic(err)
//	}
//	err = out.Sync()
//	return nil
//}

func (file File) SliceFileName() (string, string) {
	if file.FileFullName != "" {
		slices := strings.Split(file.FileFullName, ".")
		if len(slices) > 1 {
			n := len(".") * (len(slices) - 1)
			for i := 0; i < len(slices) - 1; i++ {
				n += len(slices[i])
			}
			name := make([]byte, n - 1)
			bp := copy(name, slices[0])
			for _, s := range slices[1:len(slices) - 1] {
				bp += copy(name[bp:], ".")
				bp += copy(name[bp:], s)
			}
			exten := slices[len(slices) - 1]
			return string(name), exten
		} else {
			glog.Error("分割文件名失败：该文件不含扩展名")
			return "", ""
		}
	} else {
		glog.Error("分割文件名失败：未定义文件名")
		return "", ""
	}
}

func (file File) InitDataFiles() error {
	source := "./temp/" + file.FileFullName

	sourceFile, err := os.Open(source)
	if err != nil {
		glog.Error("分割时打开源文件失败")
		panic(err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			panic(err)
		}
	}()

	for i := 0; i < SysConfig.SysConfig.DataNum; i++ {
		for j := 0; j < SysConfig.SysConfig.RowNum; j++ {
			if file.initOneDataFile(i, j, sourceFile) != nil {
				glog.Errorf("生成单个数据文件文件失败 i=%d j=%d", i, j)
				panic(err)
			}
		}
	}
	return nil
}

func (file File) initOneDataFile(col int, row int, sourceFile *os.File) error {
	outFile, err := os.Create("./temp/Data." + strconv.Itoa(col) + "/" + file.FileFullName + "." + strconv.Itoa(col) + strconv.Itoa(row))
	if err != nil {
		glog.Error("新建数据分块文件失败 " + "./temp/Data." + strconv.Itoa(col) + "/" + file.FileFullName + "." + strconv.Itoa(col) + strconv.Itoa(row))
		panic(err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			panic(err)
		}
	}()

	buffer := make([]byte, file.SliceSize)
	n, err := sourceFile.Read(buffer)
	if err != nil && err != io.EOF {
		glog.Errorf("buffer读取文件失败 col=%d row=%d", col, row)
		panic(err)
	}

	if file.FillLast && (int64)(n) != file.SliceSize {
		if (int64)(n) == 0 {
			tempBuffer := make([]byte, file.SliceSize)
			for j := 0; (int64)(j) < file.FillSize; j++ {
				tempBuffer[j] = (byte)(0)
				if _, err := outFile.Write(buffer[:file.SliceSize]); err != nil {
					glog.Errorf("写入数据分块失败 col=%d row=%d", col, row)
					panic(err)
				}
				return nil
			}
		} else {
			for j := n; (int64)(j) < file.SliceSize; j++ {
				buffer[j] = (byte)(0)
			}
		}
	}

	if _, err := outFile.Write(buffer[:file.SliceSize]); err != nil {
		glog.Errorf("写入数据分块失败 col=%d row=%d", col, row)
		panic(err)
	}
	return nil
}

func (file File) GetFile(targetFolder string) error {
	target, err := os.Create(targetFolder + file.FileFullName)
	var realSliceNum int64
	if file.Size % file.SliceSize != 0 {
		realSliceNum = file.Size / file.SliceSize + 1
	} else {
		realSliceNum = file.Size / file.SliceSize
	}
	if err != nil {
		glog.Error("无法新建目标文件")
		panic(err)
	}
	defer func() {
		if err := target.Close(); err != nil {
			panic(err)
		}
	}()

	buffer := make([]byte, file.SliceSize)
	for i := 0; (int64)(i) < realSliceNum; i++ {
		dataPosition := i / SysConfig.SysConfig.RowNum
		rowPosition := i % SysConfig.SysConfig.RowNum

		dataFile, err := os.Open("./temp/Data." + strconv.Itoa(dataPosition) + "/" + file.FileFullName + "." + strconv.Itoa(dataPosition) + strconv.Itoa(rowPosition))
		if err != nil {
			glog.Error("读取数据分块文件失败 " + "./temp/Data." + strconv.Itoa(dataPosition) + "/" + file.FileFullName + "." + strconv.Itoa(dataPosition) + strconv.Itoa(rowPosition))
			panic(err)
		}
		defer func() {
			if err := dataFile.Close(); err != nil {
				panic(err)
			}
		}()
		n, err := dataFile.Read(buffer)
		if err != nil && err != io.EOF {
			glog.Error("buffer读取文件失败 " + strconv.Itoa(i))
			panic(err)
		}

		if (file.FillLast&&(int64)(i) == realSliceNum - 1) {
			if (int64)(file.FillSize) != file.SliceSize {
				if _, err := target.Write(buffer[:((int64)(n) - file.FillSize % file.SliceSize)]); err != nil {
					glog.Error("写入数据分块失败 " + strconv.Itoa(i))
					panic(err)
				}
			}
		} else {
			if _, err := target.Write(buffer[:n]); err != nil {
				glog.Error("写入数据分块失败 " + strconv.Itoa(i))
				panic(err)
			}
		}
	}

	return nil
}

func (file File)InitRddtFiles() error {
	rddtFolderCounter := 0
	rddtRowCounter := 0
	for faultCount := 0; faultCount < SysConfig.SysConfig.FaultNum; faultCount++ {
		k := (int)((faultCount + 2) / 2 * (int)(math.Pow(-1, (float64)(faultCount + 2))))
		for i := 0; i < SysConfig.SysConfig.DataNum; i++ {
			if (rddtRowCounter % SysConfig.SysConfig.RowNum == 0&&(i != 0 || faultCount != 0)) {
				rddtRowCounter = 1;
				rddtFolderCounter++;
			} else {
				rddtRowCounter++;
			}
			file.initOneRddtFile(i, k, rddtFolderCounter)
		}
	}
	return nil
}

func (file File)initOneRddtFile(startFolderNum, k, rddtNum int) error {
	rddtFilePath, err := os.Create("./temp/Rddt." + strconv.Itoa(rddtNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startFolderNum))
	if err != nil {
		glog.Error("新建冗余分块文件失败 " + "./temp/Rddt." + strconv.Itoa(rddtNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startFolderNum))
		panic(err)
	}
	defer func() {
		if err := rddtFilePath.Close(); err != nil {
			panic(err)
		}
	}()

	buffer := make([]byte, file.SliceSize)

	for i := 0; i < SysConfig.SysConfig.RowNum; i++ {
		folderPosi := startFolderNum + k * i
		if folderPosi >= SysConfig.SysConfig.DataNum {
			folderPosi = folderPosi - SysConfig.SysConfig.DataNum
		} else if folderPosi < 0 {
			folderPosi = SysConfig.SysConfig.DataNum + folderPosi
		}

		sourceFile, err := os.Open("./temp/Data." + strconv.Itoa(folderPosi) + "/" + file.FileFullName + "." + strconv.Itoa(folderPosi) + strconv.Itoa(i))
		if err != nil {
			glog.Error("生成冗余文件 " + "./temp/Rddt." + strconv.Itoa(rddtNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startFolderNum) + " 时打开源文件失败 " + "./temp/Data." + strconv.Itoa(folderPosi) + "/" + file.FileFullName + "." + strconv.Itoa(folderPosi) + strconv.Itoa(i))
			panic(err)
		}
		defer func() {
			if err := sourceFile.Close(); err != nil {
				panic(err)
			}
		}()
		if (i == 0) {
			_, err = sourceFile.Read(buffer)
			if err != nil && err != io.EOF {
				glog.Error("buffer读取文件失败 " + strconv.Itoa(i))
				panic(err)
			}
		} else {
			tempBytes := make([]byte, file.SliceSize)
			_, err = sourceFile.Read(tempBytes)
			if err != nil && err != io.EOF {
				glog.Error("tempBuffer读取文件失败" + strconv.Itoa(i))
				panic(err)
			}
			for byteCounter := 0; byteCounter < len(buffer); byteCounter++ {
				buffer[byteCounter] = buffer[byteCounter] ^ tempBytes[byteCounter]
			}
		}
		if i == SysConfig.SysConfig.RowNum - 1 {
			if _, err := rddtFilePath.Write(buffer[:file.SliceSize]); err != nil {
				glog.Error("写入冗余分块文件失败 " + "./temp/Rddt." + strconv.Itoa(rddtNum) + "/" + file.FileFullName + "." + strconv.Itoa(k) + strconv.Itoa(startFolderNum))
				panic(err)
			}
		}
	}
	return nil
}

//todo : Delete all temp files when finished the whole upload operation