package ccnet

import (
	"errors"
	"github.com/cctechs/cctools/cclog"
	"net"
	"time"
)

type FnCallback func()

type Client struct {
	sock        *TcpSocket
	CBConnected FnCallback
	CBClosed    FnClosedCallback
	CBKeepAlive FnCallback
	CBLogin     FnCallback
	CBRecvPacket FnRecvCallback
	IsLogin     bool
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(addr string) error {
	tryToConnect := func() error {
		cclog.LogNotice("try to connect %s ...", addr)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			cclog.LogError("connect %s failed, %v", addr, err)
			return err
		}
		cclog.LogNotice("connected %s.", addr)
		c.sock = NewTcpSocket(conn)
		if c.sock != nil {
			c.sock.SetAlive(true)
			c.SetLoginFlag(false)
			c.sock.CBRecvPacket = c.CBRecvPacket
			c.sock.CBClosed = c.CBClosed
			go c.sock.doWork()
			return nil
		}
		return errors.New("create error")
	}

	go func() {
		t := time.NewTicker(time.Second)
		for {
			select {
			case <-t.C:
				if c.sock == nil {
					tryToConnect()
				} else {
					if !c.sock.isAlive() {
						tryToConnect()
					} else {
						if c.IsLogin {
							c.keepAlive()
						} else {
							c.login()
						}
					}
				}
			}
		}
	}()
	return errors.New("create sockets failed")
}

func(c *Client) closed(){
	c.SetLoginFlag(false)
	if c.CBClosed != nil{
		c.CBClosed()
	}
}

func (c *Client) connected() {
	if c.CBConnected != nil {
		c.CBConnected()
	}
}
func (c *Client) keepAlive() {
	if c.CBKeepAlive != nil {
		c.CBKeepAlive()
	}
}

func (c *Client) login() {
	if c.CBLogin != nil {
		c.CBLogin()
	}
}

func (c* Client) SetLoginFlag(flag bool){
	c.IsLogin = flag
}

func (c* Client) SendData(data []byte) error{
	if c.sock != nil{
		return c.sock.SendPacket(data)
	}
	return errors.New("no socket")
}
