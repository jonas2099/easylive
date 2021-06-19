package processor

import (
	"github.com/haroldleong/easylive/conn"
	"github.com/haroldleong/easylive/stream"
	"github.com/haroldleong/easylive/util"
	log "github.com/sirupsen/logrus"
)

type ConnProcessor struct {
	conn   *conn.Conn
	stream *stream.Stream
}

func New(conn *conn.Conn) *ConnProcessor {
	return &ConnProcessor{
		conn: conn,
		stream: &stream.Stream{
			Conn: conn,
		},
	}
}

func (p *ConnProcessor) HandleConn() {
	if err := p.handshake(); err != nil {
		return
	}
	if err := p.handleConnect(); err != nil {
		return
	}
	// 开始读数据
	log.Infof("HandleConn.ready process stream.connInfo:%s", util.JSON(p.conn.ConnInfo))
	if err := p.processStream(); err != nil {
		return
	}
}

func (p *ConnProcessor) processStream() error {
	if p.conn.ConnType == conn.ConnectionTypePublish {
		go p.stream.KeepReadingData()
	} else if p.conn.ConnType == conn.ConnectionTypePull {
		p.stream.AddAudience()
	}
	return nil
}

func (p *ConnProcessor) handshake() error {
	// handshake
	if err := p.conn.HandshakeServer(); err != nil {
		p.conn.NetConn.Close()
		log.Error("handshake.HandshakeServer err:%v ", err)
		return err
	}
	log.Debugf("handshake.handshake success")
	return nil
}

func (p *ConnProcessor) handleConnect() error {
	// 连接
	for {
		cs := p.stream.GetChunk()
		log.Infof("HandleConn.ready process chunk.len:%d", cs.Length)
		if err := p.conn.HandleChunk(cs); err != nil {
			log.Errorf("HandleConn HandleChunk err:%v", err)
			return err
		}
		if p.conn.MessageDone() {
			break
		}
	}
	return nil
}
