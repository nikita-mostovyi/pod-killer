package utils

import (
	corev1 "k8s.io/api/core/v1"
)

var protected = map[string]bool{
	"kube-system": true, "kube-public": true, "kube-node-lease": true,
}

func IsCriticalPod(pod *corev1.Pod) bool {
	return protected[pod.Namespace]
}
