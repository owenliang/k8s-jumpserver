package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/owenliang/k8s-jumpserver/config"
	"github.com/owenliang/k8s-jumpserver/controllers/records"
	"github.com/owenliang/k8s-jumpserver/controllers/ssh_controller"
	_ "go.uber.org/automaxprocs"
)

var (
	jumpserver *string
)

func init() {
	jumpserver = flag.String("jumpserver", "./jumpserver.toml", "业务配置文件")
}

func main() {
	flag.Parse()

	var err error
	var gin = gin.New()

	if err = config.LoadConfig(jumpserver); err != nil {
		goto FAIL
	}

	gin.GET("/ssh", ssh_controller.SSH)
	gin.GET("/records/play", records.Play)

	gin.Run(config.G_JumpServer.Server.Listen)
	return

FAIL:
	fmt.Println(err)
}
