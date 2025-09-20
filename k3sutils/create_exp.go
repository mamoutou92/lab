package k3sutils

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func CreateNewExperiment(config ExperimentConfig) error {
	tw := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	expSlice := ReturnAllExperiments()

	if _, ok := expSlice[config.ExperimentName]; ok {
		fmt.Fprintf(tw, "[ERROR] experiment '%s' already exist\n", config.ExperimentName)
		tw.Flush()
		return fmt.Errorf("experiment '%s' already exist", config.ExperimentName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fullName := EXPERIMENT_PREFIX + config.ExperimentName
	labels := map[string]string{
		"app":                         DEFAULT_APP_CONTAINER_NAME,
		"app.kubernetes.io/name":      fullName,
		"app.kubernetes.io/component": fullName,
		"app.kubernetes.io/startdate": time.Now().Format("2006-01-02-15-04-05"),
	}
	restartAlways := corev1.ContainerRestartPolicyAlways

	// create dedicated Headless service for peer discovery

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fullName,
			Namespace: DEFAULT_NAMESPACE,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Selector:  labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "nimp2p",
					Port:       5000,
					TargetPort: intstr.FromInt(5000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	if _, err := K3sClient.CoreV1().Services(DEFAULT_NAMESPACE).Get(ctx, fullName, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			if _, err := K3sClient.CoreV1().Services(DEFAULT_NAMESPACE).Create(ctx, svc, metav1.CreateOptions{}); err != nil {
				fmt.Fprintf(tw, "[ERROR] failed to create headless service '%s': %v\n", fullName, err)
				tw.Flush()
				return err
			}
			fmt.Fprintf(tw, "[INFO] headless service '%s' created\n", fullName)
			tw.Flush()
		} else {
			fmt.Fprintf(tw, "[ERROR] failed to get service: %v\n", err)
			tw.Flush()
			return err
		}
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fullName,
			Namespace: DEFAULT_NAMESPACE,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: fullName,
			Replicas:    int32ptr(int32(config.NumberPeers)),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"kubernetes.io/egress-bandwidth":  fmt.Sprintf("%dM", config.UlBwMbps),
						"kubernetes.io/ingress-bandwidth": fmt.Sprintf("%dM", config.DlBwMbps),
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "lab-blackbox-config",
									},
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:          SIDECAR_CONTAINER_NAME,
							Image:         SIDECAR_CONTAINER_IMAGE,
							RestartPolicy: &restartAlways,
							Args: []string{
								"--config.file=/etc/blackbox/config.yml",
							},
							Ports: []corev1.ContainerPort{{
								Name:          "sidecar-port",
								ContainerPort: 9115,
								Protocol:      corev1.ProtocolTCP,
							}},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("0.05"),
									corev1.ResourceMemory: resource.MustParse("32Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("0.1"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/etc/blackbox/config.yml",
									SubPath:   "config.yml",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  DEFAULT_APP_CONTAINER_NAME,
							Image: NIMP2P_IMAGE,
							Ports: []corev1.ContainerPort{{
								ContainerPort: int32(DEFAULT_NIMP2P_PORT), Protocol: corev1.ProtocolTCP,
							}},
							Env: []corev1.EnvVar{
								{Name: "PEERS", Value: fmt.Sprintf("%d", config.NumberPeers)},
								{Name: "MSGRATE", Value: fmt.Sprintf("%d", config.MessageRate)},
								{Name: "MSGSIZE", Value: fmt.Sprintf("%d", config.MessageSize)},
								{Name: "CONNECTTO", Value: fmt.Sprintf("%d", config.ConnectTo)},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%v", config.CpuPerInstance)),
									corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%vMi", config.RamPerInstance)),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%v", config.CpuPerInstance)),
									corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%vMi", config.RamPerInstance)),
								},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}
	sts.Spec.PodManagementPolicy = appsv1.ParallelPodManagement
	stsClient := K3sClient.AppsV1().StatefulSets(DEFAULT_NAMESPACE)
	if _, err := stsClient.Create(ctx, sts, metav1.CreateOptions{}); err != nil {

		fmt.Fprintf(tw, "[ERROR] failed to create experiment '%s': %v\n", config.ExperimentName, err)
		tw.Flush()
		return err
	}
	fmt.Fprintf(tw, "[INFO] experiment '%s' created\n", config.ExperimentName)
	tw.Flush()
	return tw.Flush()
}

func int32ptr(i int32) *int32 { return &i }
