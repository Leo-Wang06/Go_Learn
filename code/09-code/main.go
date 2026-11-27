package main

import "fmt"

var x = initX()

func initX() int {
	fmt.Println("1.变量初始化")
	return  10
}

var y = initY()

func initY() int {
	fmt.Println("1.变量初始化 Y")
	return 20
}

func init() {
	fmt.Println("2.init")
}

func main() {
	fmt.Println("3.maim",x, y)
}
