package conn

const (
	// flv格式通常只有8、9、18常用的type
	TAG_AUDIO = 8
	TAG_VIDEO = 9
	// 除了音视频数据外还有 ScriptData，这是一种类似二进制json的对象描述数据格式，JavaScript比较惨只能自己写实现，其它平台可以用 librtmp的代码去做
	TAG_SCRIPTDATAAMF0 = 18
	TAG_SCRIPTDATAAMF3 = 0xf
)

const FlvTimestampMax = 0xFFFFFF
const chunkHeaderLength = 12
