package server_controller

import "github.com/k8s-nativelb/pkg/log"

func (s *ServerController)WaitForStatusUpdate() {
	for statusUpdate := range s.serverStatsChannel{
		log.Log.V(4).Infof("New message on server status channel %+v",statusUpdate)
	}
}
