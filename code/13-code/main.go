package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Password string `json:"-"`
	age      int
}

type UpdateUserReq struct {
	Name *string `json:"name,omitempty"`
	Age  *int    `json:"age,omitempty"`
}

type Event struct {
	Name string    `json:"name"`
	At   time.Time `json:"at"`
}

func main() {
	u := User{ID: 1, Name: "张三", Password: "secret", age: 18}
	fmt.Println(u)
	b, err := json.Marshal(u)
	fmt.Println(b)
	fmt.Println(string(b), err)

	data := []byte(`{"id":2,"name":"李四","unknown":true}`)

	var u1 User
	err1 := json.Unmarshal(data, &u1)
	fmt.Println(u1)
	fmt.Println(u1.ID, u1.Name, u1.Email, err1)

	name, age := "王五", 0
	fmt.Println(&name, &age)
	b1, _ := json.Marshal(UpdateUserReq{Name: &name, Age: &age})
	fmt.Println(string(b1))

	var v any
	_ = json.Unmarshal([]byte(`{"age":18, "score":95.5}`), &v)
	fmt.Println(v)
	m := v.(map[string]any)
	fmt.Printf("%T %v\n", m["age"], m["age"])

	e := Event{
		Name: "上线了",
		At:   time.Date(2026, 10, 18, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
	}
	b2, _ := json.Marshal(e)
	fmt.Println(string(b2))

	const stream = `{"id":1, "name":"A"}`

	const stream2 = `{"id":1,"name":"A"}
	{"id":2,"name":"B"}
	`
	fmt.Println("strings.NewReader(stream)：" , strings.NewReader(stream))
	fmt.Println("json.NewDecoder(strings.NewReader(stream))：" , json.NewDecoder(strings.NewReader(stream)))

	dec := json.NewDecoder(strings.NewReader(stream))
	for {
		var u User
		err := dec.Decode(&u)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(u.ID, u.Name)
	}
	f, err := os.Open("data.txt")
}
