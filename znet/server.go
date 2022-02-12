package znet

import (
	"fmt"
	"go-tcp-server/utils"
	"go-tcp-server/ziface"
	"net"
	"time"
)

type Server struct {
	Name string

	IPVersion string

	IP string

	Port int

	msgHandler ziface.IMsgHandler

	ConnMgr ziface.IConnManager

	OnConnStart func(conn ziface.IConnection)

	OnConnStop func(conn ziface.IConnection)
}

func (s *Server) Start() {

	fmt.Printf("[START] Server name: %s, listener at IP: %s, Port %d is starting\n", s.Name, s.IP, s.Port)
	fmt.Printf("[Zinx] Version: %s, MaxConn: %d, MaxPacketSize: %d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)

	go func() {

		s.msgHandler.StartWorkerPool()

		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err: ", err)
			return
		}

		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}

		fmt.Println("start Zinx server ", s.Name, " succ, now listening")

		var cid uint32
		cid = 0

		for {

			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}

			// TODO
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				fmt.Println("too many connections MaxConn =", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}

			dealConn := NewConnection(s, conn, cid, s.msgHandler)
			cid++
			fmt.Println("conn start connId =", dealConn.ConnID)
			go dealConn.Start()
		}
	}()
}

func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)
	fmt.Println("add router succ! ")
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server , name ", s.Name)
	s.ConnMgr.ClearConn()
	//TODO
}

func (s *Server) Serve() {
	s.Start()

	//TODO

	for {
		time.Sleep(10 * time.Second)
	}
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

func (s *Server) SetOnConnStart(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

func (s *Server) SetOnConnStop(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> CallOnConnStart")
		s.OnConnStart(conn)
	}
}

func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---> CallOnConnStop")
		s.OnConnStop(conn)
	}
}

func NewServer() ziface.IServer {

	utils.GlobalObject.Reload()

	s := &Server{
		Name:       utils.GlobalObject.Name,
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TcpPort,
		msgHandler: NewMsgHandler(),
		ConnMgr:    NewConnManager(),
	}

	return s
}
