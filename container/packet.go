package container

import (
	"github.com/haroldleong/easylive/conn"
	"github.com/haroldleong/easylive/consts"
)

type PacketType int32

const (
	PacketTypeUnknown PacketType = iota
	PacketTypeAudio
	PacketTypeVideo
	PacketTypeMetaData
)

type Packet struct {
	OriginChunk conn.ChunkStream // 原始的chunk
	PacketType  PacketType       // 类型
	Tag         *Tag             // flv Tag
}

func (p *Packet) IsType(xType PacketType) bool {
	return p.PacketType == xType
}

func GetPacketByChunk(cs *conn.ChunkStream) *Packet {
	var packetType PacketType
	if cs.TypeID == consts.MsgTypeIDDataMsgAMF0 || cs.TypeID == consts.MsgTypeIDDataMsgAMF3 {
		packetType = PacketTypeUnknown
	} else if cs.TypeID == consts.MsgTypeIDAudioMsg {
		packetType = PacketTypeAudio
	} else if cs.TypeID == consts.MsgTypeIDVideoMsg {
		packetType = PacketTypeVideo
	}
	packet := &Packet{
		OriginChunk: *cs,
		PacketType:  packetType,
	}
	packet.parseTag()
	return packet
}

func (p *Packet) parseTag() {
	tag := &Tag{
		VideoHeader: &VideoTagHeader{},
		AudioHeader: &AudioTagHeader{},
	}
	var (
		n int
	)
	if p.PacketType == PacketTypeAudio {
		n, _ = tag.AudioHeader.parseAudioHeader(p.OriginChunk.Data)
	} else if p.PacketType == PacketTypeVideo {
		n, _ = tag.VideoHeader.parseVideoHeader(p.OriginChunk.Data)
	}
	tag.Data = p.OriginChunk.Data[n:]
	p.Tag = tag
}
