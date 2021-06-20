package stream

import (
	newamf "github.com/gwuhaolin/livego/protocol/amf"
	"github.com/haroldleong/easylive/conn"
	"github.com/haroldleong/easylive/consts"
	"github.com/haroldleong/easylive/util"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Stream struct {
	conn        *conn.Conn
	mutex       sync.Mutex
	packetQueue chan *conn.ChunkStream
}

func New(connection *conn.Conn) *Stream {
	return &Stream{
		conn:        connection,
		packetQueue: make(chan *conn.ChunkStream, 2048),
	}
}

func (s *Stream) writeToAudience() {
	for {
		// 接收流
		cs, ok := <-s.packetQueue
		log.Debugf("writeToAudience.get data.cs:%v", util.JSON(cs))
		if ok {
			if err := s.sendStreamChunk(cs); err != nil {
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
		if cs.TypeID == consts.MsgTypeIDAudioMsg ||
			cs.TypeID == consts.MsgTypeIDVideoMsg ||
			cs.TypeID == consts.MsgTypeIDDataMsgAMF0 ||
			cs.TypeID == consts.MsgTypeIDDataMsgAMF3 {
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
