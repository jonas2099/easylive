package conn

const (
	FlvTimestampMax   = 0xFFFFFF
	chunkHeaderLength = 12
)

type UserControl uint32

const (
	UserControlStreamBegin      UserControl = 0
	UserControlStreamEOF        UserControl = 1
	UserControlStreamDry        UserControl = 2
	UserControlSetBufferLen     UserControl = 3
	UserControlStreamIsRecorded UserControl = 4
	UserControlPingRequest      UserControl = 6
	UserControlPingResponse     UserControl = 7
)
