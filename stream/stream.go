package stream

import (
	uuid "github.com/google/uuid"
	"github.com/haroldleong/easylive/conn"
	"github.com/haroldleong/easylive/consts"
	"github.com/haroldleong/easylive/container"
	"github.com/haroldleong/easylive/util"
	log "github.com/sirupsen/logrus"
	"io"
)

type streamType int32

const (
	streamTypeAnchor streamType = iota
	streamTypeAudience
)

type Stream struct {
	id          string                 // 唯一标识
	conn        *conn.Conn             //连接
	packetQueue chan *container.Packet // 流chan，用于拉流端
	init        bool                   // 是否初始化
	streamType  streamType
}

func New(connection *conn.Conn, anchor bool) *Stream {
	s := &Stream{
		id:   uuid.New().String(),
		conn: connection,
	}
	if !anchor {
		s.packetQueue = make(chan *container.Packet, 2048)
		s.streamType = streamTypeAudience
	} else {
		s.streamType = streamTypeAnchor
	}
	return s
}

func (s *Stream) writeToAudience() {
	for {
		var cs conn.ChunkStream
		// 接收流
		p, ok := <-s.packetQueue
		if ok {
			cs.Data = p.OriginChunk.Data
			cs.Length = uint32(len(p.OriginChunk.Data))
			cs.StreamID = p.OriginChunk.StreamID
			cs.Timestamp = p.OriginChunk.Timestamp
			cs.TypeID = p.OriginChunk.TypeID
			// log.Debugf("writeToAudience.get data.cs:%v", util.JSON(cs))
			if err := s.sendStreamChunk(&cs); err != nil {
				log.Errorf("writeToAudience.sendStreamChunk err.%v", err)
				return
			}
		}
	}
}

func (s *Stream) getStreamChunkStream() *conn.ChunkStream {
	var cs *conn.ChunkStream
	for {
		cs = GetChunk(s.conn)
		if cs == nil {
			log.Errorf("getStreamChunkStream.no chunk")
			break
		}
		if cs.TypeID == consts.MsgTypeIDDataMsgAMF0 ||
			cs.TypeID == consts.MsgTypeIDDataMsgAMF3 {
			log.Infof("getStreamChunkStream.get amf data.cs:%v", util.JSON(cs))
			break
		}
		if cs.TypeID == consts.MsgTypeIDAudioMsg ||
			cs.TypeID == consts.MsgTypeIDVideoMsg {
			log.Infof("getStreamChunkStream.get media data.cs:%v", util.JSON(cs))
			break
		}
	}
	return cs
}

func (s *Stream) sendStreamChunk(cs *conn.ChunkStream) error {
	return s.conn.WriteChunk(cs)
}

func GetChunk(newConn *conn.Conn) *conn.ChunkStream {
	// read chunk
	var chunk *conn.ChunkStream
	for {
		var err error
		chunk, err = newConn.ReadChunk()
		if err == io.EOF {
			log.Infof("getChunk.ReadChunk end")
			return nil
		}
		if err != nil {
			log.Errorf("getChunk.ReadChunk fail.err:%v", err)
			return nil
		}
		if chunk.Full() {
			break
		}
	}
	if chunk != nil {
		newConn.Ack(chunk)
	}
	return chunk
}
