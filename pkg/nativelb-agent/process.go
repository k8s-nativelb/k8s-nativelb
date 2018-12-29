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

package nativelb_agent

import (
	"fmt"
	"github.com/k8s-nativelb/pkg/log"
	. "github.com/k8s-nativelb/pkg/proto"
)

func (n *NativelbAgent) UpdateAndReload(data *Data) error {
	err := n.loadBalancerController.UpdateFarm(data)
	if err != nil {
		return fmt.Errorf("failed to update loadbalancer configuration error %v", err)
	}

	err = n.keepalivedController.NewFarmForInstance(data)
	if err != nil {
		return fmt.Errorf("failed to update keepalived configuration error %v", err)
	}

	return n.reload()
}

func (n *NativelbAgent) DeleteAndReload(data *Data) error {
	err := n.loadBalancerController.RemoveFarm(data)
	if err != nil {
		return fmt.Errorf("failed to remove loadbalancer configuration error %v", err)
	}

	err = n.keepalivedController.DeleteFarmInInstance(data)
	if err != nil {
		return fmt.Errorf("failed to remove keepalived configuration error %v", err)
	}

	return n.reload()
}

func (n *NativelbAgent) LoadInitToEngines(data *InitAgentData) error {
	err := n.loadBalancerController.LoadInitData(data)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to init loadbalancer with data %v error %v", *data, err)
		return err
	}

	err = n.keepalivedController.LoadInitData(data)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to init keepalived with data %v error %v", *data, err)
		return err
	}

	err = n.loadBalancerController.WriteConfig()
	if err != nil {
		return fmt.Errorf("failed to write loadbalancer configuration error %v", err)
	}

	err = n.keepalivedController.WriteConfig()
	if err != nil {
		return fmt.Errorf("failed to write keepalived configuration error %v", err)
	}

	err = n.loadBalancerController.StartEngine()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to start loadbalancer engine error %v", err)
		return err
	}

	err = n.keepalivedController.StartEngine()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to start keepalived engine error %v", err)
		return err
	}

	return nil
}

func (n *NativelbAgent) reload() error {
	err := n.loadBalancerController.WriteConfig()
	if err != nil {
		return fmt.Errorf("failed to write loadbalancer configuration error %v", err)
	}

	err = n.keepalivedController.WriteConfig()
	if err != nil {
		return fmt.Errorf("failed to write keepalived configuration error %v", err)
	}

	err = n.loadBalancerController.ReloadEngine()
	if err != nil {
		return fmt.Errorf("failed to reload loadbalancer engine error %v", err)
	}

	err = n.keepalivedController.ReloadEngine()
	if err != nil {
		return fmt.Errorf("failed to reload keepalived engine error %v", err)
	}

	return nil
}

func (n *NativelbAgent) StopEngines() {
	n.loadBalancerController.StopEngine()
	n.keepalivedController.StopEngine()
}
