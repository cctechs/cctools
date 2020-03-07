package ccnet


type PacketHeader struct {
	Cmd     uint32 // 协议号
	Size    uint32 // 协议大小
	Seq     uint32 // 保留
	Session uint32 // 保留
	Version uint32 // 保留
}
