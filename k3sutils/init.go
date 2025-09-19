package k3sutils

import (
	"github.com/ipfs/go-log/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	log.SetAllLoggers(log.LevelInfo)
	log.SetLogLevel("nimp2p-lab:k3s", "info")
	if cfg, err := rest.InClusterConfig(); err == nil {
		if K3sClient, err = kubernetes.NewForConfig(cfg); err == nil {
			return
		}
	}

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	cfg, err := kubeconfig.ClientConfig()
	if err != nil {
		logger.Fatalf("failed to access Kubeconfig: %v", err)
	}

	if K3sClient, err = kubernetes.NewForConfig(cfg); err != nil {
		logger.Fatalf("failed to create Client Config: %v", err)
	}
}
