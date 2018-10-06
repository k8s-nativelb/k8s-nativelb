package farm_controller

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
)

func (f *FarmController) CreateOrUpdateFarm(service *corev1.Service, endpoints *corev1.Endpoints) bool {
	clusterInstance, err := f.getCluster(service)
	if err != nil {
		log.Log.V(2).Errorf("Fail to find cluster for service %s on namespace %s",
			service.Name,
			service.Namespace)
		f.MarkServiceStatusFail(service, "Fail to find a cluster for the service")
		return true
	}

	if clusterInstance.Spec.Internal && endpoints == nil {
		return false
	}

	farmName := fmt.Sprintf("%s-%s",
		service.Namespace,
		service.Name)

	farm, err := f.Farm().Get(farmName)
	if err != nil {
		if !errors.IsNotFound(err) {
			f.MarkServiceStatusFail(service, fmt.Sprintf("Fail to get farm object for service %s on namespace %s", service.Name, service.Namespace))
			return true
		}

		farm, err = f.createFarmObject(service, farmName, clusterInstance)
		if err != nil {
			f.MarkServiceStatusFail(service, fmt.Sprintf("Fail to create farm error: %v", err))
			return true
		}
	}

	needToUpdate, err := f.needToUpdate(farm, service)
	if err != nil {
		return true
	}

	if needToUpdate {
		f.updateFarm(farm, service, clusterInstance)
		return true
	}

	return f.needToAddIngressIpFromFarm(service, farm)
}

func (f *FarmController) needToAddIngressIpFromFarm(service *corev1.Service, farm *v1.Farm) bool {
	ingressList := []corev1.LoadBalancerIngress{}

	for _, externalIP := range service.Spec.ExternalIPs {
		ingressList = append(ingressList, corev1.LoadBalancerIngress{IP: externalIP})
	}

	ingressList = append(ingressList, corev1.LoadBalancerIngress{IP: farm.Status.IpAdress})

	if !reflect.DeepEqual(ingressList, service.Status.LoadBalancer.Ingress) {
		service.Status.LoadBalancer.Ingress = ingressList
		return true
	}
	return false
}

//func (f *FarmController) createFarm(service *corev1.Service) {
//	farmName := fmt.Sprintf("%s-%s", service.Namespace, service.Name)
//	log.Log.V(2).Infof("Start creating a farm object for service %s on namespace %s with farm name %s",service.Name,service.Namespace,farmName)
//	clusterInstance, err := f.getCluster(service)
//	if err != nil {
//		log.Log.V(2).Errorf("Fail to find cluster for service %s on namespace %s", service.Name, service.Namespace)
//		f.MarkServiceStatusFail(service, "Fail to find a clusterfor the service")
//		return
//	}
//
//	farm, err := f.createFarmObject(service, farmName, clusterInstance)
//	if err != nil {
//		f.MarkServiceStatusFail(service, fmt.Sprintf("Fail to create farm error: %v", err))
//		return
//	}
//
//	if len(farm.Status.NodeList) == 0 {
//		log.Log.V(2).Infof("No servers found for service %s on namespace %s",service.Name,service.Namespace)
//		return
//	}
//
//	log.Log.V(2).Infof("Start creating a farm on cluster agents for service %s on namespace %s with farm name %s",service.Name,service.Namespace,farmName)
//	farmIpAddress, err := f.clusterController.CreateFarm(farm,clusterInstance)
//	if err != nil {
//		f.MarkServiceStatusFail(service, fmt.Sprintf("Fail to create farm on cluster error: %s", err.Error()))
//		return
//	}
//	log.Log.V(2).Infof("Done creating a farm on cluster agents for service %s on namespace %s with farm name %s",service.Name,service.Namespace,farmName)
//	errCreateFarm := f.Reconcile.Client.Create(context.Background(), farm)
//	if errCreateFarm != nil {
//		log.Log.V(2).Errorf("Fail to create farm error message: %s", errCreateFarm.Error())
//		f.MarkServiceStatusFail(service, fmt.Sprintf("Fail to create farm error message: %s", errCreateFarm.Error()))
//	}
//
//	if err != nil {
//		log.Log.V(2).Errorf("Fail to create farm on cluster %s error message: %s", farm.Spec.Cluster, errCreateFarm.Error())
//		f.FarmUpdateFailStatus(farm, "Warning", "FarmCreatedFail", err.Error())
//	}
//
//	f.FarmUpdateSuccessStatus(farm, farmIpAddress, "Normal", "FarmCreated", fmt.Sprintf("Farm created on cluster %s", farm.Spec.Cluster))
//	err = f.Reconcile.Client.Update(context.Background(), farm)
//	if err != nil {
//		log.Log.V(2).Errorf("Fail to update farm status error message: %s", errCreateFarm.Error())
//		return
//	}
//
//	f.updateServiceIpAddress(service, farmIpAddress)
//	log.Log.Infof("Successfully created the farm %s for service %s on namespace %s on cluster %s", farm.Name, service.Name,service.Namespace, clusterInstance.Name)
//}

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

func (f *FarmController) updateFarm(farm *v1.Farm, service *corev1.Service, clusterInstance *v1.Cluster) {
	// TODO: Need to change this!!!!
	if farm.Spec.Cluster != clusterInstance.Name {
		err := f.clusterController.DeleteFarm(farm, clusterInstance)
		if err != nil {
			// TODO: Change this use deepCopy
			deletedProviderFarm, err := f.createFarmObject(service,
				fmt.Sprintf("%s-%s-%s", service.Namespace,
					clusterInstance.Name,
					service.Name), clusterInstance)
			if err != nil {
				log.Log.Errorf("fail to create a new farm for a delete service object error %v", err)
			}

			f.FarmUpdateFailDeleteStatus(deletedProviderFarm, "Warning", "FarmDeleteFail", err.Error())
			deletedProviderFarm, err = f.Farm().Update(deletedProviderFarm)
			if err != nil {
				log.Log.V(2).Error("Fail to create a new farm for for the deleted farm on cluster")
			}
		}

		delete(service.Labels, v1.ServiceStatusLabel)
		f.Farm().Delete(farm.Name)
		//f.createFarm(service)
		return
	}

	farm.Spec.Ports = service.Spec.Ports
	nodelist, err := f.getNodeList(service, clusterInstance)
	if err != nil {
		f.MarkServiceStatusFail(service, fmt.Sprintf("Fail to get node list service %s in namespace %s error: %v", service.Name, service.Namespace, err))
		return
	}

	if len(nodelist) == 0 {
		f.DeleteFarm(service.Namespace, service.Name)
		return
	}

	farm.Status.NodeList = nodelist
	farmIpAddress, err := f.clusterController.UpdateFarm(farm, clusterInstance)
	if err != nil {
		f.MarkServiceStatusFail(service, fmt.Sprintf("Fail to update farm on cluster error: %s", err.Error()))
		return
	}

	farm, errCreateFarm := f.Farm().Update(farm)
	if errCreateFarm != nil {
		log.Log.V(2).Errorf("Fail to update farm error message: %s", errCreateFarm.Error())
		f.MarkServiceStatusFail(service, fmt.Sprintf("Fail to update farm error message: %s", errCreateFarm.Error()))
	}

	if err != nil {
		log.Log.V(2).Errorf("Fail to update farm  on cluster %s error message: %s", farm.Spec.Cluster, errCreateFarm.Error())
		f.FarmUpdateFailStatus(farm, "Warning", "FarmUpdateFail", err.Error())
	}

	f.FarmUpdateSuccessStatus(farm, farmIpAddress, "Normal", "FarmUpdate", fmt.Sprintf("Farm updated on cluster %s", farm.Spec.Cluster))
	farm, err = f.Farm().Update(farm)
	if err != nil {
		log.Log.V(2).Errorf("Fail to update farm status error message: %s", errCreateFarm.Error())
		return
	}

	f.updateServiceIpAddress(service, farmIpAddress)
	log.Log.Infof("Successfully updated the farm %s for service %s on cluster %s", farm.Name, service.Name, clusterInstance.Name)
}

func (f *FarmController) DeleteFarm(serviceNamespace, serviceName string) {
	farm, err := f.Farm().Get(fmt.Sprintf("%s-%s", serviceNamespace, serviceName))
	if err != nil {
		log.Log.V(2).Errorf("Fail to find farm %s-%s for deletion", serviceName, serviceNamespace)
		return
	}

	clusterInstance, err := f.Cluster().Get(farm.Spec.Cluster)
	if err != nil {
		log.Log.V(2).Errorf("Fail to get cluster %s error message: %s", farm.Spec.Cluster, err.Error())
		f.FarmUpdateFailDeleteStatus(farm, "Warning", "FarmDeleteFail", err.Error())
		farm, err = f.Farm().Update(farm)
		if err != nil {
			log.Log.V(2).Errorf("Fail to update delete label on farm %s", farm.Name)
		}

		return
	}

	err = f.clusterController.DeleteFarm(farm, clusterInstance)
	if err != nil {
		log.Log.V(2).Errorf("Fail to delete farm on cluster %s error message: %s", farm.Spec.Cluster, err.Error())
		f.FarmUpdateFailDeleteStatus(farm, "Warning", "FarmDeleteFail", err.Error())
		farm, err = f.Farm().Update(farm)
		if err != nil {
			log.Log.V(2).Errorf("Fail to update delete label on farm %s", farm.Name)
		}

		return
	}

	err = f.Farm().Delete(farm.Name)
	if err != nil {
		log.Log.V(2).Errorf("Fail to delete farm %s", farm.Name)
	}
}

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

func (f *FarmController) needToUpdate(farm *v1.Farm, service *corev1.Service) (bool, error) {
	clusterInstance, err := f.getCluster(service)
	if err != nil {
		f.MarkServiceStatusFail(service, fmt.Sprintf("Fail to get provider for service %s in namespace %s error: %v", service.Name, service.Namespace, err))
		return false, err
	}

	if farm.Spec.Cluster != clusterInstance.Name {
		return true, nil
	}

	if value, ok := service.Labels[v1.ServiceStatusLabel]; !ok {
		return true, nil
	} else if value == v1.ServiceStatusLabelFailed {
		return true, nil
	}

	if !reflect.DeepEqual(farm.Spec.Ports, service.Spec.Ports) {
		return true, nil
	}

	nodeList, err := f.getNodeList(service, clusterInstance)
	if err != nil {
		f.MarkServiceStatusFail(service, fmt.Sprintf("Fail to get node lists for service %s in namespace %s error: %v", service.Name, service.Namespace, err))
		return false, err
	}

	if !reflect.DeepEqual(farm.Status.NodeList, nodeList) {
		return true, nil
	}

	return false, nil
}

func (f *FarmController) getServiceFromFarm(farmInstance *v1.Farm) (*corev1.Service, error) {
	nativeClient := f.GetClient()
	service := &corev1.Service{}
	err := nativeClient.Get(context.Background(), client.ObjectKey{Namespace: farmInstance.Spec.ServiceNamespace, Name: farmInstance.Spec.ServiceName}, service)
	return service, err
}

func (f *FarmController) serviceExist(farmInstance *v1.Farm) bool {
	_, err := f.getServiceFromFarm(farmInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			return false
		}
		log.Log.V(2).Errorf("fail to get service %s on namespace %s from farm with error message %s", farmInstance.Spec.ServiceNamespace, farmInstance.Spec.ServiceName, err.Error())
	}

	return true
}

func (f *FarmController) getEndPoints(service *corev1.Service) ([]string, error) {
	nativeClient := f.GetClient()
	endpointsList := make([]string, 0)
	endpoints := &corev1.Endpoints{}
	err := nativeClient.Get(context.Background(), client.ObjectKey{Namespace: service.Namespace, Name: service.Name}, endpoints)
	if err != nil {
		return endpointsList, err
	}

	if len(endpoints.Subsets) == 0 {
		return endpointsList, nil
	}

	for _, endpointValue := range endpoints.Subsets[0].Addresses {
		endpointsList = append(endpointsList, endpointValue.IP)
	}

	return endpointsList, nil
}

func (f *FarmController) getClusterNodes() ([]string, error) {
	nativeClient := f.GetClient()
	nodeList := make([]string, 0)
	nodes := &corev1.NodeList{}
	err := nativeClient.List(context.Background(), &client.ListOptions{}, nodes)
	if err != nil {
		return nodeList, err
	}

	for _, nodeInstance := range nodes.Items {
		for _, IpAddr := range nodeInstance.Status.Addresses {
			if IpAddr.Type == "InternalIP" {
				nodeList = append(nodeList, IpAddr.Address)
			}
		}
	}

	return nodeList, nil
}

func (f *FarmController) reSyncFailFarms() {
	resyncTick := time.Tick(120 * time.Second)

	for range resyncTick {
		// Sync farm need to be deleted
		labelSelector := labels.Set{}
		labelSelector[v1.FarmStatusLabel] = v1.FarmStatusLabelDeleted
		farmList, err := f.Farm().List(&client.ListOptions{LabelSelector: labelSelector.AsSelector()})
		if err != nil {
			log.Log.V(2).Error("reSyncProcess: Fail to get farm list")
		} else {
			for _, farmInstance := range farmList.Items {
				if !f.serviceExist(&farmInstance) {
					f.DeleteFarm(farmInstance.Spec.ServiceNamespace, farmInstance.Spec.ServiceName)
				} else {
					service, err := f.getServiceFromFarm(&farmInstance)
					if err != nil {
						log.Log.V(2).Errorf("fail to get service %s on namespace %s from farm with error message %s", farmInstance.Spec.ServiceNamespace, farmInstance.Spec.ServiceName, err.Error())
					}

					clusterInstance, err := f.Cluster().Get(farmInstance.Spec.Cluster)
					if err != nil {
						log.Log.V(2).Errorf("fail to find cluster object %s for farm %s", farmInstance.Spec.Cluster, farmInstance.Name)
					} else {
						f.updateFarm(&farmInstance, service, clusterInstance)
					}
				}
			}
		}
	}
}

func (f *FarmController) CleanRemovedServices() {
	cleanTick := time.NewTimer(10 * time.Minute)
	nativeClient := f.GetClient()

	for range cleanTick.C {
		farmList, err := f.Farm().List(nil)
		if err != nil {
			log.Log.V(2).Error("CleanRemovedServices: Fail to get farm list")
		} else {
			service := &corev1.Service{}
			for _, farmInstance := range farmList.Items {
				err := nativeClient.Get(context.Background(), client.ObjectKey{Name: farmInstance.Spec.ServiceName, Namespace: farmInstance.Spec.ServiceNamespace}, service)
				if err != nil && errors.IsNotFound(err) {
					f.DeleteFarm(farmInstance.Spec.ServiceNamespace, farmInstance.Spec.ServiceName)
				}
			}
		}
	}
}

func (f *FarmController) getCluster(service *corev1.Service) (*v1.Cluster, error) {
	var clusterInstance *v1.Cluster

	var err error

	if value, ok := service.ObjectMeta.Annotations[v1.NativeLBAnnotationKey]; ok {
		clusterInstance, err = f.Cluster().Get(value)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, fmt.Errorf("Provider Not found for service : %s", service.Name)
			}
			return nil, err
		}

	} else {
		labelSelector := labels.Set{}
		labelSelector[v1.NativeLBDefaultLabel] = "true"
		clusterList, err := f.Cluster().List(&client.ListOptions{LabelSelector: labelSelector.AsSelector()})
		if err != nil {
			return nil, err
		}

		if len(clusterList.Items) == 0 {
			return nil, fmt.Errorf("Default provider not found")
		} else if len(clusterList.Items) > 1 {
			return nil, fmt.Errorf("More then one default provider found")
		}

		clusterInstance = &clusterList.Items[0]
	}

	return clusterInstance, nil
}

func (f *FarmController) createFarmObject(service *corev1.Service, farmName string, cluster *v1.Cluster) (*v1.Farm, error) {
	farmStatus := v1.FarmStatus{NodeList: make([]string, 0), ConnectionStatus: v1.FarmStatusLabelSynced, LastUpdate: metav1.Now()}
	farmSpec := v1.FarmSpec{Cluster: cluster.Name,
		Ports:            service.Spec.Ports,
		ServiceName:      service.Name,
		ServiceNamespace: service.Namespace, Servers: make(map[string]*v1.ServerSpec)}

	farmObject := &v1.Farm{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{v1.ServiceStatusLabel: v1.ServiceStatusLabelSyncing},
		Namespace: v1.ControllerNamespace, Name: farmName},
		Spec:   farmSpec,
		Status: farmStatus}

	farmObject, err := f.Farm().Create(farmObject)
	if err != nil {
		return nil, err
	}

	return farmObject, nil
}

func (f *FarmController) getNodeList(service *corev1.Service, cluster *v1.Cluster) ([]string, error) {
	if cluster.Spec.Internal == true {
		return f.getEndPoints(service)
	}

	return f.getClusterNodes()
}
