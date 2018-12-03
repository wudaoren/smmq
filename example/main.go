package main

import (
	"fmt"
	"time"

	"github.com/wudaoren/smmq"
)

const (
	HOST = "127.0.0.1:7712"
)

var sev *smmq.Server
var cli *smmq.Client

func main() {
	fmt.Println(byte('\n'))
	go server()
	go client()
	go func() {
		id := 0
		time.Sleep(time.Second * 3)
		for {
			id++
			time.Sleep(time.Second * 1)
			cli.Publish("/user", makeStr())
			fmt.Println("发布消息：", id)
		}
	}()
	time.Sleep(time.Hour)
}
func makeStr() string {
	str := ""
	for i := 0; i < 11; i++ {
		str += "a"
	}
	return str
}

func server() {
	sev = smmq.NewServer()
	sev.RegAuthHandler(func(auth *smmq.Auth) bool {
		fmt.Println("请求连接：", auth)
		return true
	})
	sev.RegErrorHandler(func(e interface{}) {
		fmt.Println("遇到错误：", e)
	})
	sev.RegCloseHandler(func(c *smmq.Conn) {
		fmt.Println("连接关闭", c)
	})
	e := sev.Listen(HOST)
	fmt.Println("服务端启动结果：", e)
}

func client() {
	var e error
	cli, e = smmq.NewClient(HOST, &smmq.Auth{
		Username: "wang",
		Password: "hait",
	})
	cli.RegCloseHandler(func(c *smmq.Conn) {
		fmt.Println("连接已经被关闭。")
	})
	cli.Subscribe("/user", func(m *smmq.Topic) {
		fmt.Println("收到订阅x(/user):", string(m.Body))
	})
	cli.Subscribe("/age", func(m *smmq.Topic) {
		fmt.Println("收到订阅(/age):", string(m.Body))
	})
	//e := cli.Connect(HOST)
	fmt.Println("客户端连接结果：", e)
}
