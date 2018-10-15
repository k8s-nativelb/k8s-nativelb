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
package kubecli

import (
	"context"
	"fmt"
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func (n *nativeLB) Agent() AgentInterface {
	return &agent{n.Client}
}

type agent struct {
	client.Client
}

func (c *agent) Create(agent *v1.Agent) (*v1.Agent, error) {
	err := c.Client.Create(context.Background(), agent)
	if err != nil {
		return nil, err
	}

	err = c.Client.Get(context.Background(), client.ObjectKey{Name: agent.Name, Namespace: v1.ControllerNamespace}, agent)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

func (c *agent) Get(name string) (*v1.Agent, error) {
	agent := &v1.Agent{}
	getRetry := 5
	var err error

	for i := 0; i < getRetry; i++ {
		err = c.Client.Get(context.Background(), client.ObjectKey{Name: name, Namespace: v1.ControllerNamespace}, agent)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		} else if err == nil {
			return agent, nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return nil, err
}

func (c *agent) Update(agent *v1.Agent) (*v1.Agent, error) {
	result := &v1.Agent{}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		getErr := c.Client.Get(context.Background(), client.ObjectKey{Name: agent.Name, Namespace: v1.ControllerNamespace}, result)
		if getErr != nil {
			return fmt.Errorf("Failed to get latest version of Agent: %v", getErr)
		}

		result.Spec = agent.Spec
		result.Status = agent.Status
		updateErr := c.Client.Update(context.Background(), result)
		return updateErr
	})

	if retryErr != nil {
		return nil, retryErr
	}

	return result, nil
}

func (c *agent) Delete(name string) error {
	agent, err := c.Get(name)
	if err != nil {
		return err
	}

	err = c.Client.Delete(context.Background(), agent)
	return err
}

func (c *agent) List(opts *client.ListOptions) (*v1.AgentList, error) {
	agentList := &v1.AgentList{}
	err := c.Client.List(context.Background(), opts, agentList)

	return agentList, err
}
