package conn

type ChunkStream struct {
	Format             uint8
	CSID               uint32
	Timestamp          uint32
	Length             uint32 // chunk总长度
	TypeID             uint32
	StreamID           uint32
	timeDelta          uint32 // 增量时间戳
	useExtendTimeStamp bool   // 是否需要读取扩展时间戳

	remain uint32 // 剩下需要读取的字节数
	Data   []byte

	index uint32

	finish bool
}

func (cs *ChunkStream) Full() bool {
	return cs.finish
}
func (cs *ChunkStream) initData() {
	cs.finish = false
	cs.index = 0
	cs.remain = cs.Length
	cs.Data = make([]byte, cs.Length)
}
