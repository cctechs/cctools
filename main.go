package main

import (
	"github.com/cctechs/cctools/ccbase"
	"github.com/cctechs/cctools/ccnet"
	"github.com/cctechs/cctools/myspider"
	"net"
)

func main(){
	s := ccnet.NewServer(func(conn net.Conn) ccnet.TcpSocketInterface {
		return myspider.NewSpiderClient(conn)
	})
	s.Start(1001)
	ccbase.Wait()
}
