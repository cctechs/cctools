package ccbase


type PacketHeader struct {
	Cmd     uint32 // 协议号
	Size    uint32 // 协议大小
	Seq     uint32 // 保留
	Session uint16 // 保留
	Version uint16 // 保留
}

const PK_HEADER_SIZE uint32 = 16