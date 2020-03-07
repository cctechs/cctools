package ccnet

import (
	"fmt"
	"github.com/cctechs/cctools/cclog"
	"net"
	"sync"
	"time"
)

type FuncAddNewClient func(conn net.Conn) TcpSocketInterface

type Server struct {
	mutex* sync.Mutex
	mapSocks map[string]TcpSocketInterface
	fnAddClient FuncAddNewClient
}

func NewServer(f FuncAddNewClient) *Server{
	return &Server{
		mutex:    new(sync.Mutex),
		mapSocks: make(map[string]TcpSocketInterface),
		fnAddClient:f,
	}
}

func (s *Server) AcceptNewClient(f FuncAddNewClient){
	s.fnAddClient = f
}

func (s *Server) Start(port int) error{
	cclog.LogNotice("listening on port %d", port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil{
		cclog.LogError("listen on port %v failed, err:%v", port, err)
		return err
	}

	go func() {
		for{
			conn, err := listener.Accept()
			if err != nil{
				cclog.LogError("accept error %v", err.Error())
				continue
			}
			sock := s.addNewConn(conn)
			if sock != nil{
				go sock.doWork()
			}
		}
	}()

	s.checkStatus()

	return nil
}

func (s *Server) addNewConn(conn net.Conn) TcpSocketInterface{
	s.mutex.Lock()
	defer s.mutex.Unlock()
	socket := s.fnAddClient(conn)
	addr := conn.RemoteAddr().String()
	if _, ok := s.mapSocks[addr]; !ok{
		 s.mapSocks[addr] = socket
		 cclog.LogInfo("add new connection:%v success", conn.RemoteAddr())
		 return socket
	}else{
		cclog.LogError("add new connection failed, %s existed", addr)
		return nil
	}
}

func (s *Server) checkStatus(){
	go func() {
		timer := time.NewTicker(time.Second*1)
		for{
			select {
			case <- timer.C:
				s.mutex.Lock()
				// check if it's alive
				addrs := make([]string, len(s.mapSocks))
				for k, v := range s.mapSocks{
					if !v.isAlive(){
						addrs = append(addrs, k)
					}
				}

				// remove dead connection
				for _, v := range addrs{
					s.removeConn(v)
				}
				s.mutex.Unlock()
			}
		}
	}()
}

func (s *Server) removeConn(addr string){
	if _, ok := s.mapSocks[addr]; ok{
		delete(s.mapSocks, addr)
		cclog.LogInfo("remove connection %v", addr)
	}
}