package ccnet

import (
	"errors"
	"github.com/cctechs/cctools/cclog"
	"net"
)

type FnRegisterSocket func(conn net.Conn) TcpSocketInterface

type Client struct {
	sock TcpSocketInterface
}

func NewClient() *Client{
	return &Client{}
}

func (c *Client) Start(addr string, fn FnRegisterSocket) error{
	conn, err := net.Dial("tcp", addr)
	if err != nil{
		cclog.LogError("connect %s failed, err %v", addr, err)
		return err
	}
	c.sock = fn(conn)
	if c.sock != nil{
		go c.sock.DoWork()
		return nil
	}
	return errors.New("create sockets failed")
}