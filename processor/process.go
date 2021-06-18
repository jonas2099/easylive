package processor

import (
	"github.com/haroldleong/easylive/conn"
	"github.com/haroldleong/easylive/util"
	log "github.com/sirupsen/logrus"
)

type ConnProcessor struct {
	conn *conn.Conn
}

func New(conn *conn.Conn) *ConnProcessor {
	return &ConnProcessor{
		conn: conn,
	}
}

func (p *ConnProcessor) HandleConn() {
	// handshake
	if err := p.conn.HandshakeServer(); err != nil {
		p.conn.NetConn.Close()
		log.Error("HandleConn HandshakeServer err:%v ", err)
		return
	}
	log.Debugf("HandleConn.handshake success")
	// 连接
	for {
		cs := p.getChunk()
		log.Infof("HandleConn.ready process chunk.len:%d", cs.Length)
		if err := p.conn.HandleChunk(cs); err != nil {
			log.Errorf("HandleConn HandleChunk err:%v", err)
			return
		}
		if p.conn.MessageDone() {
			break
		}
	}
	// 开始读数据
	log.Infof("HandleConn.ready process stream.connInfo:%s", util.JSON(p.conn.ConnInfo))
}

func (p *ConnProcessor) getChunk() *conn.ChunkStream {
	// read chunk
	var chunk *conn.ChunkStream
	for {
		var err error
		chunk, err = p.conn.ReadChunk()
		if err != nil {
			log.Errorf("getChunk.ReadChunk fail.err:%v", err)
			return nil
		}
		if chunk.Full() {
			break
		}
	}
	if chunk != nil {
		p.conn.Ack(chunk)
	}
	return chunk
}
