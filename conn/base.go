package conn

import (
	"bytes"
	"encoding/binary"
	"fmt"
	newamf "github.com/gwuhaolin/livego/protocol/amf"
	"github.com/haroldleong/easylive/consts"
	"github.com/haroldleong/easylive/format/flv/amf"
	"github.com/haroldleong/easylive/util"
	log "github.com/sirupsen/logrus"
	"io"
)

func (c *Conn) readData(n int32) ([]byte, error) {
	mh := c.tmpReadData[:n]
	if _, err := io.ReadFull(c.bufReader, mh); err != nil {
		return nil, err
	}
	return mh, nil
}

func (c *Conn) WriteChunk(cs *ChunkStream) error {
	if cs.TypeID == consts.MsgTypeIDDataMsgAMF0 ||
		cs.TypeID == consts.MsgTypeIDDataMsgAMF3 {
		var err error
		if cs.Data, err = newamf.MetaDataReform(cs.Data, newamf.DEL); err != nil {
			return err
		}
		cs.Length = uint32(len(cs.Data))
	}
	if err := c.Write(cs); err != nil {
		return err
	}
	return c.bufWriter.Flush()
}

func (c *Conn) Write(cs *ChunkStream) error {
	if cs.TypeID == pcmSetChunkSize {
		c.writeMaxChunkSize = int(binary.BigEndian.Uint32(cs.Data))
	}
	return c.writeChunk(cs, c.writeMaxChunkSize)
}

func (c *Conn) writeChunk(cs *ChunkStream, chunkSize int) error {
	if cs.TypeID == consts.MsgTypeIDAudioMsg {
		cs.CSID = 4
	} else if cs.TypeID == consts.MsgTypeIDVideoMsg ||
		cs.TypeID == consts.MsgTypeIDDataMsgAMF0 ||
		cs.TypeID == consts.MsgTypeIDDataMsgAMF3 {
		cs.CSID = 6
	}

	totalLen := uint32(0)
	numChunks := cs.Length / uint32(chunkSize)
	for i := uint32(0); i <= numChunks; i++ {
		if totalLen == cs.Length {
			break
		}
		if i == 0 {
			cs.Format = uint8(0)
		} else {
			cs.Format = uint8(3)
		}
		if err := c.writeHeader(cs); err != nil {
			return err
		}
		inc := uint32(chunkSize)
		start := i * uint32(chunkSize)
		if uint32(len(cs.Data))-start <= inc {
			inc = uint32(len(cs.Data)) - start
		}
		totalLen += inc
		end := start + inc
		buf := cs.Data[start:end]
		if _, err := c.bufWriter.Write(buf); err != nil {
			return err
		}
	}

	return nil
}

func (c *Conn) writeHeader(cs *ChunkStream) error {
	//Chunk Basic Header
	h := cs.Format << 6
	switch {
	case cs.CSID < 64:
		h |= uint8(cs.CSID)
		c.WriteUintBE(uint32(h), 1)
	case cs.CSID-64 < 256:
		h |= 0
		c.WriteUintBE(uint32(h), 1)
		c.WriteUintLE(cs.CSID-64, 1)
	case cs.CSID-64 < 65536:
		h |= 1
		c.WriteUintBE(uint32(h), 1)
		c.WriteUintLE(cs.CSID-64, 2)
	}
	//Chunk Message Header
	ts := cs.Timestamp
	if cs.Format == 3 {
		goto END
	}
	if cs.Timestamp > consts.FlvTimestampMax {
		ts = consts.FlvTimestampMax
	}
	c.WriteUintBE(ts, 3)
	if cs.Format == 2 {
		goto END
	}
	if cs.Length > consts.FlvTimestampMax {
		return fmt.Errorf("length=%d", cs.Length)
	}
	c.WriteUintBE(cs.Length, 3)
	c.WriteUintBE(cs.TypeID, 1)
	if cs.Format == 1 {
		goto END
	}
	c.WriteUintLE(cs.StreamID, 4)
END:
	//Extended Timestamp
	if ts >= consts.FlvTimestampMax {
		c.WriteUintBE(cs.Timestamp, 4)
	}
	return nil
}

func (c *Conn) writeCommandMsg(csid, msgsid uint32, args ...interface{}) (err error) {
	return c.writeAMF0Msg(consts.MsgTypeIDCommandMsgAMF0, csid, msgsid, args...)
}

func (c *Conn) writeAMF0Msg(typeID uint32, csid, streamID uint32, args ...interface{}) error {
	encoder := &newamf.Encoder{}
	byteWriter := bytes.NewBuffer(nil)
	for _, v := range args {
		if _, err := encoder.Encode(byteWriter, v, newamf.AMF0); err != nil {
			return err
		}
	}
	cs := ChunkStream{
		Format:    0,
		CSID:      csid,
		Timestamp: 0,
		TypeID:    typeID,
		StreamID:  streamID,
		Length:    uint32(len(byteWriter.Bytes())),
		Data:      byteWriter.Bytes(),
	}
	if err := c.Write(&cs); err != nil {
		return err
	}
	return c.bufWriter.Flush()
}

func (c *Conn) userControlMsg(eventType, buflen uint32) ChunkStream {
	var ret ChunkStream
	buflen += 2
	ret = ChunkStream{
		Format:   0,
		CSID:     2,
		TypeID:   4,
		StreamID: 1,
		Length:   buflen,
		Data:     make([]byte, buflen),
	}
	ret.Data[0] = byte(eventType >> 8 & 0xff)
	ret.Data[1] = byte(eventType & 0xff)
	return ret
}

func (c *Conn) handleCommandMsgAMF0(b []byte) (cmd *Command, err error) {
	// 命令解析详见https://www.jianshu.com/p/7dd3b5b4e092
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

	cmd = &Command{}

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

	r := bytes.NewReader(b)
	vs, _ := DecodeBatch(r)
	log.Errorf("DecodeBatch.vs:%v", util.JSON(vs))
	log.Errorf("DecodeBatch.cmd:%v", util.JSON(cmd))

	return
}

func DecodeBatch(r io.Reader) (ret []interface{}, err error) {
	var v interface{}
	d := &newamf.Decoder{}
	for {
		v, err = d.Decode(r, newamf.AMF0)
		if err != nil {
			break
		}
		ret = append(ret, v)
	}
	return
}
