package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-test/deep"
	"github.com/k0kubun/pp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type Logger interface {
	Info(msg string)
	Debug(msg string, obj interface{})
	Error(err error)
	DetailsEnabled() bool
}

type DefaultLogger struct {
	details bool
}

func (l *DefaultLogger) Info(msg string) {
	log.Println(msg)
}

func (l *DefaultLogger) Debug(msg string, obj interface{}) {
	if l.details {
		pp.Println(obj)
	}
}

func (l *DefaultLogger) Error(err error) {
	log.Println(err)
}

func (l *DefaultLogger) DetailsEnabled() bool {
	return l.details
}

type PodEventHandler struct {
	logger Logger
}

func (h *PodEventHandler) OnAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	h.logger.Info("Pod created: " + pod.ObjectMeta.Name)
	h.logger.Debug("Pod details: ", pod)
}

func (h *PodEventHandler) OnDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	h.logger.Info("Pod deleted: " + pod.ObjectMeta.Name)
	h.logger.Debug("Pod details: ", pod)
}

func (h *PodEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)
	h.logger.Info("Pod updated: " + oldPod.ObjectMeta.Name)
	if h.logger.DetailsEnabled() {
		if diff := deep.Equal(oldPod, newPod); diff != nil {
			h.logger.Debug("Difference: ", diff)
		} else {
			h.logger.Info("No difference, just a cache update")
		}
	}
}

type KubernetesClient interface {
	CoreV1() kubernetes.Interface
}

type ConfigLoader struct{}

func (c *ConfigLoader) LoadConfig(kubeconfigPath string) (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

type PodWatcher struct {
	client   cache.Getter
	handlers cache.ResourceEventHandlerFuncs
}

func (pw *PodWatcher) Watch(namespace, selector string, resyncPeriod time.Duration) cache.Store {
	optionsModifier := func(options *metav1.ListOptions) {
		options.LabelSelector = selector
	}
	lw := cache.NewFilteredListWatchFromClient(pw.client, v1.ResourcePods.String(), namespace, optionsModifier)
	store, controller := cache.NewInformer(lw, &v1.Pod{}, resyncPeriod, pw.handlers)
	forever := make(chan struct{})
	go controller.Run(forever)
	return store
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	namespace := flag.String("namespace", metav1.NamespaceAll, "namespace to watch")
	details := flag.Bool("details", false, "print pod object details")
	labelSelector := labels.Set(map[string]string{"foo": "bar", "baz": "quux"}).AsSelector()
	selector := flag.String("selector", labelSelector.String(), "selector (label query) to filter on")

	flag.Parse()

	configLoader := &ConfigLoader{}
	config, err := configLoader.LoadConfig(*kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	logger := &DefaultLogger{details: *details}
	eventHandler := &PodEventHandler{logger: logger}
	podWatcher := &PodWatcher{
		client: clientset.CoreV1().RESTClient(),
		handlers: cache.ResourceEventHandlerFuncs{
			AddFunc:    eventHandler.OnAdd,
			DeleteFunc: eventHandler.OnDelete,
			UpdateFunc: eventHandler.OnUpdate,
		},
	}

	podWatcher.Watch(*namespace, *selector, 5*time.Minute)

	select {}
}
