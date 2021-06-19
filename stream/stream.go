package stream

import (
	"github.com/gwuhaolin/livego/av"
	newamf "github.com/gwuhaolin/livego/protocol/amf"
	"github.com/haroldleong/easylive/conn"
	log "github.com/sirupsen/logrus"
	"sync"
)

var streamMap *sync.Map

func init() {
	streamMap = &sync.Map{}
}

type Stream struct {
	Conn *conn.Conn
}

func (s *Stream) KeepReadingData() {
	app := s.Conn.ConnInfo.App
	log.Debugf("KeepReadingData.start,app:%s", app)
	for {
		var cs *conn.ChunkStream
		for {
			cs = s.GetChunk()
			if cs == nil {
				log.Errorf("KeepReadingData.no chunk")
				return
			}
			if cs.TypeID == av.TAG_AUDIO ||
				cs.TypeID == av.TAG_VIDEO ||
				cs.TypeID == av.TAG_SCRIPTDATAAMF0 ||
				cs.TypeID == av.TAG_SCRIPTDATAAMF3 {
				break
			}
		}

		if tmp, ok := streamMap.Load(app); ok {
			for _, audienceStream := range tmp.([]*Stream) {
				// log.Debugf("KeepReadingData.ready send data to audience")
				if err := audienceStream.setChunk(cs); err != nil {
					log.Errorf("KeepReadingData.setChunk err.%v", err)
					return
				}
			}
		}
	}
}

func (s *Stream) AddAudience() {
	app := s.Conn.ConnInfo.App
	log.Debugf("AddAudience.app:%s", app)
	var audienceList []*Stream
	if tmp, ok := streamMap.Load(app); ok {
		audienceList = tmp.([]*Stream)
	} else {
		audienceList = []*Stream{}
	}
	audienceList = append(audienceList, s)
	streamMap.Store(app, audienceList)
	log.Debugf("AddAudience finish,audienceCount:%v", streamMap)
}

func (s *Stream) GetChunk() *conn.ChunkStream {
	// read chunk
	var chunk *conn.ChunkStream
	for {
		var err error
		chunk, err = s.Conn.ReadChunk()
		if err != nil {
			log.Errorf("getChunk.ReadChunk fail.err:%v", err)
			return nil
		}
		if chunk.Full() {
			break
		}
	}
	if chunk != nil {
		s.Conn.Ack(chunk)
	}
	return chunk
}
func (s *Stream) setChunk(cs *conn.ChunkStream) error {
	if cs.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		cs.TypeID == av.TAG_SCRIPTDATAAMF3 {
		var err error
		if cs.Data, err = newamf.MetaDataReform(cs.Data, newamf.DEL); err != nil {
			return err
		}
		cs.Length = uint32(len(cs.Data))
	}
	return s.Conn.WriteAndFlush(cs)
}
