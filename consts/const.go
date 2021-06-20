package consts

const (
	FlvTimestampMax   = 0xFFFFFF
	chunkHeaderLength = 12
)

type UserControl uint32

const (
	UserControlStreamBegin      UserControl = 0
	UserControlStreamEOF        UserControl = 1
	UserControlStreamDry        UserControl = 2
	UserControlSetBufferLen     UserControl = 3
	UserControlStreamIsRecorded UserControl = 4
	UserControlPingRequest      UserControl = 6
	UserControlPingResponse     UserControl = 7
)

/*

Command Message(命令消息，Message Type ID＝17或20)：表示在客户端盒服务器间传递的在对端执行某些操作的命令消息，如connect表示连接对端，对端如果同意连接的话会记录发送端信息并返回连接成功消息，publish表示开始向对方推流，接受端接到命令后准备好接受对端发送的流信息，后面会对比较常见的Command Message具体介绍。当信息使用AMF0编码时，Message Type ID＝20，AMF3编码时Message Type ID＝17.
Data Message（数据消息，Message Type ID＝15或18）：传递一些元数据（MetaData，比如视频名，分辨率等等）或者用户自定义的一些消息。当信息使用AMF0编码时，Message Type ID＝18，AMF3编码时Message Type ID＝15.
Shared Object Message(共享消息，Message Type ID＝16或19)：表示一个Flash类型的对象，由键值对的集合组成，用于多客户端，多实例时使用。当信息使用AMF0编码时，Message Type ID＝19，AMF3编码时Message Type ID＝16.
Audio Message（音频信息，Message Type ID＝8）：音频数据。
Video Message（视频信息，Message Type ID＝9）：视频数据。
Aggregate Message (聚集信息，Message Type ID＝22)：多个RTMP子消息的集合
User Control Message Events(用户控制消息，Message Type ID=4):告知对方执行该信息中包含的用户控制事件，比如Stream Begin事件告知对方流信息开始传输。和前面提到的协议控制信息（Protocol Control Message）不同，这是在RTMP协议层的，而不是在RTMP chunk流协议层的，这个很容易弄混。该信息在chunk流中发送时，Message Stream ID=0,Chunk Stream Id=2,Message Type Id=4。
*/
const (
	MsgTypeIDSetChunkSize     = 1
	MsgTypeIDAck              = 3
	MsgTypeIDUserControl      = 4
	MsgTypeIDWindowAckSize    = 5
	MsgTypeIDSetPeerBandwidth = 6

	MsgTypeIDAudioMsg       = 8
	MsgTypeIDVideoMsg       = 9
	MsgTypeIDDataMsgAMF3    = 15 //AMF3编码，音视频metaData，传递一些元数据比如视频名，分辨率等
	MsgTypeIDCommandMsgAMF3 = 17 //AMF3编码，RTMP命令消息
	MsgTypeIDDataMsgAMF0    = 18 //AMF0编码，音视频metaData风格
	MsgTypeIDCommandMsgAMF0 = 20 //AMF0编码，RTMP命令消息

	CmdConnect         = "connect"
	CmdFcpublish       = "FCPublish"
	CmdReleaseStream   = "releaseStream"
	CmdCreateStream    = "createStream"
	CmdPublish         = "publish"
	CmdFCUnPublish     = "FCUnpublish"
	CmdDeleteStream    = "deleteStream"
	CmdGetStreamLength = "getStreamLength"
	CmdPlay            = "play"
)
