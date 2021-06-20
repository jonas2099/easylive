package processor

import (
	"github.com/haroldleong/easylive/conn"
	"github.com/haroldleong/easylive/stream"
	"github.com/haroldleong/easylive/util"
	log "github.com/sirupsen/logrus"
	"sync"
)

// 存储各直播流的读写流，转发
var streamMap *sync.Map

func init() {
	streamMap = &sync.Map{}
}

type ConnProcessor struct {
	conn *conn.Conn
}

func New(conn *conn.Conn) *ConnProcessor {
	return &ConnProcessor{
		conn: conn,
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
	app := p.conn.ConnInfo.App
	log.Debugf("processStream.start,app:%s", app)
	appStream := p.getStream(app)
	if p.conn.ConnType == conn.ConnectionTypePublish {
		go appStream.ReadingData(p.conn)
	} else if p.conn.ConnType == conn.ConnectionTypePull {
		if err := appStream.AddAudienceWriteEvent(p.conn); err != nil {
			return nil
		}
	}
	return nil
}

func (p *ConnProcessor) getStream(app string) *stream.AppStream {
	if tmp, ok := streamMap.Load(app); ok {
		return tmp.(*stream.AppStream)
	}
	newApp := stream.NewAppStream()
	streamMap.Store(app, newApp)
	return newApp
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
		cs := stream.GetChunk(p.conn)
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
