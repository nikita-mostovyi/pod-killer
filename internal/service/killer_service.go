package service

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/rs/zerolog/log"
)

type KillerService struct {
	client *kubernetes.Clientset
}

func NewKillerService(client *kubernetes.Clientset) *KillerService {
	return &KillerService{client: client}
}

func (k *KillerService) DeletePod(ctx context.Context, pod *corev1.Pod) error {
	// Build an eviction request
	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
	}

	// Try evicting the pod (respects PDBs)
	err := k.client.PolicyV1().Evictions(pod.Namespace).Evict(ctx, eviction)
	if err != nil {
		log.Error().
			Err(err).
			Str("namespace", pod.Namespace).
			Str("pod", pod.Name).
			Msg("Failed to delete (possibly due to PDB)")
		return err
	}

	log.Info().
		Str("namespace", pod.Namespace).
		Str("pod", pod.Name).
		Msg("Successfully deleted pod")
	return nil
}
