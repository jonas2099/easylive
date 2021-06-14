package conn

import (
	"encoding/binary"
	"fmt"
	"github.com/haroldleong/easylive/format/flv/amf"
	"io"
)

func (c *Conn) readData(n int32) ([]byte, error) {
	mh := c.tmpReadData[:n]
	if _, err := io.ReadFull(c.bufReader, mh); err != nil {
		return nil, err
	}
	return mh, nil
}

func (c *Conn) Write(cs *ChunkStream) error {
	if cs.TypeID == pcmSetChunkSize {
		c.writeMaxChunkSize = int(binary.BigEndian.Uint32(cs.Data))
	}
	return c.writeChunk(cs, c.writeMaxChunkSize)
}

func (c *Conn) writeChunk(cs *ChunkStream, chunkSize int) error {
	if cs.TypeID == TAG_AUDIO {
		cs.CSID = 4
	} else if cs.TypeID == TAG_VIDEO ||
		cs.TypeID == TAG_SCRIPTDATAAMF0 ||
		cs.TypeID == TAG_SCRIPTDATAAMF3 {
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
	if cs.Timestamp > FlvTimestampMax {
		ts = FlvTimestampMax
	}
	c.WriteUintBE(ts, 3)
	if cs.Format == 2 {
		goto END
	}
	if cs.Length > FlvTimestampMax {
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
	if ts >= FlvTimestampMax {
		c.WriteUintBE(cs.Timestamp, 4)
	}
	return nil
}

func (c *Conn) writeCommandMsg(csid, msgsid uint32, args ...interface{}) (err error) {
	return c.writeAMF0Msg(msgTypeIDCommandMsgAMF0, csid, msgsid, args...)
}

func (c *Conn) writeAMF0Msg(msgtypeid uint8, csid, msgsid uint32, args ...interface{}) (err error) {
	size := 0
	for _, arg := range args {
		size += amf.LenAMF0Val(arg)
	}
	if len(c.tmpWriteData) < chunkHeaderLength+size {
		c.tmpWriteData = make([]byte, chunkHeaderLength+size)
	}
	for _, arg := range args {
		amf.FillAMF0Val(c.tmpWriteData, arg)
	}
	cs := ChunkStream{
		Format:    0,
		CSID:      csid,
		Timestamp: 0,
		TypeID:    uint32(msgtypeid),
		StreamID:  msgsid,
		Length:    uint32(len(c.tmpWriteData)),
		Data:      c.tmpWriteData,
	}
	c.Write(&cs)
	return
}
