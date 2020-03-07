package ccnet

import (
	"errors"
	"github.com/cctechs/cctools/ccbase"
	"github.com/cctechs/cctools/cclog"
	"net"
)

type PacketRawData []byte

type TcpSocket struct {
	conn       net.Conn
	buff       *streamBuffer
	headerSize uint32
	inPacket   chan PacketRawData
	outPacket  chan PacketRawData
	isAlive   ccbase.AtomBool
}

func NewTcpSocket(conn net.Conn) *TcpSocket {
	return &TcpSocket{
		conn:       conn,
		buff:       newStreamBuffer(),
		headerSize: ccbase.StreamSizeof(&PacketHeader{}),
		inPacket:   make(chan PacketRawData, 32),
		outPacket:  make(chan PacketRawData, 32),
	}
}

func (t *TcpSocket) RecvPacket(buf []byte) {

}

func (t* TcpSocket) SendPacket(buf []byte) error{
	if t.IsAlive(){
		t.outPacket <- buf
		return nil
	}
	return errors.New("socket not alive")
}

func (t *TcpSocket) DoWork() {
	t.isAlive.Set(true)

	defer func() {
		if t.IsAlive(){
			t.Close()
		}
	}()

	// send packet
	go func() {
		for{
			select {
			case  d, ok := <- t.outPacket:
				if ok{
					totalLen := len(d)
					for{
						if totalLen > 0{
							n, err := t.conn.Write(d)
							if err != nil{
								cclog.LogError("write to %v error, err=%v", t.conn.RemoteAddr(), err)
								return
							}
							totalLen = totalLen - n
							d = d[n:]
						}
					}
				} else {
					cclog.LogError("why here???")
				}
			}
		}
	}()

	// recv packet
	go func() {
		for {
			select {
			case d, ok := <-t.inPacket:
				if ok {
					t.RecvPacket(d)
				} else {
					cclog.LogError("why here???")
				}
			}
		}
	}()

	// read data
	buf := make([]byte, 1024)
	for {
		n, err := t.conn.Read(buf)
		if err != nil {
			cclog.LogError("read from %s error, err=%v", t.conn.RemoteAddr(), err)
			return
		}
		t.buff.write(buf[:n])

		pkLen, err := t.checkpacket()
		if err != nil{
			cclog.LogError("checkpacket from %s error, err=%v", t.conn.RemoteAddr(), err)
			break
		}
		if pkLen > 0 {
			if t.buff.size() >= int(pkLen) {
				data := t.buff.read(pkLen)
				t.inPacket <- data
			}
		}
	}
}

func (t *TcpSocket) checkpacket() (uint32, error) {
	data := t.buff.copy(t.headerSize)
	if data != nil {
		header := &PacketHeader{}
		err := ccbase.UnSerializeFromBytes(data, header)
		if nil == err {
			return header.Size, nil
		}else{
			cclog.LogError("UnSerializeFromBytes failed, err=%v", err.Error())
			return 0, err
		}
	}
	return 0, nil
}

func (t *TcpSocket) Close() {
	close(t.inPacket)
	t.isAlive.Set(false)
	cclog.LogError("close socket %v", t.conn.RemoteAddr().String())
}

func (t *TcpSocket) IsAlive() bool {
	return t.isAlive.Get()
}
