package k3sutils

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ScaleExperiment(expName string, scaleNumber int32) error {
	tw := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	expSlice := ReturnAllExperiments()

	if _, ok := expSlice[expName]; !ok {
		fmt.Fprintf(tw, "[ERROR] experiment '%s' not found\n", expName)
		tw.Flush()
		return fmt.Errorf("experiment '%s' not found", expName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fullName := EXPERIMENT_PREFIX + expName
	scaleConfig := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fullName,
			Namespace: DEFAULT_NAMESPACE,
		},
		Spec: autoscalingv1.ScaleSpec{Replicas: scaleNumber},
	}

	stsClient := K3sClient.AppsV1().StatefulSets(DEFAULT_NAMESPACE)
	if _, err := stsClient.UpdateScale(ctx, fullName, scaleConfig, metav1.UpdateOptions{}); err != nil {

		fmt.Fprintf(tw, "[ERROR] failed to scale experiment '%s' tp '%d' peers: %v\n", expName, scaleNumber, err)
		tw.Flush()
		return err
	}
	fmt.Fprintf(tw, "[INFO] experiment '%s' scaled to '%d' peers\n", expName, scaleNumber)
	tw.Flush()
	return tw.Flush()
}
