package main

import "fmt"

type User struct {
	ID   int
	Name string
	Age  int
}

func (u User) Describe() string {
	return fmt.Sprintf("%+v", u)
}

type Student struct {
	User
	Like string
}

func main() {
	u1 := User{ID: 1, Name: "小明", Age: 20}
	var u2 User
	u2 = User{ID: 2, Name: "小红", Age: 25}
	var u3 User = User{ID: 3, Name: "小兰", Age: 30}
	fmt.Printf("%+v,%+v,%+v",u1, u2, u3)
	fmt.Println(u1.Describe())
}