package farm_controller

import (
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"time"
)

func (f *FarmController) updateServiceIpAddress(service *corev1.Service, farmIpAddress string) {
	ingressList := []corev1.LoadBalancerIngress{}

	for _, externalIP := range service.Spec.ExternalIPs {
		ingressList = append(ingressList, corev1.LoadBalancerIngress{IP: externalIP})
	}

	ingressList = append(ingressList, corev1.LoadBalancerIngress{IP: farmIpAddress})
	service.Status.LoadBalancer.Ingress = ingressList

	if service.Labels == nil {
		service.Labels = make(map[string]string)
	}

	service.Labels[v1.ServiceStatusLabel] = v1.ServiceStatusLabelSynced
}

func (f *FarmController) updateLabels(farm *v1.Farm, status string) {
	if farm.Labels == nil {
		farm.Labels = make(map[string]string)
	}
	farm.Labels[v1.FarmStatusLabel] = status
	farm.Status.ConnectionStatus = status
	farm.Status.LastUpdate = metav1.NewTime(time.Now())
}

func (f *FarmController) FarmUpdateFailStatus(farm *v1.Farm, eventType, reason, message string) {
	f.Reconcile.Event.Event(farm.DeepCopyObject(), eventType, reason, message)
	f.updateLabels(farm, v1.FarmStatusLabelFailed)
}

func (f *FarmController) FarmUpdateFailDeleteStatus(farm *v1.Farm, eventType, reason, message string) {
	f.Reconcile.Event.Event(farm.DeepCopyObject(), eventType, reason, message)
	f.updateLabels(farm, v1.FarmStatusLabelDeleted)
}

func (f *FarmController) FarmUpdateSuccessStatus(farm *v1.Farm, ipAddress, eventType, reason, message string) {
	f.Reconcile.Event.Event(farm.DeepCopy(), eventType, reason, message)
	f.updateLabels(farm, v1.FarmStatusLabelSynced)
	farm.Status.IpAdress = ipAddress
}

func (f *FarmController) MarkServiceStatusFail(service *corev1.Service, message string) {
	f.Reconcile.Event.Event(service.DeepCopyObject(), "Warning", "FarmCreatedFail", message)
	if service.Labels == nil {
		service.Labels = make(map[string]string)
	}
	service.Labels[v1.ServiceStatusLabel] = v1.ServiceStatusLabelFailed
}

func (f *FarmController) UpdateSuccessEventOnService(service *corev1.Service, message string) {
	f.Reconcile.Event.Event(service.DeepCopyObject(), "Normal", "FarmCreatedSuccess", message)
}
