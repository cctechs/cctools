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
	CBClosed    FnCallback
	CBKeepAlive FnCallback
	CBLogin     FnCallback
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
