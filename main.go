package main

import (
	"cctools/ccbase"
	"cctools/ccnet"
	"cctools/myspider"
	"net"
)


func main(){
	s := ccnet.NewServer(func(conn net.Conn) ccnet.TcpSocketInterface {
		return myspider.NewSpiderClient(conn)
	})
	s.Start(1001)
	ccbase.Wait()
}
