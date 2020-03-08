package ccnet

import (
	"errors"
	"github.com/cctechs/cctools/ccbase"
	"github.com/cctechs/cctools/cclog"
	"net"
	"time"
)

type PacketRawData []byte

type FnRecvCallback func(*ccbase.PacketHeader, []byte)
type FnCheckPacketCallback func(*ccbase.PacketHeader) error
type FnClosedCallback func()
type TcpSocket struct {
	conn          net.Conn
	buff          *streamBuffer
	headerSize    uint32
	inPacket      chan PacketRawData
	outPacket     chan PacketRawData
	alive         ccbase.AtomBool
	CBRecvPacket  FnRecvCallback
	CBClosed      FnClosedCallback
	CBCheckPacket FnCheckPacketCallback
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
	if t.CBRecvPacket != nil {
		header := &ccbase.PacketHeader{}
		err := ccbase.UnSerializeFromBytes(buf[:ccbase.PK_HEADER_SIZE], header)
		if nil == err {
			t.CBRecvPacket(header, buf[ccbase.PK_HEADER_SIZE:])
		} else {
			cclog.LogError("recvpacket error, %v", err)
		}
	}
}

func (t *TcpSocket) SendPacket(buf []byte) error {
	if t.isAlive() {
		t.outPacket <- buf
		return nil
	}
	return errors.New("socket not alive")
}

func (t *TcpSocket) SetAlive(flag bool) {
	t.alive.Set(true)
}

func (t *TcpSocket) doWork() {
	t.SetAlive(true)
	defer func() {
		cclog.LogNotice("closed, %v", t.conn.RemoteAddr())
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
					t.conn.SetWriteDeadline(time.Now().Add(time.Duration(300) * time.Second))
				} else {
					break
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
		cclog.LogInfo("read data from %v,len=%v, data=%v", t.conn.RemoteAddr(), n, buf[:n])
		t.conn.SetReadDeadline(time.Now().Add(time.Duration(300) * time.Second))
		t.buff.write(buf[:n])

		pkLen, err := t.checkpacket()
		cclog.LogInfo("checkpacket, len=%d, err=%v", pkLen, err)
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
	cclog.LogInfo("checkpacket, data=%v", data)
	if data != nil {
		header := &ccbase.PacketHeader{}
		err := ccbase.UnSerializeFromBytes(data, header)
		if nil == err {
			if t.CBCheckPacket != nil{
				err = t.CBCheckPacket(header)
				if err != nil{
					return 0, err
				}
			}
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
	t.conn.Close()
	cclog.LogError("close socket %v", t.conn.RemoteAddr().String())
	if t.CBClosed != nil {
		t.CBClosed()
	}
}

func (t *TcpSocket) isAlive() bool {
	return t.alive.Get()
}
