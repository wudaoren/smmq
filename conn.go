package smmq

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"sync"
	"time"
)

type Conn struct {
	Id         int64
	Pass       bool //auth是否验证通过
	UpdateTime time.Time
	server     *Server
	net.Conn
	sync.Map
}

func newConn(c net.Conn) *Conn {
	conn := new(Conn)
	conn.Conn = c
	return conn
}

//是否超时
func (this *Conn) IsTimeout() bool {
	return true
}

//
func (this *Conn) Marshal(v interface{}) []byte {
	bt, _ := json.Marshal(v)
	return bt
}

//打包
func (this *Conn) Write(m *Message) error {
	bt := append(this.Marshal(m), '\n')
	debug("发送消息长度:", len(bt))
	_, e := this.Conn.Write(bt)
	return e
}

//读取数据
func (this *Conn) readingHandler(handler func(m *Message)) {
	reader := bufio.NewReader(this.Conn)
	for {
		data, e := reader.ReadSlice('\n')
		if e == io.EOF || e != nil {
			return
		}
		checkErrorPanic(e)
		debug("收到消息：", string(data))
		m := new(Message)
		e = json.Unmarshal(data, m)
		checkErrorPanic(e)
		this.UpdateTime = time.Now()
		handler(m)
	}
}
