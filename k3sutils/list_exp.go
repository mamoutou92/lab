package k3sutils

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func ListAllExperiments(ctx context.Context) error {
	allSts, err := K3sClient.AppsV1().StatefulSets(DEFAULT_NAMESPACE).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list experiments: %v", err)
	}
	var rows []StatefulSetRow
	for _, sts := range allSts.Items {
		labelsMap := sts.GetLabels()
		if !hasPrefixedExperimentLabel(labelsMap, EXPERIMENT_PREFIX) {
			continue
		}

		var scale int32 = 0
		if sts.Spec.Replicas != nil {
			scale = *sts.Spec.Replicas
		}

		sel := ""
		if sts.Spec.Selector != nil {
			ls, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
			if err == nil {
				sel = ls.String()
			}
		}

		podList, err := K3sClient.CoreV1().Pods(sts.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: sel,
		})
		if err != nil {
			return fmt.Errorf("error listing Pods for %s/%s: %v", sts.Namespace, sts.Name, err)
		}
		running := 0
		for _, p := range podList.Items {
			if p.Status.Phase == corev1.PodRunning {
				running++
			}
		}

		ready := sts.Status.ReadyReplicas

		age := time.Since(sts.CreationTimestamp.Time)

		rows = append(rows, StatefulSetRow{
			Name:      sts.Name,
			Namespace: sts.Namespace,
			Scale:     scale,
			Running:   running,
			Ready:     ready,
			Age:       age,
			ExpName:   strings.Split(sts.Name, EXPERIMENT_PREFIX)[1],
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Namespace == rows[j].Namespace {
			return rows[i].Name < rows[j].Name
		}
		return rows[i].Namespace < rows[j].Namespace
	})

	tw := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "EXPERIMENT\tFULLNAME\tSCALE\tRUNNING\tREADY\tAGE")
	for _, r := range rows {
		fmt.Fprintf(
			tw,
			"%s\t%s\t%d\t%d\t%d\t%s\n",
			r.ExpName, r.Name, r.Scale, r.Running, r.Ready, humanDuration(r.Age),
		)
	}
	return tw.Flush()
}

func ReturnAllExperiments() map[string]StatefulSetRow {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	result := make(map[string]StatefulSetRow)
	defer cancel()
	allSts, err := K3sClient.AppsV1().StatefulSets(DEFAULT_NAMESPACE).List(ctx, metav1.ListOptions{})
	if err != nil {
		return result
	}
	var rows []StatefulSetRow
	for _, sts := range allSts.Items {
		labelsMap := sts.GetLabels()
		if !hasPrefixedExperimentLabel(labelsMap, EXPERIMENT_PREFIX) {
			continue
		}

		var scale int32 = 0
		if sts.Spec.Replicas != nil {
			scale = *sts.Spec.Replicas
		}

		sel := ""
		if sts.Spec.Selector != nil {
			ls, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
			if err == nil {
				sel = ls.String()
			}
		}

		podList, err := K3sClient.CoreV1().Pods(sts.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: sel,
		})
		if err != nil {
			continue
		}
		running := 0
		for _, p := range podList.Items {
			if p.Status.Phase == corev1.PodRunning {
				running++
			}
		}

		ready := sts.Status.ReadyReplicas

		age := time.Since(sts.CreationTimestamp.Time)

		result[strings.Split(sts.Name, EXPERIMENT_PREFIX)[1]] = StatefulSetRow{
			Name:      sts.Name,
			Namespace: sts.Namespace,
			Scale:     scale,
			Running:   running,
			Ready:     ready,
			Age:       age,
			ExpName:   strings.Split(sts.Name, EXPERIMENT_PREFIX)[1],
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Namespace == rows[j].Namespace {
			return rows[i].Name < rows[j].Name
		}
		return rows[i].Namespace < rows[j].Namespace
	})
	return result
}

func hasPrefixedExperimentLabel(lbls map[string]string, prefix string) bool {
	if lbls == nil {
		return false
	}
	if v, ok := lbls["app.kubernetes.io/name"]; ok && strings.HasPrefix(v, prefix) {
		return true
	}
	if v, ok := lbls["app.kubernetes.io/component"]; ok && strings.HasPrefix(v, prefix) {
		return true
	}

	return false
}

func getExperimentLabel(lbls map[string]string) string {
	if v, ok := lbls["app.kubernetes.io/name"]; ok && strings.HasPrefix(v, EXPERIMENT_PREFIX) {
		return v
	}
	return ""
}

func mapToSelector(m map[string]string) string {
	return labels.SelectorFromSet(m).String()
}

func humanDuration(d time.Duration) string {

	sec := int64(d.Seconds())
	switch {
	case sec < 60:
		return fmt.Sprintf("%ds", sec)
	case sec < 3600:
		return fmt.Sprintf("%dm%ds", sec/60, sec%60)
	case sec < 86400:
		return fmt.Sprintf("%dh%dm", sec/3600, (sec%3600)/60)
	default:
		days := sec / 86400
		hours := (sec % 86400) / 3600
		return fmt.Sprintf("%dd%dh", days, hours)
	}
}
