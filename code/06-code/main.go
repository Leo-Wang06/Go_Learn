package main

import (
	"fmt"
)

type User struct {
	ID   int
	Name string
	Age  int
}

func printValeu(p *int) {
	if p == nil {
		fmt.Println("空指针")
		return
	}
	fmt.Println(&p)
}

func double(n *int) {
	*n = *n * 2
}

func birthday(u *User) {
	u.Age++
}

func newInt(v int) *int {
	x := v
	return &x
}

func findUser(id int) *User {
	if id <= 0 {
		return nil
	}
	return &User{ID: id, Name: "张三"}
}

func (u *User) SetName(name string) {
	u.Name = name
}

type Node struct {
	Value int
	Next *Node
}

func main() {
	// x := 42
	// // p := &x
	// var q *int

	// fmt.Println(x)
	// fmt.Println(&x)
	// fmt.Println(p)
	// fmt.Println(*p)

	// *p = 100
	// fmt.Println(x)

	// q = &x
	// fmt.Println(*q)

	// var p *User
	// fmt.Printf("%T", p)
	// printValeu(q)
	// printValeu(&x)

	// double(&x)
	// fmt.Println(x)

	// u := User{ID: 1,  Name: "小明",Age: 18}

	// birthday(&u)

	// fmt.Println(u.Age)

	// p := newInt(10)
	// fmt.Println(p)
	// fmt.Println(*p)

	// u := findUser(0)
	// if u == nil {
	// 	fmt.Println("用户不存在")
	// } else {
	// 	fmt.Println(u.Name)
	// }

	// p1 := new(User)
	// p2 := &User{}
	// p3 := &User{ID: 5, Name: "小李", Age: 20}

	// fmt.Println(p1, p2, p3)
	// fmt.Println(*p2)
	// fmt.Println(p1 == p2)

	// x := 25

	// q := &x

	// fmt.Println(q)

	// u := User{ID: 6, Name: "小杨", Age: 30}

	// u.SetName("小林")
	// fmt.Println(u)
	// fmt.Println(u.Name)

	head := &Node{Value: 1}
	head.Next = &Node{Value: 2}

	fmt.Println(head.Value)
	fmt.Println(head.Next.Value)
	fmt.Println(head.Next.Next)
}
