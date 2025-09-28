package observer

import (
	"context"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	killer "pod-killer/internal/service"
	utils "pod-killer/internal/utils"

	"github.com/rs/zerolog/log"
)

// max parallel pods processed at once
const WORKERS_NUM = 10

type ObserverService struct {
	client    *kubernetes.Clientset
	dynClient dynamic.Interface
	killer    *killer.KillerService
	workers   int
}

func NewObserverService(client *kubernetes.Clientset, dynClient dynamic.Interface, killer *killer.KillerService) *ObserverService {
	return &ObserverService{
		client:    client,
		dynClient: dynClient,
		killer:    killer,
		workers:   WORKERS_NUM,
	}
}

func (o *ObserverService) ScanAndKill(ctx context.Context) error {
	pods, err := o.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	log.Info().
		Int("pods_count", len(pods.Items)).
		Str("component", "observer").
		Msg("Scanning pods")

	// semaphore channel to limit concurrency
	sem := make(chan struct{}, o.workers)
	var wg sync.WaitGroup

	for _, pod := range pods.Items {
		pod := pod // capture range variable
		wg.Add(1)

		go func() {
			defer wg.Done()

			// acquire semaphore slot
			sem <- struct{}{}
			defer func() { <-sem }()

			o.processPod(ctx, &pod)
		}()
	}

	wg.Wait()
	return nil
}

func (o *ObserverService) processPod(ctx context.Context, pod *corev1.Pod) {
	if utils.IsCriticalPod(pod) {
		log.Info().Str("namespace", pod.Namespace).Str("pod", pod.Name).
			Msg("Skipping protected namespace")
		return
	}

	isManaged, err := utils.IsManagedByCilium(ctx, o.dynClient, pod)
	if err != nil {
		log.Err(err).Str("name", pod.Name).
			Msg("Failed to retrieve cilium endpoint")
	}

	if isManaged {
		return
	}

	log.Info().Str("namespace", pod.Namespace).Str("name", pod.Name).
		Msg("Evicting unmanaged pod")

	if err := o.killer.DeletePod(ctx, pod); err != nil {
		log.Err(err).Str("namespace", pod.Namespace).Str("name", pod.Name).
			Msg("Failed to evict pod")
	}
}
