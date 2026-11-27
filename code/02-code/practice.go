package main

import (
	"fmt"
	"strconv"
)

const (
	_ = iota
	KB = 1 << (10 * iota)
	MB
	GB
)

func Practice() {
	var name string = "张三"
	var age = 25
	score := 99

	var city string
	city = "北京"

	fmt.Printf("%T %T %T %T\n", name, age, score, city)

	var a, b = 10, 20
	a,b = b, a
	fmt.Println(a, b)

	fmt.Println(KB, MB, GB)

	var num int64
	var famliy string
	var can bool
	var num2 *int
	var num3 []int
	var famliy2 map[string]int
	
	fmt.Println(num,famliy,can,num2,num3,famliy2)

	var score2 float64 = 99.80

	num4 := int(num)
	num5 := int(score2)

	var name2 = strconv.Itoa(age)
	age2, err := strconv.Atoi("123")
	
	if err != nil {
		return
	} 

	fmt.Println(num4, num5, name2, age2)
	
}