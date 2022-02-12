package znet

import (
	"errors"
	"fmt"
	"go-tcp-server/utils"
	"go-tcp-server/ziface"
	"io"
	"net"
	"sync"

)

type Connection struct {
	TcpServer ziface.IServer

	Conn *net.TCPConn

	ConnID uint32

	isClosed bool

	MsgHandler ziface.IMsgHandler

	ExitBuffChan chan bool

	msgChan chan []byte

	msgBuffChan chan []byte

	property map[string]interface{}

	// 如果有结构体的类型的属性
	// 如果创建Connection没有显示创建，会默认调用&sync.RWMutex{} 创建该类型实例
	//
	propertyLock sync.RWMutex
}

func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandler) *Connection {
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte),
		msgBuffChan:  make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}),
	}

	c.TcpServer.GetConnMgr().Add(c)
	return c
}

func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn writer exit!]", " connId =", c.ConnID)

	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("send data error: ", err, " conn writer exit")
				return
			}

		case data, ok := <-c.msgBuffChan:
			if ok {
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("send buff data error:", err, " conn writer exit!")
					return
				}
			} else {
				fmt.Println("msgBuffChan is closed")
				return
			}

		case <-c.ExitBuffChan:
			return
		}
	}
}

func (c *Connection) StartReader() {
	fmt.Println("[Reader Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn reader exit!]", " connId =", c.ConnID)
	defer c.Stop()

	dp := NewDataPack()
	for {

		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error", err)
			c.ExitBuffChan <- true
			return
		}

		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error", err)
			c.ExitBuffChan <- true
			return
		}

		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error", err)
				c.ExitBuffChan <- true
				return
			}
		}

		msg.SetData(data)
		req := &Request{
			conn: c,
			msg:  msg,
		}

		if utils.GlobalObject.WorkerPoolSize > 0 {
			c.MsgHandler.SendMsgToTaskQueue(req)
		} else {
			go c.MsgHandler.DoMsgHandler(req)
		}

	}

}

func (c *Connection) Start() {

	go c.StartReader()

	go c.StartWriter()

	c.TcpServer.CallOnConnStart(c)

	//for {
	//	select {
	//	case <-c.ExitBuffChan:
	//		return
	//	}
	//}

}

func (c *Connection) Stop() {

	fmt.Println("conn stop.. connId =", c.ConnID)
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	c.TcpServer.CallOnConnStop(c)

	// close tcp conn
	c.Conn.Close()

	c.ExitBuffChan <- true
	c.TcpServer.GetConnMgr().Remove(c)

	close(c.ExitBuffChan)
	close(c.msgChan)
	close(c.msgBuffChan)
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) GetConnId() uint32 {
	return c.ConnID
}

func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("connection closed when send MSG")
	}
	dp := NewDataPack()
	packMsg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id =", msgId)
		return errors.New("conn write error")
	}

	c.msgChan <- packMsg

	return nil
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("connection closed when send buff msg")
	}

	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("pack error msg id =", msgId)
		return errors.New("pack error msg")
	}

	c.msgBuffChan <- msg
	return nil
}

func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
