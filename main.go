package main

import (
	flv "FLV/lib"
	"syscall/js"
)

var (
	Uint8Array = js.Global().Get("Uint8Array")
)

/**
	参数1 id 需要JS生成唯一ID
	参数2 flv buffer 通常是一个完成的Flv封装文件
**/
func loadFlvDemuxer(this js.Value, args []js.Value) interface{} {
	var err error
	if len(args) == 0 {
		return map[string]interface{}{"err": "Missing required arguments"}
	}

	// js.Global().Get("console").Call("log", "id:", js.ValueOf(id))
	streamLen := args[0].Get("byteLength").Int()
	streamBytes := make([]byte, streamLen)
	js.CopyBytesToGo(streamBytes, args[0])

	if err != nil {
		return map[string]interface{}{"err": err}
	}

	version, hasVideo, hasAudio, err := flv.ReadHeader(streamBytes)
	if err != nil {
		return map[string]interface{}{"err": err}
	}

	return map[string]interface{}{
		"err":      js.ValueOf(nil),
		"version":  version,
		"hasVideo": hasVideo,
		"hasAudio": hasAudio,
	}
}

/**
	参数1 id 需要JS生成唯一ID
	参数2 tag buffer 要读取的tag数据 不包含previous 的tag头和data

	读取flv Tag head和body的信息，读取传入的tag数据并返回相应的数据结构
**/
func readTag(this js.Value, args []js.Value) interface{} {
	var err error
	if len(args) == 0 {
		return map[string]interface{}{"err": "Missing required arguments"}
	}

	// js.Global().Get("console").Call("log", "id:", js.ValueOf(id))
	tagLen := args[0].Get("byteLength").Int()
	tagBytes := make([]byte, tagLen)
	js.CopyBytesToGo(tagBytes, args[0])

	tagType, tagSize, timestamp, err := flv.ReadTagHeaderByBytes(tagBytes)
	if err != nil {
		return map[string]interface{}{"err": err.Error()}
	}

	tag := tagBytes[11:tagLen]
	// js.Global().Get("console").Call("log", "tagType:", js.ValueOf(tagType.String()))
	if tagType == 9 {
		frame, err := flv.Decode(tag)
		if err != nil {
			return map[string]interface{}{"err": err.Error()}
		}
		buffer := Uint8Array.New(len(frame.Raw))
		js.CopyBytesToJS(buffer, frame.Raw)
		return map[string]interface{}{
			"err": js.ValueOf(nil),
			"header": map[string]interface{}{
				"tagType":   tagType.String(),
				"tagSize":   tagSize,
				"timestamp": timestamp,
			},
			"body": map[string]interface{}{
				"videocodec": frame.CodecID.String(),
				"frameType":  frame.FrameType.String(),
				"trait":      frame.Trait.String(),
				"cts":        frame.CTS,
				"Raw":        buffer,
			},
		}
	}

	return js.ValueOf(nil)
}

func registerCallbacks() {
	js.Global().Set("loadFlvDemuxer", js.FuncOf(loadFlvDemuxer))
	js.Global().Set("readTag", js.FuncOf(readTag))
}

func main() {
	c := make(chan struct{}, 0)
	registerCallbacks()

	<-c
}
