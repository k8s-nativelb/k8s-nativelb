package cluster_controller

import (
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *ClusterController) updateLabels(cluster *v1.Cluster, status string) {
	if cluster.Labels == nil {
		cluster.Labels = make(map[string]string)
	}
	cluster.Status.ConnectionStatus = status
	cluster.Status.LastUpdate = metav1.Now()
	cluster, err := c.Cluster().Update(cluster)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to update labels and status on cluster %s error %v", cluster.Name, err)
	}
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
	c.updateLabels(cluster, v1.ClusterConnectionStatusFail)
}

func (c *ClusterController) updateSuccessStatus(cluster *v1.Cluster, eventType, reason, message string) {
	c.Reconcile.Event.Event(cluster.DeepCopy(), eventType, reason, message)
	c.updateLabels(cluster, v1.ClusterConnectionStatusSuccess)
}
