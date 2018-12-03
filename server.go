package smmq

import (
	"net"
	"sync"
	"time"
)

type Server struct {
	index        int64            //索引编号
	topicsPool   *sync.Map        //topic连接池
	errorHandler ErrorHandlerFunc //错误处理
	authHandler  AuthHandlerFunc  //登录认证
	closeHandler CloseHandlerFunc //关闭处理
	pool         *sync.Map        //连接池
	length       int              //连接数量
}

func NewServer() *Server {
	server := new(Server)
	server.topicsPool = new(sync.Map)
	server.pool = new(sync.Map)
	server.checkConnectTimeout()
	return server
}

//连接数量
func (this *Server) ConnectLength() int {
	return this.length
}

//检查连接是否超时
func (this *Server) checkConnectTimeout() {
	go func() {
		defer func() {
			if e := recover(); e != nil && this.errorHandler != nil {
				this.errorHandler(e)
			}
		}()
		for {
			time.Sleep(time.Second * 30)
			this.length = 0
			this.pool.Range(func(id, c interface{}) bool {
				conn := c.(*Conn)
				if !conn.IsTimeout() {
					conn.Close()
				} else {
					this.length++
				}
				return true
			})
		}
	}()
}

//注册关闭处理方法
func (this *Server) RegCloseHandler(handler CloseHandlerFunc) {
	this.closeHandler = handler
}

//注册身份验证
func (this *Server) RegAuthHandler(handler AuthHandlerFunc) {
	this.authHandler = handler
}

//注册错误处理方法
func (this *Server) RegErrorHandler(handler ErrorHandlerFunc) {
	this.errorHandler = handler
}

//监听地址
func (this *Server) Listen(addr string) error {
	sev, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		c, e := sev.Accept()
		if e != nil {
			return e
		}
		go this.handleConnect(c)
	}
}

//创建连接索引编号
func (this *Server) makeIndex() int64 {
	this.index++
	return this.index
}

//处理接入的连接
func (this *Server) handleConnect(c net.Conn) {
	conn := newConn(c)
	conn.Id = this.makeIndex()
	//10秒后还没有认证通过则中断连接
	time.AfterFunc(time.Second*3, func() {
		if !conn.Pass {
			conn.Close()
		}
	})
	conn.readingHandler(this.handleMessage(conn))
	if this.closeHandler != nil {
		this.closeHandler(conn)
	}
}

//处理收到的消息
func (this *Server) handleMessage(c *Conn) func(*Message) {
	return func(m *Message) {
		defer func() {
			if e := recover(); e != nil && this.errorHandler != nil {
				this.errorHandler(e)
			}
		}()
		switch m.Type {
		case MSG_TYPE_PUB: //发布的消息
			topicMsg := new(Topic)
			topicMsg.unmarshal(m.Payload)
			if topic, has := this.topicsPool.Load(topicMsg.Topic); has {
				m.Type = MSG_TYPE_MSG
				topicMap := topic.(*sync.Map)
				topicMap.Range(func(id, conn interface{}) bool {
					//发送消息失败则删除连接
					if conn.(*Conn).Write(m) != nil {
						topicMap.Delete(id)
					}
					return true
				})
			}
			//返回发布成功
			c.Write(&Message{Type: MSG_TYPE_PUB_SUCC})
		case MSG_TYPE_SUB: //接收订阅消息
			topicName := string([]byte(m.Payload))
			topics, has := this.topicsPool.Load(topicName)
			if !has {
				topics = new(sync.Map)
				this.topicsPool.Store(topicName, topics)
			}
			topics.(*sync.Map).Store(c.Id, c)
			//返回订阅成功
			c.Write(&Message{Type: MSG_TYPE_SUB_SUCC})
		case MSG_TYPE_AUTH: //登录认证
			auth := new(Auth)
			auth.unmarshal(m.Payload)
			auth.RemoteAddr = c.RemoteAddr().String()
			if this.authHandler != nil && this.authHandler(auth) {
				c.Pass = true
				c.Write(&Message{Type: MSG_TYPE_AUTH_SUCC})
			} else {
				c.Write(&Message{Type: MSG_TYPE_AUTH_ERR})
			}
		case MSG_TYPE_HEARTBEAT: //心跳
			c.Write(&Message{Type: MSG_TYPE_HEARTBEAT_SUCC})
		}
	}
}
