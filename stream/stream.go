package stream

import (
	newamf "github.com/gwuhaolin/livego/protocol/amf"
	"github.com/haroldleong/easylive/conn"
	"github.com/haroldleong/easylive/consts"
	"github.com/haroldleong/easylive/container"
	"github.com/haroldleong/easylive/util"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Stream struct {
	conn        *conn.Conn
	mutex       sync.Mutex
	packetQueue chan *container.Packet
	init        bool
}

func New(connection *conn.Conn) *Stream {
	return &Stream{
		conn:        connection,
		packetQueue: make(chan *container.Packet, 2048),
	}
}

func (s *Stream) CheckPull() {
	for {
		if cs := GetChunk(s.conn); cs == nil {
			log.Errorf("CheckPull.close")
			return
		} else {
			log.Debugf("CheckPull.cs:%v", util.JSON(cs))
		}
	}
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
			log.Infof("getStreamChunkStream.get media data.cs:%v", util.JSON(cs))
			break
		}
		if cs.TypeID == consts.MsgTypeIDAudioMsg ||
			cs.TypeID == consts.MsgTypeIDVideoMsg {
			break
		}
	}
	return cs
}

func (s *Stream) sendStreamChunk(cs *conn.ChunkStream) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if cs.TypeID == consts.MsgTypeIDDataMsgAMF0 ||
		cs.TypeID == consts.MsgTypeIDDataMsgAMF3 {
		var err error
		if cs.Data, err = newamf.MetaDataReform(cs.Data, newamf.DEL); err != nil {
			return err
		}
		cs.Length = uint32(len(cs.Data))
	}
	return s.conn.WriteAndFlush(cs)
}

func GetChunk(newConn *conn.Conn) *conn.ChunkStream {
	// read chunk
	var chunk *conn.ChunkStream
	for {
		var err error
		chunk, err = newConn.ReadChunk()
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
