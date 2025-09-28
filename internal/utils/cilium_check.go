package utils

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var ciliumEndpointGVR = schema.GroupVersionResource{
	Group:    "cilium.io",
	Version:  "v2",
	Resource: "ciliumendpoints",
}

func IsManagedByCilium(ctx context.Context, dynClient dynamic.Interface, pod *corev1.Pod) (bool, error) {
	// Get cilium endpoint by pod name and namespace
	_, err := dynClient.Resource(ciliumEndpointGVR).
		Namespace(pod.Namespace).
		Get(ctx, pod.Name, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
