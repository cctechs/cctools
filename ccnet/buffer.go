package ccnet

type streamBuffer struct {
	buf []byte
}

func newStreamBuffer()*streamBuffer{
	return &streamBuffer{buf:make([]byte, 1024)}
}

func (s *streamBuffer) write(data []byte) {
	s.buf = append(s.buf, data...)
}

func (s *streamBuffer) read(size uint32) []byte {
	if 0 == size {
		panic("size == 0")
	}

	nLen := uint32(len(s.buf))
	if nLen < size {
		return nil
	}

	data := s.buf[:size]
	s.buf = s.buf[size:]
	return data
}

func (s *streamBuffer) copy(size uint32) []byte {
	if 0 == size {
		panic("size == 0")
	}
	nLen := uint32(len(s.buf))
	if nLen < size {
		return nil
	}
	data := make([]byte, size)
	copy(data, s.buf[:size])
	return data
}

func(s*streamBuffer) clear(){
	s.buf = make([]byte, 0)
}

func (s* streamBuffer) size() int{
	return len(s.buf)
}
