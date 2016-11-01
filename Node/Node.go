package Node

var AllNodes []Node
var LostNodes []int

type Node struct {
	ID      int
	Address string
	port    string
	Volume  int64
	//Status:
	//0 -> 待命
	//1 -> 计算中
	//2 -> 丢失
	Status  int8
}

func InitNodes() {

}