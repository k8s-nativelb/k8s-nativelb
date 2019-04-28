package farm_controller

import (
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"

	corev1 "k8s.io/api/core/v1"
)

func (f *FarmController) updateServiceIpAddress(service *corev1.Service, farmIpAddress string) {
	ingressList := []corev1.LoadBalancerIngress{}

	for _, externalIP := range service.Spec.ExternalIPs {
		ingressList = append(ingressList, corev1.LoadBalancerIngress{IP: externalIP})
	} //func (f *FarmController) markServiceStatusAsFail(service *corev1.Service, message string) {
	//	f.Reconcile.Event.Event(service.DeepCopyObject(), "Warning", "FarmCreatedFail", message)
	//	if service.Labels == nil {
	//		service.Labels = make(map[string]string)
	//	}
	//	service.Labels[v1.ServiceStatusLabel] = v1.ServiceStatusLabelFailed
	//}

	//func (f *FarmController) UpdateSuccessEventOnService(service *corev1.Service, message string) {
	//	f.Reconcile.Event.Event(service.DeepCopyObject(), "Normal", "FarmCreatedSuccess", message)
	//}

	ingressList = append(ingressList, corev1.LoadBalancerIngress{IP: farmIpAddress})
	service.Status.LoadBalancer.Ingress = ingressList
}

func (f *FarmController) updateLabels(farm *v1.Farm, status string) {
	if farm.Labels == nil {
		farm.Labels = make(map[string]string)
	}
	farm.Labels[v1.FarmStatusLabel] = status
}

func (f *FarmController) FarmUpdateFailStatus(farm *v1.Farm, eventType, reason, message string) {
	log.Log.Errorf(message)
	f.Reconcile.Event.Event(farm.DeepCopyObject(), eventType, reason, message)
	f.updateLabels(farm, v1.FarmStatusLabelFailed)
}

func (f *FarmController) IfFailedUpdateFailDeletedStatus(err error, farm *v1.Farm, message string) bool {
	if err != nil {
		log.Log.Reason(err).Errorf(message)
		f.FarmUpdateFailDeleteStatus(farm, "Warning", "FarmDeleteFail", message)
		return true
	}
	return false
}

func (f *FarmController) FarmUpdateFailDeleteStatus(farm *v1.Farm, eventType, reason, message string) {
	var err error
	f.Reconcile.Event.Event(farm.DeepCopyObject(), eventType, reason, message)
	f.updateLabels(farm, v1.FarmStatusLabelDeleted)

	farm, err = f.Farm(farm.Namespace).Update(farm)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to create delete label on farm %s error %v", farm.Name, err)
	}
}

func (f *FarmController) FarmUpdateSuccessStatus(farm *v1.Farm, eventType, reason, message string) {
	log.Log.Infof(message)
	f.Reconcile.Event.Event(farm.DeepCopy(), eventType, reason, message)
	f.updateLabels(farm, v1.FarmStatusLabelSynced)
}
