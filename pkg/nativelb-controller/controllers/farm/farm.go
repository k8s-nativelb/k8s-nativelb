/*
Copyright 2018 Sebastian Sch.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package farm_controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (f *FarmController) CreateOrUpdateFarm(service *corev1.Service) bool {
	clusterInstance, err := f.clusterController.GetClusterFromService(service)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to find cluster for service %s on namespace %s",
			service.Name,
			service.Namespace)
		return true
	}

	needToCreate := false
	needToUpdate := true
	farm, err := f.Farm(service.Namespace).Get(service.Name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return false
		}

		// Create a new farm object
		farm, err = f.createFarmObject(service, clusterInstance)
		if err != nil {
			return false
		}
		needToCreate = true
	} else {
		needToUpdate, err = f.needToUpdate(farm, service, clusterInstance)
		if err != nil {
			return false
		}
	}
	if needToUpdate {
		err = f.createOrUpdateFarm(farm, service, clusterInstance, needToCreate)
		if err != nil {
			return false
		}
		return true
	}

	return f.needToAddIngressIpFromFarm(service, farm)
}

func (f *FarmController) needToUpdate(farm *v1.Farm, service *corev1.Service, clusterInstance *v1.Cluster) (bool, error) {
	if farm.Spec.Cluster != clusterInstance.Name {
		return true, nil
	}

	if value, ok := farm.Labels[v1.FarmStatusLabel]; !ok {
		return true, nil
	} else if value == v1.FarmStatusLabelFailed {
		return true, nil
	}

	if !reflect.DeepEqual(farm.Spec.Ports, service.Spec.Ports) {
		return true, nil
	}

	endpoints, err := f.getEndpointsList(service, clusterInstance)
	if err != nil {
		f.FarmUpdateFailStatus(farm, "Warning", "FarmCreateFail", fmt.Sprintf("failed to get node list service %s in namespace %s error: %v", service.Name, service.Namespace, err))
		_, err = f.Farm(service.Namespace).Update(farm)
		if err != nil {
			log.Log.Reason(err).Errorf("Fail to update farm error message: %v", err)
		}
		//f.markServiceStatusAsFail(service, fmt.Sprintf("failed to get node lists for service %s on namespace %s error: %v", service.Name, service.Namespace, err))
		return false, err
	}

	if !reflect.DeepEqual(farm.Status.Endpoints, endpoints) {
		return true, nil
	}

	log.Log.Infof("no update needed for service %s on namespace %s", service.Name, service.Namespace)
	return false, nil
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

func (f *FarmController) createOrUpdateFarm(farm *v1.Farm, service *corev1.Service, clusterInstance *v1.Cluster, needToCreate bool) error {
	if farm.Spec.Cluster != clusterInstance.Name {
		err := f.clusterController.DeleteFarm(farm, clusterInstance)
		if err != nil {
			deletedProviderFarm, err := f.createFarmObject(service, clusterInstance)
			if err != nil {
				log.Log.Errorf("fail to create a new farm for a delete service object error %v", err)
			}

			f.FarmUpdateFailDeleteStatus(deletedProviderFarm, "Warning", "FarmDeleteFail", fmt.Sprintf("%v", err))
			deletedProviderFarm, err = f.Farm(service.Namespace).Update(deletedProviderFarm)
			if err != nil {
				log.Log.V(2).Error("Fail to create a new farm for for the deleted farm on cluster")
			}
		}

		farm.Spec.Cluster = clusterInstance.Name
	}

	farm.Spec.Ports = service.Spec.Ports
	endpoints, err := f.getEndpointsList(service, clusterInstance)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to create farm %s on cluster %s error: %v", farm.Name, clusterInstance.Name, err)
		f.FarmUpdateFailStatus(farm, "Warning", "FarmCreateFail", fmt.Sprintf("failed to get node list service %s in namespace %s error: %v", service.Name, service.Namespace, err))
		_, updateErr := f.Farm(service.Namespace).Update(farm)
		if updateErr != nil {
			log.Log.Reason(err).Errorf("Fail to update farm error message: %v", updateErr)
		}
		return err
	}
	farm.Status.Endpoints = endpoints

	if needToCreate {
		err = f.clusterController.CreateFarm(farm, clusterInstance)
		if err != nil {
			f.FarmUpdateFailStatus(farm, "Warning", "FarmCreateFail", fmt.Sprintf("failed to create farm %s on cluster %s error: %v", farm.Name, clusterInstance.Name, err))
		}
	} else {
		err = f.clusterController.UpdateFarm(farm, clusterInstance)
		if err != nil {
			f.FarmUpdateFailStatus(farm, "Warning", "FarmUpdateFail", fmt.Sprintf("failed to update farm %s on cluster %s error: %v", farm.Name, clusterInstance.Name, err))
		}
	}

	farm, errCreateFarm := f.Farm(service.Namespace).Update(farm)
	if errCreateFarm != nil {
		log.Log.Reason(err).Errorf("Fail to update farm error message: %s", errCreateFarm.Error())
	}

	if err != nil {
		return err
	}

	f.FarmUpdateSuccessStatus(farm, "Normal", "FarmUpdate", fmt.Sprintf("Farm updated on cluster %s", farm.Spec.Cluster))
	farm, err = f.Farm(service.Namespace).Update(farm)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to update farm status error message: %v", errCreateFarm)
		return err
	}

	f.updateServiceIpAddress(service, farm.Status.IpAdress)
	log.Log.Infof("Successfully updated the farm %s for service %s on cluster %s", farm.Name, service.Name, clusterInstance.Name)
	return nil
}

func (f *FarmController) DeleteFarm(serviceNamespace, serviceName string) {
	farm, err := f.Farm(serviceNamespace).Get(serviceName)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to find farm %s-%s for deletion", serviceName, serviceNamespace)
		return
	}

	clusterInstance, err := f.Cluster(v1.ControllerNamespace).Get(farm.Spec.Cluster)
	if f.IfFailedUpdateFailDeletedStatus(err, farm, fmt.Sprintf("failed to get cluster %s error %v", farm.Spec.Cluster, err)) {
		return
	}

	err = f.clusterController.DeleteFarm(farm, clusterInstance)
	if f.IfFailedUpdateFailDeletedStatus(err, farm, fmt.Sprintf("failed to delete farm %s on cluster %s error %v", farm.Name, farm.Spec.Cluster, err)) {
		return
	}

	err = f.Farm(serviceNamespace).Delete(farm.Name)
	if f.IfFailedUpdateFailDeletedStatus(err, farm, fmt.Sprintf("failed to delete farm %s from the k8s database error %v", farm.Name, err)) {
		return
	}

	log.Log.Infof("successfully removed the farm %s from cluster %s", farm.Name, farm.Spec.Cluster)
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
		log.Log.V(2).Errorf("failed to get service %s on namespace %s from farm with error message %v", farmInstance.Spec.ServiceNamespace, farmInstance.Spec.ServiceName, err)
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
	resyncTick := time.Tick(v1.ResyncFailFarms * time.Second)

	for range resyncTick {
		// Sync farm need to be deleted
		labelSelector := labels.Set{}
		farmList, err := f.Farm("").List(&client.ListOptions{LabelSelector: labelSelector.AsSelector()})
		if err != nil {
			log.Log.Reason(err).Error("failed to get farm list")
		} else {
			for _, farmInstance := range farmList.Items {
				if value, exist := farmInstance.Labels[v1.FarmStatusLabel]; !exist {
					continue
				} else if value == v1.FarmStatusLabelSynced || value == v1.FarmStatusLabelSyncing {
					continue
				}

				if !f.serviceExist(&farmInstance) {
					f.DeleteFarm(farmInstance.Spec.ServiceNamespace, farmInstance.Spec.ServiceName)
				} else {
					service, err := f.getServiceFromFarm(&farmInstance)
					if err != nil {
						log.Log.Reason(err).Errorf("failed to get service %s on namespace %s error %v", farmInstance.Spec.ServiceNamespace, farmInstance.Spec.ServiceName, err)
					}

					clusterInstance, err := f.Cluster(v1.ControllerNamespace).Get(farmInstance.Spec.Cluster)
					if err != nil {
						log.Log.Reason(err).Errorf("failed to find cluster %s for farm %s error %v", farmInstance.Spec.Cluster, farmInstance.Name, err)
					} else {
						f.createOrUpdateFarm(&farmInstance, service, clusterInstance, false)
					}
				}
			}
		}
	}
}

func (f *FarmController) createFarmObject(service *corev1.Service, cluster *v1.Cluster) (*v1.Farm, error) {
	farmStatus := v1.FarmStatus{Endpoints: make([]string, 0)}
	farmSpec := v1.FarmSpec{Cluster: cluster.Name,
		Ports:            service.Spec.Ports,
		ServiceName:      service.Name,
		ServiceNamespace: service.Namespace, Servers: make(map[string]*v1.ServerSpec)}

	farmObject := &v1.Farm{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{v1.FarmStatusLabel: v1.FarmStatusLabelSyncing, v1.ClusterLabel: cluster.Name},
		Namespace: service.Namespace, Name: service.Name},
		Spec:   farmSpec,
		Status: farmStatus}

	farmObject, err := f.Farm(service.Namespace).Create(farmObject)
	if err != nil {
		return nil, err
	}

	return farmObject, nil
}

func (f *FarmController) getEndpointsList(service *corev1.Service, cluster *v1.Cluster) ([]string, error) {
	if cluster.Spec.Internal == true {
		return f.getEndPoints(service)
	}

	return f.getClusterNodes()
}

func (f *FarmController) reSyncCleanRemovedServices() {
	cleanTick := time.NewTimer(v1.ResyncCleanRemovedServices * time.Second)
	nativeClient := f.GetClient()

	for range cleanTick.C {
		farmList, err := f.Farm("").List(nil)
		if err != nil {
			log.Log.Reason(err).Error("failed to get farm list")
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
