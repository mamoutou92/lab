package k3sutils

import "time"

type StatefulSetRow struct {
	Name      string
	Namespace string
	Scale     int32
	Running   int
	Ready     int32
	Age       time.Duration
	ExpName   string
}

type ExperimentConfig struct {
	ExperimentName string
	NumberPeers    int
	MessageRate    int
	MessageSize    int
	ConnectTo      int
	CpuPerInstance float32
	RamPerInstance int
	Scale          int
	UlBwMbps       int
	DlBwMbps       int
}
