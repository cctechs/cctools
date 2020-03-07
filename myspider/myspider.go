package myspider

import (
	"cctools/ccnet"
	"net"
)

type SpiderClient struct {
	*ccnet.TcpSocket  // drived
}

func NewSpiderClient(conn net.Conn) *SpiderClient{
	s := new(SpiderClient)
	s.TcpSocket = ccnet.NewTcpSocket(conn)
	return s
}
