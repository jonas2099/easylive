package entity

import (
	"github.com/gwuhaolin/livego/av"
	"github.com/haroldleong/easylive/conn"
)

type Packet struct {
	CStream    conn.ChunkStream
	IsMetadata bool
	IsAudio    bool
	IsVideo    bool
	Header     av.PacketHeader
}
