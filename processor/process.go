package processor

import (
	"fmt"
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

func (p *ConnProcessor) HandleConn() error {
	if err := p.handshake(); err != nil {
		return err
	}
	if err := p.handleConnect(); err != nil {
		return err
	}
	// 开始读数据
	log.Infof("HandleConn.ready process stream.connInfo:%s", util.JSON(p.conn.ConnInfo))
	if err := p.processStream(); err != nil {
		return err
	}
	return nil
}

func (p *ConnProcessor) processStream() error {
	app := p.conn.ConnInfo.App
	if p.conn.ConnType == conn.ConnectionTypePublish {
		log.Debugf("processStream.start publish,app:%s", app)
		appStream, _ := p.getStream(app, false)
		go appStream.ReadingData(p.conn)
	} else if p.conn.ConnType == conn.ConnectionTypePull {
		appStream, err := p.getStream(app, true)
		if err != nil {
			log.Errorf("processStream.stream not exsit,app:%s", app)
			return err
		}
		log.Debugf("processStream.start pull,app:%s", app)
		if err := appStream.AddAudienceWriteEvent(p.conn); err != nil {
			return nil
		}
	}
	return nil
}

func (p *ConnProcessor) getStream(app string, mustExist bool) (*stream.AppStream, error) {
	if tmp, ok := streamMap.Load(app); ok {
		return tmp.(*stream.AppStream), nil
	}
	if mustExist {
		return nil, fmt.Errorf("not exist")
	}
	newApp := stream.NewAppStream()
	streamMap.Store(app, newApp)
	return newApp, nil
}

func (p *ConnProcessor) handshake() error {
	// handshake
	if err := p.conn.HandshakeServer(); err != nil {
		p.conn.NetConn.Close()
		log.Errorf("handshake.HandshakeServer err:%v ", err)
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
