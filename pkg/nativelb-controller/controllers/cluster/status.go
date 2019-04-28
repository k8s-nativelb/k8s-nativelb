package cluster_controller

import (
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *ClusterController) updateClusterLabelsAnStatus(cluster *v1.Cluster, status string) (*v1.Cluster, error) {
	if cluster.Labels == nil {
		cluster.Labels = make(map[string]string)
	}
	updatedCluster := cluster.DeepCopy()
	updatedCluster.Status = cluster.Status
	updatedCluster.Status.ConnectionStatus = status
	updatedCluster.Status.LastUpdate = metav1.Now()
	updatedCluster, err := c.Cluster(v1.ControllerNamespace).Update(updatedCluster)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to update labels and status on cluster %s error %v", cluster.Name, err)
		return nil, err
	}
	return updatedCluster, nil
}

func (c *ClusterController) ifUpdateFailedUpdateFailStatus(err error, cluster *v1.Cluster, reason, message string) bool {
	if err != nil {
		log.Log.Reason(err).Errorf(message)
		c.updateFailStatus(cluster, "Warning", reason, message)
		return true
	}
	return false
}

func (c *ClusterController) updateFailStatus(cluster *v1.Cluster, eventType, reason, message string) {
	c.Reconcile.Event.Event(cluster.DeepCopyObject(), eventType, reason, message)
	_, err := c.updateClusterLabelsAnStatus(cluster, v1.ClusterConnectionStatusFail)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to update cluster %s with fail status", cluster.Name)
	}
}

func (c *ClusterController) updateSuccessStatus(cluster *v1.Cluster, eventType, reason, message string) {
	c.Reconcile.Event.Event(cluster.DeepCopy(), eventType, reason, message)
	_, err := c.updateClusterLabelsAnStatus(cluster, v1.ClusterConnectionStatusSuccess)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to update cluster %s with success status", cluster.Name)
	}
}
