package main

import (
	"errors"
	"fmt"
)

type MyError struct {
	Msg string
}

func (e MyError) Error() string {
	return e.Msg
}

var ErrUserNotFound = errors.New("user not found")

func findUserName(id int) (string, error) {
	if id != 1 {
		return "", ErrUserNotFound
	}
	return "张三", nil
}

func main() {
	var err error = MyError{Msg: "something went wrong"}
	fmt.Println(err)
	fmt.Printf("%T\n", err)

	name, err := findUserName(2)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			fmt.Println("用户不存在", err)
			// return
		}
		fmt.Println("其他错误", err)
		// return
	}
	fmt.Println(name)

	base := errors.New("not found")
	wrapped := fmt.Errorf("find user: %w", base)

	fmt.Println("wrapped:",wrapped)
	fmt.Println(wrapped == base)
	fmt.Println(errors.Is(wrapped, base))
}
