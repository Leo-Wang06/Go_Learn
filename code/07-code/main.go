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

type Circle struct{ R float64 }

func (c Circle) Area() float64 {
	return math.Pi * c.R * c.R
}

func (c Circle) Perimeter() float64 {
	return  2 * math.Pi * c.R
}

type Rect struct{ W, H float64}
func (r Rect) Area() float64 {
	return r.W * r.H
}
func (r Rect) Perimeter() float64 {
	return 2 * (r.W + r.H)
}

func PrintShape(s Shape) {
	fmt.Printf("%T: area=%.2f perimeter=%.2f\n", s, s.Area(), s.Perimeter())
}

func printAny(v any) {
	fmt.Printf("type=%T valur=%v\n", v, v)
}

type Dog struct {
	Name string
}

type Speaker interface {
	speak() string
}

func (d Dog) speak() string {
	return  d.Name + "汪"
}

type Point struct {
	X, Y int
}

func (p Point) string() string {
	return  fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

func descirbe(v any) {
	switch x := v.(type) {
	case nil:
		fmt.Println("nil 接口")
	case int:
		fmt.Printf("int: %d\n", x)
	case string:
		fmt.Printf("string: %q\n", x)
	case fmt.Stringer:
		fmt.Printf("Stringer: %s\n", x)
	default:
		fmt.Printf("其他类型：%T\n", x)
	}
}

type UserFinder interface {
	FindByID(ID int) string
}

type UserService struct{ repo UserFinder }

func NewUserService(repo UserFinder) *UserService {
	return &UserService{repo: repo}
}

func (u UserService) FindByID(ID int) string {
	return u.repo.FindByID(ID)
}

type Stringer interface {
	String() string
}

type UpperWriter struct {
	W io.Writer
}

func (u UpperWriter) Write(p []byte) (int, error) {
	return u.W.Write(bytes.ToUpper(p))
}

func main() {

	// PrintShape(Circle{R: 10})
	// PrintShape(Rect{10, 10})

	// printAny(42)
	// printAny("hello")
	// printAny(nil)

	// var x any = 42
	// // x + 1
	// n, ok := x.(int)
	// if !ok {
	// 	panic("not an int")
	// }
	// fmt.Println(n + 1)

	// var x any = "hello"
	// s, ok := x.(string)
	// fmt.Println(s, ok)

	// n, ok := x.(int)
	// fmt.Println(n, ok)

	// var x any = Dog{Name: "旺财"}
	// s, ok := x.(Speaker)
	// fmt.Println(ok, s.speak())

	descirbe(nil)
	descirbe(42)
	descirbe("go")
	descirbe(Point{1, 2})
	descirbe(3.14)

}
