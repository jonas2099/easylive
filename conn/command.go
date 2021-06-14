package conn

import (
	"fmt"
	"github.com/haroldleong/easylive/command"
	"github.com/haroldleong/easylive/format/flv/amf"
	"github.com/haroldleong/easylive/util"
	log "github.com/sirupsen/logrus"
)

const (
	msgTypeIDSetChunkSize     = 1
	msgTypeIDAck              = 3
	msgTypeIDUserControl      = 4
	msgTypeIDWindowAckSize    = 5
	msgTypeIDSetPeerBandwidth = 6

	msgTypeIDAudioMsg       = 8
	msgTypeIDVideoMsg       = 9
	msgTypeIDDataMsgAMF3    = 15
	msgTypeIDCommandMsgAMF3 = 17
	msgTypeIDDataMsgAMF0    = 18
	msgTypeIDCommandMsgAMF0 = 20

	cmdConnect       = "connect"
	cmdFcpublish     = "FCPublish"
	cmdReleaseStream = "releaseStream"
	cmdCreateStream  = "createStream"
	cmdPublish       = "publish"
	cmdFCUnpublish   = "FCUnpublish"
	cmdDeleteStream  = "deleteStream"
	cmdPlay          = "play"
)

func (c *Conn) HandleChunk(cs *ChunkStream) (err error) {
	var cmd *command.Command
	switch cs.TypeID {
	case msgTypeIDCommandMsgAMF3:
		log.Debugf("HandleChunk.type msgTypeIDCommandMsgAMF3")
		if len(cs.Data) < 1 {
			err = fmt.Errorf("rtmp: short packet of CommandMsgAMF3")
			return
		}
		// skip first byte
		if cmd, err = c.handleCommandMsgAMF0(cs.Data[1:]); err != nil {
			return
		}
	case msgTypeIDCommandMsgAMF0:
		log.Debugf("HandleChunk.type msgTypeIDCommandMsgAMF0")
		if cmd, err = c.handleCommandMsgAMF0(cs.Data); err != nil {
			return err
		}
	default:
		log.Warnf("HandleChunk.ignore type.id:%d", cs.TypeID)
	}
	if cmd == nil {
		return fmt.Errorf("no cmd")
	}
	log.Infof("HandleChunk.get cmd.%s", util.JSON(cmd))
	return c.processCMD(cs, cmd)
}

func (c *Conn) processCMD(cs *ChunkStream, cmd *command.Command) error {
	switch cmd.CommandName {
	case cmdConnect:
		if err := c.connect(cmd); err != nil {
			return err
		}
		if err := c.connectResp(cs, cmd); err != nil {
			return err
		}
		log.Infof("processCMD.cmdConnect finish")
	case cmdCreateStream:
		/*if err = connServer.createStream(vs[1:]); err != nil {
			return err
		}
		if err = connServer.createStreamResp(c); err != nil {
			return err
		}*/
		log.Infof("processCMD.cmdCreateStream finish")
	case cmdPublish:
		/*if err = connServer.publishOrPlay(vs[1:]); err != nil {
			return err
		}
		// rtmp适合于flv container.如果是mp4会出错
		if err = connServer.publishResp(c); err != nil {
			return err
		}
		connServer.done = true
		connServer.isPublisher = true*/
		log.Infof("processCMD.cmdPublish finish")
	case cmdPlay:
	case cmdFcpublish:
	case cmdReleaseStream:
	case cmdFCUnpublish:
	case cmdDeleteStream:
	default:
		log.Debug("processCMD.no support command:", cmd.CommandName)
	}
	return nil
}

func (c *Conn) connect(cmd *command.Command) error {
	if app, ok := cmd.CommandObj["app"]; ok {
		c.ConnInfo.App = app.(string)
	}
	if flashVer, ok := cmd.CommandObj["flashVer"]; ok {
		c.ConnInfo.Flashver = flashVer.(string)
	}
	if tcurl, ok := cmd.CommandObj["tcUrl"]; ok {
		c.ConnInfo.TcUrl = tcurl.(string)
	}
	if encoding, ok := cmd.CommandObj["objectEncoding"]; ok {
		c.ConnInfo.ObjectEncoding = int(encoding.(float64))
	}
	return nil
}

func (c *Conn) connectResp(cur *ChunkStream, cmd *command.Command) error {
	cs := c.NewWindowAckSize(2500000)
	c.Write(&cs)
	cs = c.NewSetPeerBandwidth(2500000)
	c.Write(&cs)
	cs = c.NewSetChunkSize(uint32(1024))
	c.Write(&cs)

	resp := make(amf.AMFMap)
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(amf.AMFMap)
	event["level"] = "status"
	event["code"] = "NetConnection.Connect.Success"
	event["description"] = "Connection succeeded."
	event["objectEncoding"] = c.ConnInfo.ObjectEncoding
	return c.writeCommandMsg(cur.CSID, cur.StreamID, "_result", cmd.CommandTransId, resp, event)
}

func (c *Conn) handleCommandMsgAMF0(b []byte) (cmd *command.Command, err error) {
	/*	{
		"GotCommand": true,
		"CommandName": "connect",
		"CommandTransId": 1,
		"CommandObj": {
			"app": "movie",
			"flashVer": "FMLE/3.0 (compatible; Lavf58.76.100)",
			"tcUrl": "rtmp://localhost:1936/movie",
			"type": "nonprivate"
		},
		"CommandParams": []
	}*/
	var name, transid, obj interface{}
	var (
		size int
		n    int
	)

	cmd = &command.Command{}

	if name, size, err = amf.ParseAMF0Val(b[n:]); err != nil {
		return
	}
	n += size
	if transid, size, err = amf.ParseAMF0Val(b[n:]); err != nil {
		return
	}
	n += size
	if obj, size, err = amf.ParseAMF0Val(b[n:]); err != nil {
		return
	}
	n += size

	var ok bool
	if cmd.CommandName, ok = name.(string); !ok {
		err = fmt.Errorf("rtmp: CommandMsgAMF0 command is not string")
		return
	}
	cmd.CommandTransId, _ = transid.(float64)
	cmd.CommandObj, _ = obj.(amf.AMFMap)
	cmd.CommandParams = []interface{}{}

	for n < len(b) {
		if obj, size, err = amf.ParseAMF0Val(b[n:]); err != nil {
			return
		}
		n += size
		cmd.CommandParams = append(cmd.CommandParams, obj)
	}
	if n < len(b) {
		err = fmt.Errorf("rtmp: CommandMsgAMF0 left bytes=%d", len(b)-n)
		return
	}
	cmd.GotCommand = true
	return
}
