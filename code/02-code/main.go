package main

import "fmt"

type OrderStatus int

const (
	OrderStatusPending OrderStatus = iota // 0
    OrderStatusPaid                       // 1
    OrderStatusShipped                    // 2
    OrderStatusCompleted                  // 3
    OrderStatusCancelled                  // 4
)

type UserID int

func (id UserID) String() string {
	return  fmt.Sprint("user-%d", int(id))
}

func main() {
	var name string = "张三"

	var age = 25

	score := 95.45

	var city string
	city = "北京"

	fmt.Printf("%T %T %T %T\n", name, age, score, city)

	_, err := fmt.Println("hello")
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range []int{1, 2, 3} {
    	fmt.Println(v)
	}

	fmt.Println(KB, MB, GB)

	var s OrderStatus = OrderStatusPaid
	fmt.Println(s)
	fmt.Printf("%T\n", s)

	// var m map[string]int

	// fmt.Println(m["a"])
	// m["a"] = 1

	m := make(map[string]int)

	m["a"] = 1
	fmt.Println(m["a"])

	type UserID int
	type OrderID int

	var uid UserID = 1
	var oid OrderID = 2

	// uid = oid
	uid = UserID(oid)

	fmt.Println(uid)

	var id UserID = 42

	fmt.Println(id)

	Practice()
}
