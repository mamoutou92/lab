package k3sutils

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DeleteExperiment(expName string) error {
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
	stsClient := K3sClient.AppsV1().StatefulSets(DEFAULT_NAMESPACE)
	if err := stsClient.Delete(ctx, fullName, metav1.DeleteOptions{}); err != nil {

		fmt.Fprintf(tw, "[ERROR] failed to delete experiment '%s': %v\n", expName, err)
		tw.Flush()
		return err
	}
	svcClient := K3sClient.CoreV1().Services(DEFAULT_NAMESPACE)
	if err := svcClient.Delete(ctx, fullName, metav1.DeleteOptions{}); err != nil {

		fmt.Fprintf(tw, "[ERROR] failed to delete service '%s': %v\n", expName, err)
		tw.Flush()
		return err
	}
	fmt.Fprintf(tw, "[INFO] experiment '%s' deleted\n", expName)
	tw.Flush()
	return tw.Flush()
}
