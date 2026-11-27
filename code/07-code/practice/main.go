package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
)

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Circle struct {
	R float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.R * c.R
}

func (c Circle) Perimeter() float64 {
	return  2 * math.Pi * c.R
}

type Rect struct {
	W, H float64
}

// func (r Rect) Area() float64 {
// 	return r.W * r.H
// }

func (r Rect) Perimeter() float64 {
	return 2 * ( r.W + r.H )
}

type Animal interface {
	speak() string
}

type Dog struct {
	Name string
}

func (d Dog) speak() string {
	return d.Name + "汪汪"
}

type Cat struct {
	Name string
}

func (c Cat) speak() string {
	return  c.Name + "喵喵"
}

func printType(v any) {
	fmt.Printf("type=%T valur=%v\n", v, v)
}

func descirbe(v any) {
	switch x := v.(type) {
	case nil:
		fmt.Println("nil 接口")
	case int:
		fmt.Printf("int: %d\n", x)
	case string:
		fmt.Printf("string: %q\n", x)
	case bool:
		fmt.Printf("bool: %t\n", x)
	default:
		fmt.Printf("其他类型:%v", x)
	}
}

type Money struct {
	Amount float64
}

func (m Money) String() string {
	return fmt.Sprintf("CNY %.2f", m.Amount)
}

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type ReadWriter interface {
	Reader
	Writer
}

func main() {
	// c := Circle{R: 5}
	// r := Rect{W: 4, H: 5}

	// fmt.Println(c.Area(), c.Perimeter(), r.Area(), r.Perimeter())

	animal := []Animal{
		Dog{"大黄"},
		Cat{"小白"},
		Dog{"小黑"},
	}

	for _, a := range animal {
		fmt.Println(a.speak())
	}

	x := 15
	y := "小红"
	n := Dog{"大红"}
	t := true
	printType(x)
	printType(y)
	printType(n)

	values := []any{1, "a", true, 3.14}

	var strings []string

	for _, v := range values {
		if s, ok := v.(string); ok {
			strings = append(strings, s)
		}
	}

	fmt.Println(strings)

	descirbe(x)
	descirbe(y)
	descirbe(n)
	descirbe(nil)
	descirbe(t)

	m := Money{123.45}
	fmt.Println(m)

	var _ Shape = Circle{}
	// var _ Shape = Rect{}

	var _ io.ReadWriter = (*bytes.Buffer)(nil)

}
