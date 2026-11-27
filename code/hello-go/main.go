package main

import (
	"fmt"
	"hello-go/mathutil"
)

func main() {
	sum := mathutil.Add(3, 5)
	fmt.Println("3 + 5 =", sum)
	fmt.Println(greet())
	// fmt.Println(mathutil.subtract(3, 5))

}
