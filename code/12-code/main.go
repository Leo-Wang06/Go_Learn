package main

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type User struct {
	Name string
	Age   int
	Tags []string
}

type ByAge []User

func(a ByAge) Len() int {
	return  len(a)
}

func (a ByAge) Less(i, j int) bool {
	return  a[i].Age < a[j].Age
}

func (a ByAge) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func main() {
	u := User{"张三", 1, []string{"go", "web"}}

	fmt.Printf("%v\n", u)
	fmt.Printf("%+v\n", u)
	fmt.Printf("%#v\n", u)
	fmt.Printf("%T\n", u)

	fmt.Printf("%05d|%.2f|%8s|%-8s|\n", 42, 3.14159, "go", "go")

	base := errors.New("db down")
	wrapped := fmt.Errorf("query user: %w", base)

	fmt.Println(wrapped)
	fmt.Println(errors.Is(wrapped, base))

	s := " Hello, Go World "

	fmt.Println(strings.Contains(s, "Go"))
	fmt.Println(strings.HasPrefix(strings.TrimSpace(s), "Hello"))
	fmt.Println(strings.HasSuffix(strings.TrimSpace(s), "World"))
	fmt.Println(strings.EqualFold("Go", "GO"))

	parts := strings.Split("go,ts,js", ",")
	fmt.Println(parts)
	fmt.Println(strings.Join(parts, " | "))

	before, after, found := strings.Cut("key=value", "=")
	fmt.Println(before, after, found)

	var b strings.Builder
	for i := 1; i <= 3; i++ {
		fmt.Fprintf(&b, "%d", i)
	}
	fmt.Println(b.String())
	fmt.Println(b.Len())

	n, err := strconv.Atoi("42")
	fmt.Println(n, err)

	_, err = strconv.Atoi("4.2")
	fmt.Println(err)

	s1 := strconv.Itoa(42)
	fmt.Println(s1)

	i64, err := strconv.ParseInt("-123", 10, 64)
	fmt.Println(i64, err)

	fmt.Println(strconv.FormatFloat(3.1415926, 'f', 2, 64))

	var buf bytes.Buffer
	buf.WriteString("hello")
	buf.WriteByte(' ')
	buf.WriteString("go")

	fmt.Println(buf.String())
	fmt.Println(buf.Len())

	fmt.Fprintf(&buf, " %d", 2026)
	fmt.Println(buf.String())

	buf.Reset()
	fmt.Println(buf.Len())

	users := []User{{Name:"张三",Age: 28}, {Name: "李四",Age: 22}, {Name: "王五",Age: 35}}
	sort.Slice(users, func (i, j int) bool {
		return users[i].Age < users[j].Age
	})
	fmt.Println(users)

	sort.Sort(ByAge(users))
	fmt.Println(users)

	nums := []int{3, 2, 1}
	slices.Sort(nums)
	fmt.Println(nums)

	slices.SortFunc(users, func(a, b User) int{
		return cmp.Compare(a.Age, b.Age)
	})
	fmt.Println(users)

	dup := []int{1, 1, 2, 2, 3, 2}
	dup = slices.Compact(dup)
	fmt.Println(dup)
}