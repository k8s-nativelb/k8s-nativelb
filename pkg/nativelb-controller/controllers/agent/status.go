package agent_controller

import "github.com/k8s-nativelb/pkg/log"

func (a *AgentController) WaitForStatusUpdate() {
	for statusUpdate := range a.AgentStatusChannel {
		log.Log.V(4).Infof("New message on agent status channel %+v", statusUpdate)
	}
}
