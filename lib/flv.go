// The MIT License (MIT)
//
// Copyright (c) 2013-2017 Oryx(ossrs)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// The oryx FLV package support bytes from/to FLV tags.
package flv

import (
	"bytes"
	"errors"
)

// FLV Tag Type is the type of tag,
// refer to @doc video_file_format_spec_v10.pdf, @page 9, @section FLV tags
type TagType uint8

const (
	TagTypeForbidden  TagType = 0
	TagTypeAudio      TagType = 8
	TagTypeVideo      TagType = 9
	TagTypeScriptData TagType = 18
)

func (v TagType) String() string {
	switch v {
	case TagTypeVideo:
		return "Video"
	case TagTypeAudio:
		return "Audio"
	case TagTypeScriptData:
		return "Data"
	default:
		return "Forbidden"
	}
}

// FLV Demuxer is used to demux FLV file.
// Refer to @doc video_file_format_spec_v10.pdf, @page 74, @section Annex E. The FLV File Format
// A FLV file must consist the bellow parts:
//	1. A FLV header, refer to @doc video_file_format_spec_v10.pdf, @page 8, @section The FLV header
//	2. One or more tags, refer to @doc video_file_format_spec_v10.pdf, @page 9, @section FLV tags
// @remark We always ignore the previous tag size.
type Demuxer interface {
	// Read the FLV header, return the version of FLV, whether hasVideo or hasAudio in header.
	ReadHeader() (version uint8, hasVideo, hasAudio bool, err error)
	// Read the FLV tag header, return the tag information, especially the tag size,
	// then user can read the tag payload.
	// Compare with ReadTagHeader, the only difference is the index of the parameter.
	ReadTagHeaderByBytes(p []byte) (tagType TagType, tagSize, timestamp uint32, err error)
	// Read the FLV tag header, return the tag information, especially the tag size,
	// then user can read the tag payload.
	ReadTagHeader() (tagType TagType, tagSize, timestamp uint32, err error)
	// Read the FLV tag body, drop the next 4 bytes previous tag size.
	ReadTag(tagSize uint32) (tag []byte, err error)
	// Close the demuxer.
	Close() error
}

// When FLV signature is not "FLV"
var errSignature = errors.New("FLV signatures are illegal")

func ReadHeader(head []byte) (version uint8, hasVideo, hasAudio bool, err error) {

	p := head

	if !bytes.Equal([]byte{byte('F'), byte('L'), byte('V')}, p[:3]) {
		err = errSignature
		return
	}

	version = uint8(p[3])
	hasVideo = (p[4] & 0x01) == 0x01
	hasAudio = ((p[4] >> 2) & 0x01) == 0x01

	return
}

func ReadTagHeaderByBytes(p []byte) (tagType TagType, tagSize uint32, timestamp uint32, err error) {
	tagType = TagType(p[0])
	tagSize = uint32(p[1])<<16 | uint32(p[2])<<8 | uint32(p[3])
	timestamp = uint32(p[7])<<24 | uint32(p[4])<<16 | uint32(p[5])<<8 | uint32(p[6])

	return
}

// The video frame type.
// Refer to @doc video_file_format_spec_v10.pdf, @page 78, @section E.4.3 Video Tags
type VideoFrameType uint8

const (
	VideoFrameTypeForbidden  VideoFrameType = iota
	VideoFrameTypeKeyframe                  //  1 = key frame (for AVC, a seekable frame)
	VideoFrameTypeInterframe                // 2 = inter frame (for AVC, a non-seekable frame)
	VideoFrameTypeDisposable                // 3 = disposable inter frame (H.263 only)
	VideoFrameTypeGenerated                 // 4 = generated key frame (reserved for server use only)
	VideoFrameTypeInfo                      // 5 = video info/command frame
)

func (v VideoFrameType) String() string {
	switch v {
	case VideoFrameTypeKeyframe:
		return "Keyframe"
	case VideoFrameTypeInterframe:
		return "Interframe"
	case VideoFrameTypeDisposable:
		return "DisposableInterframe"
	case VideoFrameTypeGenerated:
		return "GeneratedKeyframe"
	case VideoFrameTypeInfo:
		return "Info"
	default:
		return "Forbidden"
	}
}

// The video codec id.
// Refer to @doc video_file_format_spec_v10.pdf, @page 78, @section E.4.3 Video Tags
// It's 4bits, that is 0-16.
type VideoCodec uint8

const (
	VideoCodecForbidden   VideoCodec = iota + 1
	VideoCodecH263                   // 2 = Sorenson H.263
	VideoCodecScreen                 // 3 = Screen video
	VideoCodecOn2VP6                 // 4 = On2 VP6
	VideoCodecOn2VP6Alpha            // 5 = On2 VP6 with alpha channel
	VideoCodecScreen2                // 6 = Screen video version 2
	VideoCodecAVC                    // 7 = AVC
	// See page 79 at @doc https://github.com/CDN-Union/H265/blob/master/Document/video_file_format_spec_v10_1_ksyun_20170615.doc
	VideoCodecHEVC VideoCodec = 12 // 12 = HEVC
)

func (v VideoCodec) String() string {
	switch v {
	case VideoCodecH263:
		return "H.263"
	case VideoCodecScreen:
		return "Screen"
	case VideoCodecOn2VP6:
		return "VP6"
	case VideoCodecOn2VP6Alpha:
		return "On2VP6(alpha)"
	case VideoCodecScreen2:
		return "Screen2"
	case VideoCodecAVC:
		return "AVC"
	case VideoCodecHEVC:
		return "HEVC"
	default:
		return "Forbidden"
	}
}

// The video AVC frame trait, whethere sequence header or not.
// Refer to @doc video_file_format_spec_v10.pdf, @page 78, @section E.4.3 Video Tags
// If AVC or HEVC, it's 8bits.
type VideoFrameTrait uint8

const (
	VideoFrameTraitSequenceHeader VideoFrameTrait = iota // 0 = AVC/HEVC sequence header
	VideoFrameTraitNALU                                  // 1 = AVC/HEVC NALU
	VideoFrameTraitSequenceEOF                           // 2 = AVC/HEVC end of sequence (lower level NALU sequence ender is
	VideoFrameTraitForbidden
)

func (v VideoFrameTrait) String() string {
	switch v {
	case VideoFrameTraitSequenceHeader:
		return "SequenceHeader"
	case VideoFrameTraitNALU:
		return "NALU"
	case VideoFrameTraitSequenceEOF:
		return "SequenceEOF"
	default:
		return "Forbidden"
	}
}

type VideoFrame struct {
	CodecID   VideoCodec
	FrameType VideoFrameType
	Trait     VideoFrameTrait
	CTS       int32
	Raw       []byte
}

func NewVideoFrame() *VideoFrame {
	return &VideoFrame{}
}

// The packager used to codec the FLV video tag body.
// Refer to @doc video_file_format_spec_v10.pdf, @page 78, @section E.4.3 Video Tags
type VideoPackager interface {
	// Decode the FLV video tag to video frame.
	// @remark For RTMP/FLV: pts = dts + cts, where dts is timestamp in packet/tag.
	Decode(tag []byte) (frame *VideoFrame, err error)
	// Encode the video frame to FLV video tag.
	Encode(frame *VideoFrame) (tag []byte, err error)
}

type videoPackager struct {
}

var errDataNotEnough = errors.New("Data not enough")

func Decode(tag []byte) (frame *VideoFrame, err error) {
	if len(tag) < 5 {
		err = errDataNotEnough
		return
	}

	p := tag
	frame = &VideoFrame{}
	frame.FrameType = VideoFrameType(byte(p[0]>>4) & 0x0f)
	frame.CodecID = VideoCodec(byte(p[0]) & 0x0f)

	if frame.CodecID == VideoCodecAVC || frame.CodecID == VideoCodecHEVC {
		frame.Trait = VideoFrameTrait(p[1])
		frame.CTS = int32(uint32(p[2])<<16 | uint32(p[3])<<8 | uint32(p[4]))
		frame.Raw = tag[5:]
	} else {
		frame.Raw = tag[1:]
	}

	return
}

func (v videoPackager) Encode(frame *VideoFrame) (tag []byte, err error) {
	if frame.CodecID == VideoCodecAVC || frame.CodecID == VideoCodecHEVC {
		return append([]byte{
			byte(frame.FrameType)<<4 | byte(frame.CodecID), byte(frame.Trait),
			byte(frame.CTS >> 16), byte(frame.CTS >> 8), byte(frame.CTS),
		}, frame.Raw...), nil
	} else {
		return append([]byte{
			byte(frame.FrameType)<<4 | byte(frame.CodecID),
		}, frame.Raw...), nil
	}
}
