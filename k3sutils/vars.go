package k3sutils

import (
	"github.com/ipfs/go-log/v2"
	"k8s.io/client-go/kubernetes"
)

var logger = log.Logger("nimp2p-lab")
var EXPERIMENT_PREFIX string = "nimp2p-exp-"
var DEFAULT_APP_CONTAINER_NAME string = "nimp2p"
var NIMP2P_IMAGE string = "katakuri100/nimp2p:v2.0.14"
var HEADLESS_SERVICE_NAME string = "nimp2p-service"
var SIDECAR_CONTAINER_IMAGE string = "prom/blackbox-exporter:v0.27.0"
var SIDECAR_CONTAINER_NAME string = "rtt-exporter"
var DEFAULT_NAMESPACE string = "dst-lab"
var DEFAULT_NIMP2P_PORT = 5000

var K3sClient *kubernetes.Clientset
