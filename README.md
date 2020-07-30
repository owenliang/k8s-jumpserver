# 开源K8S跳板机  - 轻松接入现有发布系统

## 架构

![架构](https://raw.githubusercontent.com/owenliang/k8s-jumpserver/master/arch.jpeg)

## 关键组件

* xtermjs+websocket实现web终端效果
* asciinema存储格式与录像回放

```
test/index.html提供了xtermjs示例。
test/player.html提供了asciinema示例。
```

## 关键原理

* xtermjs：
    * 上行数据：窗口resize，键盘input。
    * 下行数据：SSH服务端数据流。
* asciinema：
    * 录像写入：将SSH下行单向流量写入文件（用户输入其实也是SSH服务端回显的）。

## 接口定义

```
/ssh?ssh_token=：长连接websocket
/records/play?filename=：下载录像
```

## 启动方法

```
go run app.go -jumpserver ./jumpserver.toml
```

## 对接到发布系统

根据架构图示意，大家需要自定义几个hook点。

bizes/platform/api.go：

```
type TokenAuthData struct {
	Namespace string `json:"namespace"`	// 登录POD的ns
	ClusterCfg string `json:"cluster_cfg"`	// 登录K8S集群的yaml
	PodName string `json:"pod_name"`		// 登录POD的name
	ContainerName string `json:"container_name"`		// 登录POD中的哪个container

	// 其他字段自行扩展
}

func ValidateSSHToken(ctx context.Context, sshToken string) (tokenAuthData *TokenAuthData, err error) {
	// TODO: 在这里调用自建发布系统，完成身份校验，返回TokenAuthData 。 （可以自行扩展TokenAuthData字段，后续可以记录到数据库）
	return
}
```

ssh_token由发布系统生成，用户携带ssh token前来登录，需要回调发布系统进行身份校验，关键信息索取。

---

bizes/k8s/stream.go：

```
func (handler *websocketProxy) onLogin() {
	// TODO：根据sessionID,tokenAuthData,recordFilename信息，生成一条ssh会话记录到数据库
}

func (handler *websocketProxy) onLogout() {
	// TODO：在这里更新ssh会话的登出时间字段
}
```

为了可以在发布系统中进行审计和录像回放，应将本次用户的ssh会话详细信息记入数据库，相关信息可以从handler对象中索取。

这里可以选择回调发布平台进入存储，或者直接存入数据库。

## 生产部署

通过systemd拉起：

```
[Unit]
# 服务描述
Description=jumpserver
# 要求必须执行网络
Requires=network-online.target
# 在网络启动之后启动
After=network-online.target

[Service]
# 简单服务
Type=simple
# 运行用户与用户组
User=root
Group=root
# 进程退出立即重启
Restart=always
# 进程工作目录
WorkingDirectory=/path/to/jumpserver
# 执行命令
ExecStart=/path/to/jumpserver/jumpserver -jumpserver /path/to/jumpserver/jumpserver/jumpserver.toml

[Install]
# 在系统启动后加载UNIT
WantedBy=multi-user.target
```

通过发布平台的nginx统一接入：

```
map $http_upgrade $connection_upgrade {
        default upgrade;
        ''      close;
}
server {
        ....
        location /jumpserver/{
                proxy_pass http://jumpserver部署IP:7000/;
                proxy_http_version 1.1;
                proxy_set_header Upgrade $http_upgrade;
                proxy_set_header Connection $connection_upgrade;
        }
}
```

此时URL将变为：

```
/jumpserver/ssh?ssh_token=：长连接websocket
/jumpserver/records/play?filename=：下载录像
```