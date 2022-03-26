package main

import (
	flv "FLV/lib"
	"bytes"
	"syscall/js"
)

type DemuxWorker struct {
	reader        *bytes.Reader
	demuxer       flv.Demuxer
	videoPackager flv.VideoPackager
	audioPackager flv.AudioPackager
}

var (
	Uint8Array                         = js.Global().Get("Uint8Array")
	dws        map[string]*DemuxWorker = make(map[string]*DemuxWorker)
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
	id := args[0].String()
	// js.Global().Get("console").Call("log", "id:", js.ValueOf(id))
	streamLen := args[1].Get("byteLength").Int()
	streamBytes := make([]byte, streamLen)
	js.CopyBytesToGo(streamBytes, args[1])
	reader := bytes.NewReader(streamBytes)
	demuxer, err := flv.NewDemuxer(reader)

	if err != nil {
		return map[string]interface{}{"err": err}
	}

	if _, ok := dws[id]; ok {
		return map[string]interface{}{"err": "Can not found Demuxer"}
	}

	videoPackager, err := flv.NewVideoPackager()
	if err != nil {
		return map[string]interface{}{"err": err}
	}

	audioPackager, err := flv.NewAudioPackager()
	if err != nil {
		return map[string]interface{}{"err": err}
	}

	dws[id] = &DemuxWorker{
		reader:        reader,
		demuxer:       demuxer,
		videoPackager: videoPackager,
		audioPackager: audioPackager,
	}

	version, hasVideo, hasAudio, err := demuxer.ReadHeader()
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
	var ok bool
	var demuxWorker *DemuxWorker
	var err error
	if len(args) == 0 {
		return map[string]interface{}{"err": "Missing required arguments"}
	}

	id := args[0].String()
	if demuxWorker, ok = dws[id]; !ok {
		return map[string]interface{}{"err": "Can not found Demuxer"}
	}

	// js.Global().Get("console").Call("log", "id:", js.ValueOf(id))
	tagLen := args[1].Get("byteLength").Int()
	tagBytes := make([]byte, tagLen)
	js.CopyBytesToGo(tagBytes, args[1])

	tagType, tagSize, timestamp, err := demuxWorker.demuxer.ReadTagHeaderByBytes(tagBytes)
	if err != nil {
		return map[string]interface{}{"err": err.Error()}
	}

	tag := tagBytes[11:tagLen]
	// js.Global().Get("console").Call("log", "tagType:", js.ValueOf(tagType.String()))
	if tagType == 9 {
		frame, err := demuxWorker.videoPackager.Decode(tag)
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
	} else if tagType == 8 {
		frame, err := demuxWorker.audioPackager.Decode(tag)
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
				"audiocodec":      frame.SoundFormat.String(),
				"frameType":       frame.SoundType.String(),
				"trait":           frame.Trait.String(),
				"audioLevel":      frame.AudioLevel,
				"audioSampleBits": frame.SoundSize.String(),
				"Raw":             buffer,
			},
		}
	}
	return js.ValueOf(nil)
}

/**
	参数1 id 需要JS生成唯一ID

	读取flv Tag head和body的信息，每调取一次读去一条（一对）
**/
func read(this js.Value, args []js.Value) interface{} {
	var ok bool
	var demuxWorker *DemuxWorker
	var err error
	if len(args) == 0 {
		return map[string]interface{}{"err": "Missing required arguments"}
	}
	id := args[0].String()

	if demuxWorker, ok = dws[id]; !ok {
		return map[string]interface{}{"err": "Can not found Demuxer"}
	}

	tagType, tagSize, timestamp, err := demuxWorker.demuxer.ReadTagHeader()
	if err != nil {
		return map[string]interface{}{"err": err.Error()}
	}

	tag, err := demuxWorker.demuxer.ReadTag(tagSize)
	if err != nil {
		return map[string]interface{}{"err": err.Error()}
	}
	// js.Global().Get("console").Call("log", "tagType:", js.ValueOf(tagType.String()))
	if tagType == 9 {
		frame, err := demuxWorker.videoPackager.Decode(tag)
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
	} else if tagType == 8 {
		frame, err := demuxWorker.audioPackager.Decode(tag)
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
				"audiocodec":      frame.SoundFormat.String(),
				"frameType":       frame.SoundType.String(),
				"trait":           frame.Trait.String(),
				"audioLevel":      frame.AudioLevel,
				"audioSampleBits": frame.SoundSize.String(),
				"Raw":             buffer,
			},
		}
	}
	return js.ValueOf(nil)
}

/**
	参数1 id 需要JS生成唯一ID

	关闭当前flv工作区，读取完成后执行
**/
func cancelDemuxer(this js.Value, args []js.Value) interface{} {
	var ok bool
	var demuxWorker *DemuxWorker
	if len(args) == 0 {
		return map[string]interface{}{"err": "Can not found id"}
	}
	id := args[0].String()

	if demuxWorker, ok = dws[id]; !ok {
		return map[string]interface{}{"err": "Can not found Demuxer"}
	}
	demuxWorker.reader = nil
	demuxWorker.demuxer.Close()
	demuxWorker = nil

	return js.ValueOf(nil)
}

func registerCallbacks() {
	js.Global().Set("loadFlvDemuxer", js.FuncOf(loadFlvDemuxer))
	js.Global().Set("cancelDemuxer", js.FuncOf(cancelDemuxer))
	js.Global().Set("read", js.FuncOf(read))
	js.Global().Set("readTag", js.FuncOf(readTag))
}

func main() {
	c := make(chan struct{}, 0)
	registerCallbacks()

	<-c
}
