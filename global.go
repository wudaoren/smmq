package smmq

import (
	"encoding/json"
	"fmt"
)

//订阅处理函数
type SubscribeHandlerFunc func(*Topic)

//错误处理函数
type ErrorHandlerFunc func(interface{})

//心跳处理函数
type HeartbeatHandlerFunc func()

//关闭处理函数
type CloseHandlerFunc func(*Conn)

//消息主体
type Message struct {
	Type    int8   `json:",omitempty"` //消息发布类型
	Payload []byte `json:",omitempty"` //
	Token   int    `json:",omitempty"`
}

//登录认证
type Auth struct {
	Username   string `json:",omitempty"`
	Password   string `json:",omitempty"`
	RemoteAddr string `json:",omitempty"`
}

func (this *Auth) unmarshal(data []byte) error {
	return json.Unmarshal(data, this)
}

//认证回调
type AuthHandlerFunc func(*Auth) bool

//话题
type Topic struct {
	Topic string
	Body  []byte
}

func (this *Topic) unmarshal(data []byte) error {
	return json.Unmarshal(data, this)
}

//将body的值赋值给v
func (this *Topic) Set(v interface{}) error {
	return json.Unmarshal(this.Body, v)
}

const (
	MSG_TYPE_SUB            = 1  //订阅消息
	MSG_TYPE_SUB_SUCC       = 10 //订阅成功
	MSG_TYPE_PUB            = 2  //发布消息
	MSG_TYPE_PUB_SUCC       = 20 //发布成功
	MSG_TYPE_MSG            = 3  //收到消息
	MSG_TYPE_AUTH           = 4  //身份验证
	MSG_TYPE_AUTH_SUCC      = 40 //身份验证通过
	MSG_TYPE_AUTH_ERR       = 41 //身份验证失败
	MSG_TYPE_HEARTBEAT      = 5  //心跳
	MSG_TYPE_HEARTBEAT_SUCC = 50 //心跳返回成功
)

//检查错误是否pannic
func checkErrorPanic(e error) {
	if e != nil {
		panic(e)
	}
}
func debug(args ...interface{}) {
	fmt.Println(args...)
}
