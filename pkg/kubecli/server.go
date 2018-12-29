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

func (n *nativeLB) Server() ServerInterface {
	return &server{n.Client}
}

func (c *server) Create(server *v1.Server) (*v1.Server, error) {
	err := c.Client.Create(context.Background(), server)
	if err != nil {
		return nil, err
	}

	server, err = c.Get(server.Name)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (c *server) Get(name string) (*v1.Server, error) {
	server := &v1.Server{}
	getRetry := 5
	var err error

	for i := 0; i < getRetry; i++ {
		err = c.Client.Get(context.Background(), client.ObjectKey{Name: name, Namespace: v1.ControllerNamespace}, server)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		} else if err == nil {
			return server, nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return nil, err
}

type server struct {
	client.Client
}

func (c *server) Update(server *v1.Server) (*v1.Server, error) {
	result := &v1.Server{}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		getErr := c.Client.Get(context.Background(), client.ObjectKey{Name: server.Name, Namespace: v1.ControllerNamespace}, result)
		if getErr != nil {
			return fmt.Errorf("Failed to get latest version of Server: %v", getErr)
		}

		result.Spec = server.Spec
		result.Status = server.Status
		updateErr := c.Client.Update(context.Background(), result)
		return updateErr
	})

	if retryErr != nil {
		return nil, retryErr
	}

	return result, nil
}

func (c *server) Delete(name string) error {
	server, err := c.Get(name)
	if err != nil {
		return err
	}

	err = c.Client.Delete(context.Background(), server)
	return err
}

func (c *server) List(opts *client.ListOptions) (*v1.ServerList, error) {
	serverList := &v1.ServerList{}
	err := c.Client.List(context.Background(), opts, serverList)

	return serverList, err
}
