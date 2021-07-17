// Package conn Protocol Control Messages（协议控制消息）。主要使用来沟通 RTMP 初始状态的相关连接信息，比如，windows size，chunk size
package conn

import (
	"github.com/Monibuca/engine/v2/util/bits/pio"
	"github.com/haroldleong/easylive/consts"
)

func (c *Conn) NewAck(size uint32) ChunkStream {
	return initControlMsg(consts.MsgTypeIDAck, 4, size)
}

func (c *Conn) NewSetChunkSize(size uint32) ChunkStream {
	return initControlMsg(consts.MsgTypeIDSetChunkSize, 4, size)
}

func (c *Conn) NewWindowAckSize(size uint32) ChunkStream {
	return initControlMsg(consts.MsgTypeIDWindowAckSize, 4, size)
}

func (c *Conn) NewSetPeerBandwidth(size uint32) ChunkStream {
	ret := initControlMsg(consts.MsgTypeIDSetPeerBandwidth, 5, size)
	ret.Data[4] = 2
	return ret
}

func initControlMsg(id, size, value uint32) ChunkStream {
	ret := ChunkStream{
		Format:   0,
		CSID:     2,
		TypeID:   id,
		StreamID: 0,
		Length:   size,
		Data:     make([]byte, size),
	}
	pio.PutU32BE(ret.Data[:size], value)
	return ret
}
