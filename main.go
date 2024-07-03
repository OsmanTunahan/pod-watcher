package main

import (
	"log"

	"github.com/go-test/deep"
	"github.com/k0kubun/pp"
	v1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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
