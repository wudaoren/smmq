package smmq

import (
	"errors"
	"net"
	"sync"
	"time"
)

type Client struct {
	topics           *sync.Map //
	errHandler       ErrorHandlerFunc
	heartbeatHandler HeartbeatHandlerFunc
	closeHandler     CloseHandlerFunc
	auth             *Auth
	conn             *Conn //
	responseCode     chan int8
	sync.Mutex
}

func NewClient(addr string, auth *Auth) (*Client, error) {
	client := new(Client)
	client.topics = new(sync.Map)
	client.auth = auth
	client.responseCode = make(chan int8)
	return client, client.connect(addr)
}

//订阅
func (this *Client) Subscribe(topic string, handler SubscribeHandlerFunc) error {
	this.topics.Store(topic, handler)
	msg := &Message{Type: MSG_TYPE_SUB}
	msg.Payload = []byte(topic)
	if code, e := this.write(msg); e != nil {
		return e
	} else if code != MSG_TYPE_SUB_SUCC {
		return errors.New("subscribe err.")
	}
	return nil
}

//发布
func (this *Client) Publish(topic string, v interface{}) error {
	topicMsg := &Topic{Topic: topic}
	topicMsg.Body = this.conn.Marshal(v)

	msg := &Message{Type: MSG_TYPE_PUB}
	msg.Payload = this.conn.Marshal(topicMsg)
	if code, e := this.write(msg); e != nil {
		return e
	} else if code != MSG_TYPE_PUB_SUCC {
		return errors.New("publish err.")
	}
	return nil
}

//发送消息
func (this *Client) write(m *Message) (int8, error) {
	this.Lock()
	defer this.Unlock()
	var code int8
	var e error
	if this.conn == nil {
		return code, errors.New("not connect.")
	}
	if e = this.conn.Write(m); e != nil {
		return code, e
	}
	select {
	case <-time.After(time.Second * 5):
		e = errors.New("time out")
	case code = <-this.responseCode:
		e = nil
	}
	return code, e
}

//注册错误处理方法
func (this *Client) RegErrorHandler(handler ErrorHandlerFunc) {
	this.errHandler = handler
}

//注册关闭处理方法
func (this *Client) RegCloseHandler(handler CloseHandlerFunc) {
	this.closeHandler = handler
}

//注册心跳处理方法
func (this *Client) RegHeartbeatHandler(handler HeartbeatHandlerFunc) {
	this.heartbeatHandler = handler
}

//连接到消息队列服务器
func (this *Client) connect(addr string) error {
	c, e := net.Dial("tcp", addr)
	if e != nil {
		return e
	}
	go this.handleConnect(c)
	msg := &Message{Type: MSG_TYPE_AUTH}
	msg.Payload = this.conn.Marshal(this.auth)
	if code, e := this.write(msg); e != nil {
		return e
	} else if code != MSG_TYPE_AUTH_SUCC {
		return errors.New("auth err.")
	}
	return nil
}

//处理连接数据
func (this *Client) handleConnect(c net.Conn) {
	this.conn = newConn(c)
	this.conn.readingHandler(this.handleMessage)
	if this.closeHandler != nil {
		this.closeHandler(this.conn)
	}
}

//处理收到的消息
func (this *Client) handleMessage(m *Message) {
	switch m.Type {
	case MSG_TYPE_MSG:
		topicMsg := new(Topic)
		topicMsg.unmarshal(m.Payload)
		if handler, has := this.topics.Load(topicMsg.Topic); has {
			handler.(SubscribeHandlerFunc)(topicMsg)
		}
	default:
		this.responseCode <- m.Type
	}
}
