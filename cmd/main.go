package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"path/filepath"
	"syscall"
	"time"

	"pod-killer/internal/api"

	observer "pod-killer/internal/observer"
	killer "pod-killer/internal/service"

	"github.com/rs/zerolog/log"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	// "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		<-ch
		cancel()
	}()

	// Start API server for health & metrics
	go api.StartServer()

	// Kuberntes out of cluster client
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "~/Users/nikita/.kube/config")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "~/Users/nikita/.kube/config")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Dynamic client for CiliumEndpoints
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create dynamic client")
	}

	// Kubernetes in-cluster client
	// cfg, err := rest.InClusterConfig()
	// if err != nil {
	// 	log.Err(err).Msg("Failed to load in-cluster config")
	// }
	// clientset, err := kubernetes.NewForConfig(cfg)
	// if err != nil {
	// 	log.Err(err).Msg("Failed to create k8s client")
	// }

	killerSvc := killer.NewKillerService(clientset)
	observerSvc := observer.NewObserverService(clientset, dynClient, killerSvc)

	log.Info().Msg("Pod-Killer started")

	// Start main loop
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Shutting down pod-killer...")
			return
		case <-ticker.C:
			if err := observerSvc.ScanAndKill(ctx); err != nil {
				log.Err(err).Msg("Error during scan")
			}
		}
	}
}
