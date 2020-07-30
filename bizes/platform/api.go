package platform

import (
	"context"
)

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