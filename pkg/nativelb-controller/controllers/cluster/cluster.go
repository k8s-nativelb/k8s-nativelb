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

package cluster_controller

import (
	"fmt"
	"time"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *ClusterController) CreateFarm(farm *v1.Farm, cluster *v1.Cluster) error {
	err := c.allocateIpAddrAndRouterID(farm, cluster)
	if c.ifUpdateFailedUpdateFailStatus(err, cluster, "FarmCreatedFail", fmt.Sprintf("failed to allocate IP address for farm %s on cluster %s error %v", farm.Name, cluster.Name, err)) {
		return err
	}

	servers := farm.UpdateServers(cluster.Spec.Internal)
	err = c.UpdateServerData(farm, servers)
	if err != nil {
		return err
	}

	cluster.Status.AllocatedIps[farm.Status.IpAdress] = farm.Name
	err = c.clusterConnection.CreateFarmOnCluster(farm, cluster)
	if c.ifUpdateFailedUpdateFailStatus(err, cluster, "FarmCreatedFail", fmt.Sprintf("failed to create farm %s on cluster %s error %v", farm.Name, cluster.Name, err)) {
		return err
	}

	c.updateSuccessStatus(cluster, "Normal", "FarmCreateSuccess", fmt.Sprintf("Farm %s was created on cluster", farm.FarmName()))
	return nil
}

func (c *ClusterController) UpdateFarm(farm *v1.Farm, cluster *v1.Cluster) error {
	err := c.allocateIpAddrAndRouterID(farm, cluster)
	if c.ifUpdateFailedUpdateFailStatus(err, cluster, "FarmUpdatedFail", fmt.Sprintf("failed to get allocated IP address for farm %s on cluster %s error %v", farm.Name, cluster.Name, err)) {
		return err
	}

	servers := farm.UpdateServers(cluster.Spec.Internal)
	err = c.UpdateServerData(farm, servers)
	if err != nil {
		return err
	}

	err = c.clusterConnection.UpdateFarmOnCluster(farm, cluster)
	if c.ifUpdateFailedUpdateFailStatus(err, cluster, "FarmUpdatedFail", fmt.Sprintf("failed to update farm %s on cluster %s error %v", farm.Name, cluster.Name, err)) {
		return err
	}

	c.updateSuccessStatus(cluster, "Normal", "FarmUpdateSuccess", fmt.Sprintf("Farm %s was updated on cluster", farm.FarmName()))
	return nil
}

func (c *ClusterController) DeleteFarm(farm *v1.Farm, cluster *v1.Cluster) error {
	err := c.clusterConnection.DeleteFarmOnCluster(farm, cluster)
	if c.ifUpdateFailedUpdateFailStatus(err, cluster, "FarmDeletedFail", fmt.Sprintf("failed to remove farm %s on cluster %s error %v", farm.FarmName(), cluster.Name, err)) {
		return err
	}

	err = c.releaseIpAddrAndRouterID(farm, cluster)
	if err != nil {
		log.Log.V(2).Errorf("failed to release allocated IP address for farm %s on cluster %s error %v", farm.Name, cluster.Name, err)
		return err
	}

	c.updateSuccessStatus(cluster, "Normal", "FarmDeleteSuccess", fmt.Sprintf("the farm %s was deleted on cluster", farm.FarmName()))
	return nil
}

func (c *ClusterController) GetClusterFromService(service *corev1.Service) (*v1.Cluster, error) {
	var clusterInstance *v1.Cluster

	var err error

	if value, ok := service.ObjectMeta.Annotations[v1.NativeLBAnnotationKey]; ok {
		clusterInstance, err = c.Cluster().Get(value)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, fmt.Errorf("provider wasn't found for service %s", service.Name)
			}
			return nil, err
		}
	} else {
		labelSelector := labels.Set{}
		labelSelector[v1.NativeLBDefaultLabel] = "true"
		clusterList, err := c.Cluster().List(&client.ListOptions{LabelSelector: labelSelector.AsSelector()})
		if err != nil {
			return nil, err
		}

		if len(clusterList.Items) == 0 {
			return nil, fmt.Errorf("default provider wans't found")
		} else if len(clusterList.Items) > 1 {
			return nil, fmt.Errorf("more then one default provider was found")
		}

		clusterInstance = &clusterList.Items[0]
	}

	return clusterInstance, nil
}

// TODO: need to check this function maybe split it to smaller pieces
func (c *ClusterController) UpdateServerData(farm *v1.Farm, serverData []v1.ServerData) error {
	farmName := farm.Name
	existServersMap := make(map[string]interface{})

	serverOwnerRef := []metav1.OwnerReference{{Name: farmName, APIVersion: v1.SchemeGroupVersion.Version, Kind: "Farm", UID: farm.UID}}

	labelSelector := labels.Set{}
	labelSelector[v1.NativeLBFarmRef] = farmName

	serversList, err := c.Server().List(&client.ListOptions{LabelSelector: labelSelector.AsSelector()})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get a list of servers related to %s farm error %v", farmName, err)
		return err
	}

	for _, server := range serversList.Items {
		existServersMap[server.Name] = struct{}{}
	}

	for _, serverDataObject := range serverData {
		serverName := serverDataObject.Server.Name
		serverObject, err := c.Server().Get(serverName)
		if err != nil && errors.IsNotFound(err) {
			serverObject = serverDataObject.Server.DeepCopy()
			serverObject.OwnerReferences = serverOwnerRef
			serverObject, err = c.Server().Create(serverObject)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			serverObject.OwnerReferences = serverOwnerRef
			serverObject.Spec = serverDataObject.Server.Spec
			serverObject.Status = serverDataObject.Server.Status

			serverObject, err = c.Server().Update(serverObject)
			if err != nil {
				return err
			}
		}

		farm.Spec.Servers[serverObject.Name] = &serverObject.Spec
		ownerRef := []metav1.OwnerReference{{Name: serverObject.Name, APIVersion: v1.SchemeGroupVersion.Version, Kind: "Server", UID: serverObject.UID}}
		delete(existServersMap, serverObject.Name)

		existBackendsMap := make(map[string]interface{})

		backendLabelSelector := labels.Set{}
		backendLabelSelector[v1.NativeLBServerRef] = serverObject.Name

		backendsList, err := c.Backend().List(&client.ListOptions{LabelSelector: backendLabelSelector.AsSelector()})
		if err != nil {
			log.Log.Reason(err).Errorf("failed to get a list of backends related to %s server error %v", serverObject.Name, err)
			return err
		}

		for _, backend := range backendsList.Items {
			existBackendsMap[backend.Name] = struct{}{}
		}

		for _, backend := range serverDataObject.Backends {
			backendObject, err := c.Backend().Get(backend.Name)
			if err != nil && errors.IsNotFound(err) {
				backendObject = backend.DeepCopy()
				backendObject.OwnerReferences = ownerRef
				backendObject, err = c.Backend().Create(backendObject)
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				backendObject.OwnerReferences = ownerRef
				backendObject.Spec = backend.Spec
				backendObject.Status = backend.Status
				backendObject, err = c.Backend().Update(backendObject)
				if err != nil {
					return err
				}
			}

			delete(existBackendsMap, backendObject.Name)
		}

		for deletedBackendName := range existBackendsMap {
			err = c.Backend().Delete(deletedBackendName)
			if err != nil {
				log.Log.Reason(err).Errorf("failed to delete backend %s error %v", deletedBackendName, err)
				return err
			}
		}
	}

	for deletedServerName := range existServersMap {
		err = c.Server().Delete(deletedServerName)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to delete server %s error %v", deletedServerName, err)
			return err
		}
		delete(farm.Spec.Servers, deletedServerName)
	}
	return nil
}

func (c *ClusterController) KeepAliveAgents() {
	c.GetManager().GetCache().WaitForCacheSync(c.clusterConnection.StopChan)
	for {
		select {
		case <-c.clusterConnection.StopChan:
			return
		case <-time.Tick(v1.KeepaliveTime * time.Second):
			c.keepalive()
		}
	}
}

func (c *ClusterController) keepalive() {
	agents, err := c.Agent().List(&client.ListOptions{})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get agent list for keepalive check")
	}

	numOfAgents := len(agents.Items)
	for idx, agent := range agents.Items {
		cluster, err := c.Cluster().Get(agent.Spec.Cluster)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to get cluster %s object for agent %s", agent.Spec.Cluster, agent.Name)
			continue
		}
		c.clusterConnection.GetAgentStatus(&agent, idx+1, numOfAgents)

		if cluster.Status.Agents == nil {
			cluster.Status.Agents = make(map[string]*v1.Agent)
		}

		cluster.Status.Agents[agent.Name] = &agent
		updatedCluster, err := c.updateLabels(cluster, v1.ClusterConnectionStatusSuccess)
		if err == nil {
			cluster = updatedCluster
		}
	}
}
