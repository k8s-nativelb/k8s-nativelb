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

	cluster.Status.AllocatedIps[farm.Status.IpAdress] = farm.FarmName()
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
		clusterInstance, err = c.Cluster(v1.ControllerNamespace).Get(value)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, fmt.Errorf("provider wasn't found for service %s", service.Name)
			}
			return nil, err
		}
	} else {
		labelSelector := labels.Set{}
		labelSelector[v1.NativeLBDefaultLabel] = "true"
		clusterList, err := c.Cluster(v1.ControllerNamespace).List(&client.ListOptions{LabelSelector: labelSelector.AsSelector()})
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
func (c *ClusterController) UpdateServerData(farm *v1.Farm, serverData []*v1.Server) error {
	farmName := farm.Name
	existServersMap := make(map[string]interface{})

	serverOwnerRef := []metav1.OwnerReference{{Name: farmName, APIVersion: v1.SchemeGroupVersion.Version, Kind: "Farm", UID: farm.UID}}

	labelSelector := labels.Set{}
	labelSelector[v1.NativeLBFarmRef] = farmName

	serversList, err := c.Server(farm.Namespace).List(&client.ListOptions{LabelSelector: labelSelector.AsSelector()})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get a list of servers related to %s farm error %v", farmName, err)
		return err
	}

	for _, server := range serversList.Items {
		existServersMap[server.Name] = struct{}{}
	}

	for _, serverDataObject := range serverData {
		serverName := serverDataObject.Name
		serverObject, err := c.Server(farm.Namespace).Get(serverName)
		if err != nil && errors.IsNotFound(err) {
			serverObject = serverDataObject.DeepCopy()
			serverObject.OwnerReferences = serverOwnerRef
			serverObject, err = c.Server(farm.Namespace).Create(serverObject)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			serverObject.OwnerReferences = serverOwnerRef
			serverObject.Spec = serverDataObject.Spec
			serverObject.Status = serverDataObject.Status

			serverObject, err = c.Server(farm.Namespace).Update(serverObject)
			if err != nil {
				return err
			}
		}

		farm.Spec.Servers[serverObject.Name] = &serverObject.Spec
		delete(existServersMap, serverObject.Name)
	}

	for deletedServerName := range existServersMap {
		err = c.Server(farm.Namespace).Delete(deletedServerName)
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
	clusters, err := c.Cluster(v1.ControllerNamespace).List(&client.ListOptions{})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get cluster list for keepalive check")
	}

	for _, cluster := range clusters.Items {
		labelSelector := labels.Set{}
		labelSelector[v1.ClusterLabel] = cluster.Name
		agents, err := c.Agent(v1.ControllerNamespace).List(&client.ListOptions{LabelSelector: labelSelector.AsSelector()})
		if err != nil {
			log.Log.Reason(err).Errorf("failed to get agent list for keepalive check on cluster %s", cluster.Name)
		}

		numOfAgents := len(agents.Items)
		for idx, agent := range agents.Items {
			c.clusterConnection.GetAgentStatus(&agent, idx+1, numOfAgents)

			if cluster.Status.Agents == nil {
				cluster.Status.Agents = make(map[string]*v1.Agent)
			}

			cluster.Status.Agents[agent.Name] = &agent
			_, err := c.updateClusterLabelsAnStatus(&cluster, v1.ClusterConnectionStatusSuccess)
			if err != nil {
				log.Log.Reason(err).Errorf("failed to update cluster %s with agent %s status", cluster.Name, agent.Name)
			}
		}

	}
}
