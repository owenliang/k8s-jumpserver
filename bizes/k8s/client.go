package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func InitClient(clusterYAML string) (restConf *rest.Config, clientset *kubernetes.Clientset, err error) {
	// 解析配置
	if restConf, err = clientcmd.RESTConfigFromKubeConfig([]byte(clusterYAML)); err != nil {
		return
	}
	// 创建客户端
	if clientset, err = kubernetes.NewForConfig(restConf); err != nil {
		return
	}
	return
}