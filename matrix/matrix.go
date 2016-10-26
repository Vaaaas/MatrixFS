package main

import (
	"flag"
	"fmt"
)

func main() {
	//args := flag.Args()
	//if len(args) < 1 {
	//	fmt.Fprintf(os.Stderr, "No args\n")
	//	//usage()
	//	os.Exit(2)
	//}
	//
	//if args[0] == "help" {
	//	fmt.Fprintf(os.Stderr, "Help command\n")
	//}

	//args_num := len(os.Args)
	//if args_num != 0 {
	//	fmt.Fprintf(os.Stderr, "Args !\n")
	//} else {
	//	fmt.Fprintf(os.Stderr, "No args !\n")
	//}

	//arg_num := len(os.Args)
	//fmt.Printf("the num of input is %d\n",arg_num)
	//fmt.Printf("they are :\n")
	//for i := 0 ; i < arg_num ;i++{
	//	fmt.Println(os.Args[i])
	//}
	//sum := 0
	//for i := 1 ; i < arg_num; i++{
	//	curr,err := strconv.Atoi(os.Args[i])
	//	if(err != nil){
	//		fmt.Println("error happened ,exit")
	//	}
	//	sum += curr
	//}
	//fmt.Printf("sum of Args is %d\n",sum)

	flag.Parse()
	//args := flag.Args()
	args_num := len(flag.Args())
	if args_num != 0 {
		fmt.Printf("sum of Args is %d\n", args_num)
		fmt.Printf("Arg is : %s\n", flag.Args()[0])
	}else{
		fmt.Printf("sum of Args is %d\n", args_num)
	}
}
