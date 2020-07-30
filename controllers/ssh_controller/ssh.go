package ssh_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/owenliang/k8s-jumpserver/bizes/k8s"
	"github.com/owenliang/k8s-jumpserver/bizes/platform"
	"github.com/owenliang/k8s-jumpserver/bizes/websocket"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func SSH(ctx *gin.Context) {
	sshToken := ctx.Query("ssh_token")

	var err error
	var tokenAuthData *platform.TokenAuthData

	// 调用发布系统校验ssh token
	if tokenAuthData, err = platform.ValidateSSHToken(ctx, sshToken); err != nil {
		ctx.Status(403)
		return
	}

	// 建立websocket连接
	var wsConn *websocket.WsConnection
	if wsConn, err = websocket.InitWebsocket(ctx.Writer, ctx.Request); err != nil {
		return
	}

	// 创建K8S客户端
	var restConf *rest.Config
	var clientset *kubernetes.Clientset
	if restConf, clientset, err = k8s.InitClient(tokenAuthData.ClusterCfg); err != nil {
		goto END
	}

	// 建立到容器的ssh代理
	if err = k8s.ProxySSHStreaming(tokenAuthData, wsConn, restConf, clientset); err != nil {
		goto END
	}

END:
	wsConn.Close()
	return
}