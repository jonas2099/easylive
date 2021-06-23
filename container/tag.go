package container

import (
	"fmt"
)

// 参考https://juejin.cn/post/6844903678688624647
const (
	SoundAAC = 10

	AACSequenceHeader = 0
	AACRaw            = 1

	AVCSequenceHeader = 0
	AVCNalu           = 1
	AVCEos            = 2

	// 实际上RTMP的帧类型中，只是区分了IDR帧和非IDR帧，非IDR帧中包含P帧和B帧
	FrameKey   = 1
	FrameInter = 2
	VideoH264  = 7
)

type VideoTagHeader struct {
	/*
		1: keyframe (for AVC, a seekable frame)
		2: inter frame (for AVC, a non- seekable frame)
		3: disposable inter frame (H.263 only)
		4: generated keyframe (reserved for server use only)
		5: video info/command frame
	*/
	FrameType uint8

	/*
		1: JPEG (currently unused)
		2: Sorenson H.263
		3: Screen video
		4: On2 VP6
		5: On2 VP6 with alpha channel
		6: Screen video version 2
		7: AVC
	*/
	CodecID uint8

	/*
		0: AVC sequence header
		1: AVC NALU
		2: AVC end of sequence (lower level NALU sequence ender is not required or supported)
	*/
	AVCPacketType uint8

	CompositionTime int32
}

type AudioTagHeader struct {
	/*
		SoundFormat: UB[4]
		0 = Linear PCM, platform endian
		1 = ADPCM
		2 = MP3
		3 = Linear PCM, little endian
		4 = Nellymoser 16-kHz mono
		5 = Nellymoser 8-kHz mono
		6 = Nellymoser
		7 = G.711 A-law logarithmic PCM
		8 = G.711 mu-law logarithmic PCM
		9 = reserved
		10 = AAC
		11 = Speex
		14 = MP3 8-Khz
		15 = Device-specific sound
		Formats 7, 8, 14, and 15 are reserved for internal use
		AAC is supported in Flash Player 9,0,115,0 and higher.
		Speex is supported in Flash Player 10 and higher.
	*/
	SoundFormat uint8

	/*
		SoundRate: UB[2]
		Sampling rate
		0 = 5.5-kHz For AAC: always 3
		1 = 11-kHz
		2 = 22-kHz
		3 = 44-kHz
	*/
	SoundRate uint8

	/*
		SoundSize: UB[1]
		0 = snd8Bit
		1 = snd16Bit
		Size of each sample.
		This parameter only pertains to uncompressed formats.
		Compressed formats always decode to 16 bits internally
	*/
	SoundSize uint8

	/*
		SoundType: UB[1]
		0 = sndMono
		1 = sndStereo
		Mono or stereo sound For Nellymoser: always 0
		For AAC: always 1
	*/
	SoundType uint8

	/*
		0: AAC sequence header
		1: AAC raw
	*/
	AACPacketType uint8
}

func (vh *VideoTagHeader) parseVideoHeader(b []byte) (n int, err error) {
	if len(b) < n+5 {
		err = fmt.Errorf("invalid videodata len=%d", len(b))
		return
	}
	flags := b[0]
	vh.FrameType = flags >> 4
	vh.CodecID = flags & 0xf
	n++
	if vh.FrameType == FrameInter || vh.FrameType == FrameKey {
		vh.AVCPacketType = b[1]
		for i := 2; i < 5; i++ {
			vh.CompositionTime = vh.CompositionTime<<8 + int32(b[i])
		}
		n += 4
	}
	return
}

func (ah *AudioTagHeader) parseAudioHeader(b []byte) (n int, err error) {
	if len(b) < n+1 {
		err = fmt.Errorf("invalid audiodata len=%d", len(b))
		return
	}
	flags := b[0]
	ah.SoundFormat = flags >> 4
	ah.SoundRate = (flags >> 2) & 0x3
	ah.SoundSize = (flags >> 1) & 0x1
	ah.SoundType = flags & 0x1
	n++
	switch ah.SoundFormat {
	case SoundAAC:
		ah.AACPacketType = b[1]
		n++
	}
	return
}

type Tag struct {
	Data        []byte
	VideoHeader *VideoTagHeader
	AudioHeader *AudioTagHeader
}

func (tag *Tag) IsKeyFrame() bool {
	return tag.VideoHeader.FrameType == FrameKey
}

func (tag *Tag) IsSeq() bool {
	return tag.VideoHeader.FrameType == FrameKey &&
		tag.VideoHeader.AVCPacketType == AVCSequenceHeader
}
