package ccnet

import (
	"errors"
	"github.com/cctechs/cctools/ccbase"
	"github.com/cctechs/cctools/cclog"
	"net"
)

type PacketRawData []byte

type FnRecvCallback func([]byte)
type FnClosedCallback func()
type TcpSocket struct {
	conn         net.Conn
	buff         *streamBuffer
	headerSize   uint32
	inPacket     chan PacketRawData
	outPacket    chan PacketRawData
	alive      ccbase.AtomBool
	CBRecvPacket FnRecvCallback
	CBClosed     FnClosedCallback
}

func NewTcpSocket(conn net.Conn) *TcpSocket {
	return &TcpSocket{
		conn:       conn,
		buff:       newStreamBuffer(),
		headerSize: ccbase.PK_HEADER_SIZE,
		inPacket:   make(chan PacketRawData, 32),
		outPacket:  make(chan PacketRawData, 32),
	}
}

func (t *TcpSocket) recvpacket(buf []byte) {
	if t.CBRecvPacket != nil{
		t.CBRecvPacket(buf)
	}
}

func (t *TcpSocket) SendPacket(buf []byte) error {
	if t.isAlive() {
		t.outPacket <- buf
		return nil
	}
	return errors.New("socket not alive")
}

func (t *TcpSocket) doWork() {
	t.alive.Set(true)

	defer func() {
		if t.isAlive() {
			t.Close()
		}
	}()

	// send packet
	go func() {
		for ch := range t.outPacket {
			totalLen := len(ch)
			for {
				if totalLen > 0 {
					n, err := t.conn.Write(ch)
					if err != nil {
						cclog.LogError("write to %v error, err=%v", t.conn.RemoteAddr(), err)
						return
					}
					totalLen = totalLen - n
					ch = ch[n:]
				}
			}
		}
	}()

	// recv packet
	go func() {
		for ch := range t.inPacket {
			t.recvpacket(ch)
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
		if err != nil {
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
		header := &ccbase.PacketHeader{}
		err := ccbase.UnSerializeFromBytes(data, header)
		if nil == err {
			return header.Size, nil
		} else {
			cclog.LogError("UnSerializeFromBytes failed, err=%v", err.Error())
			return 0, err
		}
	}
	return 0, nil
}

func (t *TcpSocket) Close() {
	close(t.inPacket)
	close(t.outPacket)
	t.alive.Set(false)
	cclog.LogError("close socket %v", t.conn.RemoteAddr().String())
	if t.CBClosed != nil{
		t.CBClosed()
	}
}

func (t *TcpSocket) isAlive() bool {
	return t.alive.Get()
}
