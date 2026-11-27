package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// 核心：这个函数只认 io.Reader，不关心数据到底来自哪
func printAll(r io.Reader) {
	data, err := io.ReadAll(r)
	if err != nil {
		fmt.Println("读失败:", err)
		return
	}
	fmt.Println("读到的内容:", string(data))
}

// 自定义数据源：只会吐 "custom data"
type CustomSource struct {
	sent bool // 标记是否已经吐过了
}

func (c *CustomSource) Read(p []byte) (int, error) {
	if c.sent {
		return 0, io.EOF // 吐完了
	}
	c.sent = true
	return copy(p, "custom data"), nil
}

func main() {
	// ① 从字符串读（strings.Reader 已实现 io.Reader）
	printAll(strings.NewReader("hello from string"))

	// ② 从内存读（bytes.Buffer 已实现 io.Reader）
	var buf bytes.Buffer
	buf.WriteString("hello from buffer")
	printAll(&buf)

	// ③ 从自定义数据源读（自己实现的 io.Reader）
	printAll(&CustomSource{})
}
