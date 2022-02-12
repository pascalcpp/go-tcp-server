package utils

import (
	"encoding/json"
	"go-tcp-server/ziface"
	"io/ioutil"
)

type GlobalObj struct {
	TcpServer ziface.IServer
	Host      string
	TcpPort   int
	Name      string

	Version          string
	MaxPacketSize    uint32
	MaxConn          int
	WorkerPoolSize   uint32
	MaxWorkerTaskLen uint32
	MaxMsgChanLen    int

	ConfFilePath string
}

var GlobalObject *GlobalObj

func (g *GlobalObj) Reload() {
	data, err := ioutil.ReadFile(g.ConfFilePath)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, GlobalObject)
	if err != nil {
		panic(err)
	}
}

func init() {
	GlobalObject = &GlobalObj{
		Host:             "0.0.0.0",
		TcpPort:          4396,
		Name:             "ZinxServer",
		Version:          "v1.0",
		ConfFilePath:     "conf/zinx.json",
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
		MaxPacketSize:    4096,
		MaxConn:          12000,
		MaxMsgChanLen:    128,
	}

	GlobalObject.Reload()
}
