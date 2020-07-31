package k8s

import (
	"bytes"
	"crypto/md5"
	"fmt"
	websocket2 "github.com/gorilla/websocket"
	"github.com/owenliang/k8s-jumpserver/bizes/platform"
	"github.com/owenliang/k8s-jumpserver/bizes/protocol"
	"github.com/owenliang/k8s-jumpserver/bizes/record"
	"github.com/owenliang/k8s-jumpserver/bizes/websocket"
	"github.com/owenliang/k8s-jumpserver/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"strconv"
	"strings"
	"time"
)

type websocketProxy struct {
	sessionID int64
	recordFilename string
	tokenAuthData *platform.TokenAuthData
	wsConn *websocket.WsConnection
	resizeEvent chan remotecommand.TerminalSize
	inputBuffer bytes.Buffer	// 输入缓冲区
	recordChan chan string  // 录像数据队列
	closeChan chan byte // 关闭通知
}

func (handler *websocketProxy) init() (err error) {
	handler.recordFilename = handler.generateRecordPath()
	return
}

func (handler *websocketProxy) generateRecordPath() (filename string) {
	// 算会话ID的MD5
	strSessionID := strconv.Itoa(int(handler.sessionID))
	hexMD5 := fmt.Sprintf("%x", md5.Sum([]byte(strSessionID)))

	// 分隔成16级
	paths := make([]string, 0)
	for i := 0; i < 16; i+=2{
		paths = append(paths, string([]byte{hexMD5[i], hexMD5[i+1]}))
	}
	md5Path := strings.Join(paths, "/")

	filename = fmt.Sprintf("%s/%s/%d.cast", config.G_JumpServer.Record.Path, md5Path, handler.sessionID)
	return
}

func (handler *websocketProxy) record() {
	var recorder *record.Recorder
	var err error
	if recorder, err = record.NewRecorder(handler.recordFilename); err != nil {
		return
	}
	defer recorder.Close()

	header := &record.Header{
		Height: 53,
		Width: 210,
		Env: record.Env{Shell: "/bin/bash", Term: "xterm-256color"},
		Version: 2,
		Timestamp: int(time.Now().Unix()),
	}
	if err = recorder.WriteHeader(header); err != nil {
		return
	}

	for {
		select {
		case <- handler.closeChan:
			if len(handler.recordChan)  == 0 {
				goto END
			}
		case data := <- handler.recordChan:
			if err = recorder.WriteData(data); err != nil {
				goto END
			}
		}
	}
END:
}

func (handler *websocketProxy) onLogin() {
	// TODO：根据sessionID,tokenAuthData,recordFilename信息，生成一条ssh会话记录到数据库
}

func (handler *websocketProxy) onLogout() {
	// TODO：在这里更新ssh会话的登出时间字段
}

func (handler *websocketProxy) handleResize(xtermMsg *protocol.XtermMsg) (err error) {
	handler.resizeEvent <- remotecommand.TerminalSize{Width: xtermMsg.Cols, Height: xtermMsg.Rows}
	return
}

func (handler *websocketProxy) handleInput(xtermMsg *protocol.XtermMsg, buf []byte) (size int, err error) {
	// 追加到输入缓冲区
	handler.inputBuffer.Write([]byte(xtermMsg.Input))

	// 向容器copy数据
	size = handler.transInput(buf)
	return
}

func (handler *websocketProxy) transInput(buf []byte) (size int) {
	// 输入缓冲有剩余数据，则推给容器
	if handler.inputBuffer.Len() > 0 {
		size = copy(buf, handler.inputBuffer.Bytes())
		handler.inputBuffer.Next(size)
	}
	return
}

func (handler *websocketProxy) Next() (size *remotecommand.TerminalSize) {
	ret := <- handler.resizeEvent
	size = &ret
	return
}

func (handler *websocketProxy) Read(buf []byte) (size int, err error) {
	// 缓冲区仍有数据, 直接返回
	if size = handler.transInput(buf); size > 0 {
		return
	}

	// 从websocket读取数据
	var wsMsg *websocket.WsMessage
	if wsMsg, err = handler.wsConn.Read(); err != nil {
		return
	}

	var xtermMsg *protocol.XtermMsg
	if xtermMsg, err = protocol.DecodeXtermMsg(wsMsg.Data); err != nil {
		return
	}

	if xtermMsg.MsgType == protocol.XtermMsgTypeResize {	// 窗口缩放事件
		err = handler.handleResize(xtermMsg)
	} else if xtermMsg.MsgType == protocol.XtermMsgTypeInput {	// 输入
		size, err = handler.handleInput(xtermMsg, buf)
	}
	if err != nil {
		handler.wsConn.Close()
	}
	return
}

func (handler *websocketProxy) Write(buf []byte) (size int, err error) {
	var dupBuf = make([]byte, len(buf))
	copy(dupBuf, buf)

	size = len(dupBuf)
	if err = handler.wsConn.Write(websocket2.TextMessage, dupBuf); err != nil {
		handler.wsConn.Close()
		return
	}

	// 将ssh服务端回显数据作为录像内容
	handler.recordChan <- string(dupBuf)
	return
}

func ProxySSHStreaming(tokenAuthData *platform.TokenAuthData, wsConn *websocket.WsConnection, restConf *rest.Config, clientset *kubernetes.Clientset) (err error) {
	// 构造请求URL
	// 长相：https://172.18.11.25:6443/api/v1/namespaces/default/pods/nginx-deployment-5cbd8757f-d5qvx/exec?command=sh&container=nginx&stderr=true&stdin=true&stdout=true&tty=true
	request := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(tokenAuthData.PodName).
		Namespace(tokenAuthData.Namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: tokenAuthData.ContainerName,
			Command:   []string{"bash"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	// 创建SSH连接
	var executor remotecommand.Executor
	if executor, err = remotecommand.NewSPDYExecutor(restConf, "POST", request.URL()); err != nil {
		return
	}

	// 配置转发代理
	var handler = &websocketProxy{
		sessionID: time.Now().UnixNano(),
		tokenAuthData: tokenAuthData,
		wsConn: wsConn,
		resizeEvent: make(chan remotecommand.TerminalSize),
		recordChan: make(chan string, 1024),
		closeChan: make(chan byte, 1),
	}
	if err = handler.init(); err != nil {
		return
	}

	// 登录/登出回调
	handler.onLogin()
	defer handler.onLogout()

	// 录像文件
	go handler.record()

	if err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             handler,
		Stdout:            handler,
		Stderr:            handler,
		TerminalSizeQueue: handler,
		Tty:               true,
	}); err != nil {
		return
	}
	close(handler.closeChan)
	return
}
