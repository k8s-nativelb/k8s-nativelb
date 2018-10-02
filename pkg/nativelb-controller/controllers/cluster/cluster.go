package cluster_controller

import (
	"context"
	"fmt"
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

//func (c *ClusterController) CreateFarm(farm *v1.Farm,cluster *v1.Cluster) (string, error) {
//	farmIpAddress, err := c.getIpAddrFromAllocator(farm,cluster)
//	if err != nil {
//		log.Log.V(2).Errorf("Fail to allocate Ip address for farm: %s on cluster %s error message: %v", farm.Name, cluster.Name, err)
//		c.clusterUpdateFailStatus(cluster, "Warning", "FarmCreateFail", err.Error())
//		return "", err
//	}
//
//	servers := farm.UpdateServers(cluster.Spec.Internal,farmIpAddress)
//	err = c.UpdateServerData(farm,servers)
//	if err != nil {
//		return "", err
//	}
//
//	err = c.clusterConnection.CreateFarmOnCluster(farm,cluster)
//	if err != nil {
//		log.Log.Errorf("Fail to create farm %s on cluster %s error %v",farm.Name,cluster.Name,err)
//		c.clusterUpdateFailStatus(cluster, "Warning", "FarmCreateFail", err.Error())
//		return "",err
//	}
//
//	c.updateClusterObject(cluster)
//	c.clusterUpdateSuccessStatus(cluster, "Normal", "FarmCreateSuccess", fmt.Sprintf("Farm %s-%s was created on cluster", farm.Namespace, farm.Name))
//	return farmIpAddress, nil
//}

func (c *ClusterController) UpdateFarm(farm *v1.Farm, cluster *v1.Cluster) (string, error) {
	farmIpAddress, err := c.getIpAddrFromAllocator(farm, cluster)
	if err != nil {
		log.Log.V(2).Errorf("Fail to allocate Ip address for farm: %s on cluster %s error message: %v", farm.Name, cluster.Name, err)
		c.clusterUpdateFailStatus(cluster, "Warning", "FarmCreateFail", err.Error())
		return "", err
	}
	farm.Status.IpAdress = farmIpAddress

	servers := farm.UpdateServers(cluster.Spec.Internal, farm.Status.IpAdress)
	err = c.UpdateServerData(farm, servers)
	if err != nil {
		return "", err
	}

	cluster.Status.AllocatedIps[farm.Status.IpAdress] = farm.Name
	c.updateClusterObject(cluster)

	err = c.clusterConnection.UpdateFarmOnCluster(farm, cluster)
	if err != nil {
		log.Log.V(2).Errorf("Fail to update farm: %s on cluster %s error message: %v", farm.FarmName(), cluster.Name, err)
		c.clusterUpdateFailStatus(cluster, "Warning", "FarmUpdateFail", err.Error())
		return "", err
	}

	log.Log.V(2).Infof("successfully updated farm: %s on cluster %s", farm.FarmName(), cluster.Name)
	c.clusterUpdateSuccessStatus(cluster, "Normal", "FarmUpdateSuccess", fmt.Sprintf("Farm %s-%s was updated on cluster", farm.Namespace, farm.Name))
	return farm.Status.IpAdress, nil
}

func (c *ClusterController) DeleteFarm(farm *v1.Farm, cluster *v1.Cluster) error {
	err := c.clusterConnection.DeleteFarmOnCluster(farm, cluster)
	if err != nil {
		log.Log.V(2).Errorf("Fail to remove farm: %s on cluster %s error message: %s", farm.FarmName(), cluster.Name, err.Error())
		c.clusterUpdateFailStatus(cluster, "Warning", "FarmDeleteFail", err.Error())
		return err
	}

	c.allocator[cluster.Name].Release(farm.Status.IpAdress, cluster)
	c.updateClusterObject(cluster)

	log.Log.V(2).Infof("successfully removed farm: %s on cluster %s", farm.FarmName(), cluster.Name)
	c.clusterUpdateSuccessStatus(cluster, "Normal", "FarmDeleteSuccess", fmt.Sprintf("Farm %s-%s was deleted on cluster", farm.Namespace, farm.Name))

	return nil
}

func (c *ClusterController) getIpAddrFromAllocator(farm *v1.Farm, cluster *v1.Cluster) (string, error) {
	_, isExist := c.allocator[cluster.Name]
	if !isExist {
		allocator, err := NewAllocator(cluster)
		if err != nil {
			return "", fmt.Errorf("Fail to create allocator for cluster %s error %v", cluster.Name, err)
		}

		c.allocator[cluster.Name] = allocator
	}

	return c.allocator[cluster.Name].Allocate(farm, cluster)
}

func (c *ClusterController) UpdateServerData(farm *v1.Farm, serverData []v1.ServerData) error {
	farmName := farm.Name
	existServersMap := make(map[string]interface{})

	serverOwnerRef := []metav1.OwnerReference{{Name: farmName, APIVersion: v1.SchemeGroupVersion.Version, Kind: "Farm", UID: farm.UID}}

	labelSelector := labels.Set{}
	labelSelector[v1.NativeLBFarmRef] = farmName

	serversList := &v1.ServerList{}
	err := c.Reconcile.Client.List(context.Background(), &client.ListOptions{LabelSelector: labelSelector.AsSelector()}, serversList)
	if err != nil {
		log.Log.V(2).Errorf("fail to get a list of servers related to %s farm error: %v", farmName, err)
		return err
	}

	for _, server := range serversList.Items {
		existServersMap[server.Name] = struct{}{}
	}

	for _, serverDataObject := range serverData {
		serverName := serverDataObject.Server.Name
		serverObject, err := c.GetServer(serverName)
		if err != nil && errors.IsNotFound(err) {
			serverObject = serverDataObject.Server.DeepCopy()
			serverObject.OwnerReferences = serverOwnerRef
			err := c.Reconcile.Create(context.TODO(), serverObject)
			if err != nil {
				return err
			}

			serverObject, err = c.GetServer(serverName)
			if err != nil {
				return err
			}

		} else if err != nil {
			return err
		} else {
			serverObject.OwnerReferences = serverOwnerRef
			serverObject.Spec = serverDataObject.Server.Spec
			serverObject.Status = serverDataObject.Server.Status

			err := c.Reconcile.Client.Update(context.TODO(), serverObject)
			if err != nil {
				return err
			}
		}

		ownerRef := []metav1.OwnerReference{{Name: serverObject.Name, APIVersion: v1.SchemeGroupVersion.Version, Kind: "Server", UID: serverObject.UID}}
		delete(existServersMap, serverObject.Name)

		existBackendsMap := make(map[string]interface{})

		backendLabelSelector := labels.Set{}
		backendLabelSelector[v1.NativeLBServerRef] = serverObject.Name

		backendsList := &v1.BackendList{}
		err = c.Reconcile.Client.List(context.TODO(), &client.ListOptions{LabelSelector: backendLabelSelector.AsSelector()}, backendsList)
		if err != nil {
			log.Log.V(2).Errorf("fail to get a list of backends related to %s server error: %v", serverObject.Name, err)
			return err
		}

		for _, backend := range backendsList.Items {
			existBackendsMap[backend.Name] = struct{}{}
		}

		for _, backend := range serverDataObject.Backends {
			backendObject, err := c.GetBackend(backend.Name)
			if err != nil && errors.IsNotFound(err) {
				backendObject = backend.DeepCopy()
				backendObject.OwnerReferences = ownerRef
				err := c.Reconcile.Create(context.Background(), backendObject)
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				backendObject.OwnerReferences = ownerRef
				backendObject.Spec = backend.Spec
				backendObject.Status = backend.Status
				err := c.Reconcile.Update(context.Background(), backendObject)
				if err != nil {
					return err
				}
			}

			delete(existBackendsMap, backendObject.Name)
		}

		for deletedBackendName := range existBackendsMap {
			deletedBackend := &v1.Backend{}
			err = c.Reconcile.Get(context.Background(), client.ObjectKey{Name: deletedBackendName, Namespace: v1.ControllerNamespace}, deletedBackend)
			if err != nil && !errors.IsNotFound(err) {
				log.Log.Errorf("Fail to get backend %s for deletion error: %v", deletedBackendName, err)
				return err
			}

			err = c.Reconcile.Delete(context.Background(), deletedBackend)
			if err != nil {
				log.Log.Errorf("Fail to delete backend %s error: %v", deletedBackend.Name, err)
				return err
			}
		}
	}

	for deletedServerName := range existServersMap {
		deletedServer, err := c.GetServer(deletedServerName)
		if err != nil {
			log.Log.Errorf("Fail to get server %s for deletion error: %v", deletedServerName, err)
			return err
		}

		err = c.Reconcile.Delete(context.Background(), deletedServer)
		if err != nil {
			log.Log.Errorf("Fail to delete server %s error: %v", deletedServerName, err)
			return err
		}
	}

	return nil
}

func (c *ClusterController) updateLabels(cluster *v1.Cluster, status string) {
	if cluster.Labels == nil {
		cluster.Labels = make(map[string]string)
	}
	cluster.Status.ConnectionStatus = status
	cluster.Status.LastUpdate = metav1.Now()
	c.Reconcile.Update(context.TODO(), cluster)
}

func (c *ClusterController) clusterUpdateFailStatus(cluster *v1.Cluster, eventType, reason, message string) {
	c.Reconcile.Event.Event(cluster.DeepCopyObject(), eventType, reason, message)
	c.updateLabels(cluster, v1.ClusterConnectionStatusFail)
}

func (c *ClusterController) clusterUpdateSuccessStatus(cluster *v1.Cluster, eventType, reason, message string) {
	c.Reconcile.Event.Event(cluster.DeepCopy(), eventType, reason, message)
	c.updateLabels(cluster, v1.ClusterConnectionStatusSuccess)
}

func (c *ClusterController) updateClusterObject(cluster *v1.Cluster) error {
	err := c.Reconcile.Update(context.Background(), cluster)
	if err != nil {
		log.Log.Errorf("fail to update cluster %s error %v", cluster.Name, err)
		return err
	}

	return nil
}

func (c *ClusterController) GetServer(serverName string) (*v1.Server, error) {
	server := &v1.Server{}
	retry := 5
	var err error

	for i := 0; i < retry; i++ {
		err = c.Reconcile.Get(context.TODO(), client.ObjectKey{Namespace: v1.ControllerNamespace, Name: serverName}, server)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		} else if err == nil {
			return server, nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return nil, err
}

func (c *ClusterController) GetBackend(backendName string) (*v1.Backend, error) {
	backend := &v1.Backend{}
	retry := 5
	var err error

	for i := 0; i < retry; i++ {
		err = c.Reconcile.Get(context.TODO(), client.ObjectKey{Namespace: v1.ControllerNamespace, Name: backendName}, backend)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		} else if err == nil {
			return backend, nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return nil, err
}
