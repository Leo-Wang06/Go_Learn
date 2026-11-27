package main

import (
	"fmt"
)

func double(n *int) {
	fmt.Printf(" 函数内修改前：*n = %d\n", *n)
	*n = *n * 2
	fmt.Printf(" 函数内修改后：*n = %d\n", *n)
}

func doubleValue(n int) {
	n = n * 2
}

func addOne(n *int) {
	*n++
}

type User struct {
	ID int
	Name string
	Age *int
}

func (u *User) Setname(name string) {
	u.Name = name
}

func (u *User) Greet() string {
	return "你好，" + u.Name
}

func safeGreet(u *User) string {
	if u == nil {
		return "用户不存在"
	}
	return "你好，" + u.Name
}

func safeCall(u *User) string {
	if u == nil {
		return "(nil 用户，跳过)"
	}
	return u.Greet()
}

func smallUse(x int) *int {
	x ++
	return  &x
}

func main() {
	// x := 10

	// double(&x)
	// fmt.Println(x)
	// doubleValue(x)
	// fmt.Println(x)
	// addOne(&x)
	// fmt.Println(x)

	// u := User{}
	// fmt.Println(u.Name)
	// u.Setname("王五")
	// fmt.Println(u.Name)

	var p *User
	fmt.Println("打印 nil 指针:", p)
	fmt.Println("判空结果:", p == nil)

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("捕获到 panic", r)
			}
		}()
		fmt.Println(p.Name)
	}()

	var q *User
	fmt.Println(safeGreet(q))

	q = &User{ID: 1, Name: "小明"}
	fmt.Println(safeGreet(q))

	fmt.Println("方法调用：", safeCall(q))

	var r *User
	fmt.Println("方法调用(nil):", safeCall(r))

	x := 0

	u1 := new(User)
	u1 = &User{ID: 1, Name: "张三", Age: nil}
	u2 := &User{ID: 2, Name: "李四", Age: &x}

	fmt.Println(*u1, *u2)

	fmt.Println(*smallUse(x))

	m := make(map[string]int)
	m["key"] = 1
	m1 := m["key"]
	fmt.Println(&m1)
}
