package conn

import (
	"bufio"
	"fmt"
	"github.com/haroldleong/easylive/command"
	"github.com/haroldleong/easylive/util"
	log "github.com/sirupsen/logrus"
	"io"
	"net"

	"github.com/Monibuca/engine/v2/util/bits/pio"
)

const (
	remoteWindowAckSize = 5000000 // 客户端可接受的最大数据包的值
)

type Conn struct {
	NetConn   net.Conn
	bufReader *bufio.Reader // 读数据io
	bufWriter *bufio.Writer // 写数据io
	csidMap   map[uint32]*ChunkStream

	writeMaxChunkSize int
	readMaxChunkSize  int // chunk size

	tmpReadData  []byte // 用于临时读
	tmpWriteData []byte // 用于临时写

	received    uint32
	ackReceived uint32

	messageDone bool // 是否处理完

	ConnInfo *command.ConnectInfo // 连接信息
}

func (c *Conn) MessageDone() bool {
	return c.messageDone
}

func NewConn(netConn net.Conn) *Conn {
	conn := &Conn{
		writeMaxChunkSize: 128,
		readMaxChunkSize:  128,
	}
	conn.NetConn = netConn
	conn.bufWriter = bufio.NewWriterSize(netConn, pio.RecommendBufioSize)
	conn.bufReader = bufio.NewReaderSize(netConn, pio.RecommendBufioSize)
	conn.csidMap = make(map[uint32]*ChunkStream)

	conn.tmpWriteData = make([]byte, 4096)
	conn.tmpReadData = make([]byte, 4096)

	conn.ConnInfo = &command.ConnectInfo{}
	return conn
}

func (c *Conn) HandshakeServer() error {
	var random [(1 + 1536*2) * 2]byte

	C0C1C2 := random[:1536*2+1]
	C0 := C0C1C2[:1]
	C1 := C0C1C2[1 : 1536+1]
	C0C1 := C0C1C2[:1536+1]
	C2 := C0C1C2[1536+1:]

	S0S1S2 := random[1536*2+1:]
	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	S0S1 := S0S1S2[:1536+1]
	S2 := S0S1S2[1536+1:]

	// < C0C1
	if _, err := io.ReadFull(c.bufReader, C0C1); err != nil {
		return err
	}
	if C0[0] != 3 {
		return fmt.Errorf("rtmp: handshake version=%d invalid", C0[0])
	}

	S0[0] = 3

	clitime := pio.U32BE(C1[0:4])
	srvtime := clitime
	srvver := uint32(0x0d0e0a0d)
	cliver := pio.U32BE(C1[4:8])

	if cliver != 0 {
		var ok bool
		var digest []byte
		if ok, digest = hsParse1(C1, hsClientPartialKey, hsServerFullKey); !ok {
			return fmt.Errorf("rtmp: handshake server: C1 invalid")
		}
		hsCreate01(S0S1, srvtime, srvver, hsServerPartialKey)
		hsCreate2(S2, digest)
	} else {
		copy(S1, C1)
		copy(S2, C2)
	}

	// > S0S1S2
	if _, err := c.bufWriter.Write(S0S1S2); err != nil {
		return err
	}
	if err := c.bufWriter.Flush(); err != nil {
		return err
	}

	// < C2
	if _, err := io.ReadFull(c.bufReader, C2); err != nil {
		return err
	}
	return nil
}

func (c *Conn) ReadChunk() (*ChunkStream, error) {
	// 读取basic header
	var (
		data []byte
		err  error
	)
	if data, err = c.readData(1); err != nil {
		return nil, err
	}
	header := data[0]
	format := header >> 6
	csid := uint32(header) & 0x3f
	if csid, err = c.getRealCSID(csid); err != nil {
		return nil, err
	}
	cs, ok := c.csidMap[csid]
	if !ok {
		cs = &ChunkStream{
			CSID: csid,
		}
		c.csidMap[cs.CSID] = cs
	}
	// 读取message header https://github.com/AlexWoo/doc/blob/master/Media/RTMP%20Chunk%20Header.md
	switch format {
	case 0: //	0: Message Header 为 11 字节编码，完整的header，处于流的开头
		var mh []byte
		if mh, err = c.readData(11); err != nil {
			return nil, err
		}
		cs.Format = format
		cs.Timestamp = pio.U24BE(mh[0:3])
		cs.Length = pio.U24BE(mh[3:6])
		cs.TypeID = uint32(mh[6])
		cs.StreamID = pio.U32LE(mh[7:11])
		if cs.Timestamp == 0xffffff {
			if data, err = c.readData(4); err != nil {
				return nil, err
			}
			cs.Timestamp = pio.U32BE(data)
			cs.useExtendTimeStamp = true
		} else {
			cs.useExtendTimeStamp = false
		}
		cs.initData()
	case 1: //1: Message Header 为 7 字节编码，通常在fmt0之后
		var mh []byte
		if mh, err = c.readData(7); err != nil {
			return nil, err
		}
		cs.Format = format
		timestamp := pio.U24BE(mh[0:3])
		cs.Length = pio.U24BE(mh[3:6])
		cs.TypeID = uint32(mh[6])
		if timestamp == 0xffffff {
			if data, err = c.readData(4); err != nil {
				return nil, err
			}
			cs.Timestamp = pio.U32BE(data)
			cs.useExtendTimeStamp = true
		} else {
			cs.useExtendTimeStamp = false
		}
		cs.timeDelta = timestamp
		cs.Timestamp += timestamp
		cs.initData()
	case 2: //2: Message Header 为 3 字节编码，只有一个timestamp delta
		var mh []byte
		if mh, err = c.readData(3); err != nil {
			return nil, err
		}
		cs.Format = format
		timestamp := pio.U24BE(mh[0:3])
		if timestamp == 0xffffff {
			if data, err = c.readData(4); err != nil {
				return nil, err
			}
			cs.Timestamp = pio.U32BE(data)
			cs.useExtendTimeStamp = true
		} else {
			cs.useExtendTimeStamp = false
		}
		cs.timeDelta = timestamp
		cs.Timestamp += timestamp
		cs.initData()
	case 3: //3: Message Header 为 0 字节编码，如果前面一个 chunk 里面存在 timestrameDelta，那么计算 fmt 为 3 的 chunk 时，就直接相加，如果没有，则是使用前一个 chunk 的 timestamp 来进行相加
		if cs.remain == 0 {
			switch cs.Format {
			case 0:
				if cs.useExtendTimeStamp {
					if data, err = c.readData(4); err != nil {
						return nil, err
					}
					cs.Timestamp = pio.U32BE(data)
				}
			case 1, 2:
				var timestamp uint32
				if cs.useExtendTimeStamp {
					if data, err = c.readData(4); err != nil {
						return nil, err
					}
					timestamp = pio.U32BE(data)
				} else {
					timestamp = cs.timeDelta
				}
				cs.Timestamp += timestamp
			}
			cs.initData()
		} else {
			log.Errorf("ReadChunk. cs.remain is not 0,use:%v", cs.useExtendTimeStamp)
		}
	default:
		return nil, fmt.Errorf("invalid format=%d", format)
	}
	log.Debugf("ReadChunk.cs:%s", util.JSON(cs))
	size := int(cs.remain)
	if size > c.readMaxChunkSize {
		size = c.readMaxChunkSize
	}

	buf := cs.Data[cs.index : cs.index+uint32(size)]
	if _, err = io.ReadFull(c.bufReader, buf); err != nil {
		return nil, err
	}

	cs.index += uint32(size)
	cs.remain -= uint32(size)
	if cs.remain == 0 {
		cs.finish = true
	}

	return cs, nil
}

func (c *Conn) Ack(cs *ChunkStream) {
	c.received += cs.Length
	c.ackReceived += cs.Length
	// 处理溢出，acknowledge如果累积超过0xf0000000，就置零
	if c.received >= 0xf0000000 {
		c.received = 0
	}
	if c.ackReceived >= remoteWindowAckSize {
		cs := c.NewAck(c.ackReceived)
		c.writeChunk(&cs, c.writeMaxChunkSize)
		c.ackReceived = 0
	}
}
