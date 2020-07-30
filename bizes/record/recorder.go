package record

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ssh录像：
// https://github.com/asciinema/asciinema

type Env struct {
	Shell string `json:"SHELL"`
	Term string `json:"TERM"`
}

// 头信息
type Header struct {
	Height int `json:"height"`
	Width int `json:"width"`
	Env Env `json:"env"`
	Version int `json:"version"`
	Timestamp int `json:"timestamp"`	// 秒
}

type Recorder struct {
	file *os.File
	timestamp int	// 录屏开始时间(纳秒)
}

func NewRecorder(filename string) (recorder *Recorder, err error) {
	recorder = &Recorder{}

	// 创建目录
	dir := filepath.Dir(filename)
	if err = os.MkdirAll(dir, 0777); err != nil {
		return
	}

	var file *os.File
	if file, err = os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		return
	}

	recorder.file = file
	return
}

func (recorder *Recorder) Close() {
	if recorder.file != nil {
		recorder.file.Close()
	}
}

func (recorder *Recorder) WriteHeader(header *Header) (err error) {
	var s []byte

	if s, err = json.Marshal(header); err != nil {
		return
	}

	recorder.file.Write(s)
	recorder.file.Write([]byte("\n"))

	// 记录开始时间
	recorder.timestamp = header.Timestamp

	return
}

func (recorder *Recorder) WriteData(data string) (err error) {
	now := int(time.Now().UnixNano())

	deltaSeconds := float64(now - recorder.timestamp * 1000 * 1000 * 1000) / 1000 / 1000 / 1000

	row := make([]interface{}, 0)
	row = append(row, deltaSeconds)
	row = append(row, "o")
	row = append(row, data)

	var s []byte
	if s, err = json.Marshal(row); err != nil {
		return
	}
	recorder.file.Write(s)
	recorder.file.Write([]byte("\n"))
	return
}