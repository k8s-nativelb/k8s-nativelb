package cluster_controller

import (
	"fmt"
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"context"
)

func (c *ClusterController) CreateFarm(farm *v1.Farm,cluster *v1.Cluster) (string, error) {
	farmIpAddress, err := c.allocator[cluster.Name].Allocate(farm,cluster)
	if err != nil {
		log.Log.V(2).Errorf("Fail to allocate Ip address for farm: %s on cluster %s error message: %v", farm.Name, cluster.Name, err)
		c.clusterUpdateFailStatus(cluster, "Warning", "FarmCreateFail", err.Error())
		return "", err
	}

	farm.Status.IpAdress= farmIpAddress
	farm.UpdateServers(cluster.Spec.Internal)

	err = c.clusterConnection.CreateFarmOnCluster(farm,cluster)
	if err != nil {
		log.Log.Errorf("Fail to create farm %s on cluster %s error %v",farm.Name,cluster.Name,err)
		c.clusterUpdateFailStatus(cluster, "Warning", "FarmCreateFail", err.Error())
		return "",err
	}

	c.updateClusterObject(cluster)
	c.clusterUpdateSuccessStatus(cluster, "Normal", "FarmCreateSuccess", fmt.Sprintf("Farm %s-%s was created on provider", farm.Namespace, farm.Name))
	return farmIpAddress, nil
}

func (c *ClusterController) UpdateFarm(farm *v1.Farm,cluster *v1.Cluster) (string, error) {
	farm.UpdateServers(cluster.Spec.Internal)

	cluster.Status.AllocatedIps[farm.Status.IpAdress] = farm
	c.updateClusterObject(cluster)

	err := c.clusterConnection.UpdateFarmOnCluster(farm,cluster)
	if err != nil {
		log.Log.V(2).Errorf("Fail to update farm: %s on cluster %s error message: %v", farm.FarmName(), cluster.Name, err)
		c.clusterUpdateFailStatus(cluster, "Warning", "FarmUpdateFail", err.Error())
		return "", err
	}

	log.Log.V(2).Infof("successfully updated farm: %s on cluster %s", farm.FarmName(), cluster.Name)
	c.clusterUpdateSuccessStatus(cluster, "Normal", "FarmUpdateSuccess", fmt.Sprintf("Farm %s-%s was updated on cluster", farm.Namespace, farm.Name))
	return farm.Status.IpAdress, nil
}

func (c *ClusterController) DeleteFarm(farm *v1.Farm,cluster *v1.Cluster) error {
	err:=c.clusterConnection.DeleteFarmOnCluster(farm,cluster)
	if err != nil {
		log.Log.V(2).Errorf("Fail to remove farm: %s on cluster %s error message: %s", farm.FarmName(), cluster.Name, err.Error())
		c.clusterUpdateFailStatus(cluster, "Warning", "FarmDeleteFail", err.Error())
		return err
	}

	c.allocator[cluster.Name].Release(farm.Status.IpAdress,cluster)
	c.updateClusterObject(cluster)

	log.Log.V(2).Infof("successfully removed farm: %s on cluster %s", farm.FarmName(), cluster.Name)
	c.clusterUpdateSuccessStatus(cluster, "Normal", "FarmDeleteSuccess", fmt.Sprintf("Farm %s-%s was deleted on cluster", farm.Namespace, farm.Name))
	return nil
}

func (c *ClusterController) updateLabels(cluster *v1.Cluster, status string) {
	if cluster.Labels == nil {
		cluster.Labels = make(map[string]string)
	}
	cluster.Status.ConnectionStatus = status
	cluster.Status.LastUpdate = metav1.Now()
	c.Reconcile.Client.Update(context.TODO(), cluster)
}

func (c *ClusterController) clusterUpdateFailStatus(cluster *v1.Cluster, eventType, reason, message string) {
	c.Reconcile.Event.Event(cluster.DeepCopyObject(), eventType, reason, message)
	c.updateLabels(cluster, v1.ClusterConnectionStatusFail)
}

func (c *ClusterController) clusterUpdateSuccessStatus(cluster *v1.Cluster, eventType, reason, message string) {
	c.Reconcile.Event.Event(cluster.DeepCopy(), eventType, reason, message)
	c.updateLabels(cluster,v1.ClusterConnectionStatusSuccess)
}

func (c *ClusterController) updateClusterObject(cluster *v1.Cluster) (error) {
	err := c.Reconcile.Update(context.Background(),cluster)
	if err != nil {
		log.Log.Errorf("fail to update cluster %s error %v",cluster.Name,err)
		return err
	}

	return nil
}