package protocol

import "encoding/json"

// 前端上报JSON消息
type XtermMsg struct {
	MsgType string `json:"type"`	// 类型:resize客户端调整终端, input客户端输入
	Input string `json:"input"`	// msgtype=input情况下使用
	Rows uint16 `json:"rows"`	// msgtype=resize情况下使用
	Cols uint16 `json:"cols"`// msgtype=resize情况下使用
}

const XtermMsgTypeResize = "resize"	// web终端尺寸调整
const XtermMsgTypeInput = "input"		// web终端输入

func DecodeXtermMsg(buffer []byte) (xtermMsg *XtermMsg, err error) {
	xtermMsg = &XtermMsg{}
	if err = json.Unmarshal(buffer, xtermMsg); err != nil {
		return
	}
	return
}