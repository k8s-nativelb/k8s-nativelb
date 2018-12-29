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

func (n *nativeLB) Backend() BackendInterface {
	return &backend{n.Client}
}

type backend struct {
	client.Client
}

func (c *backend) Create(backend *v1.Backend) (*v1.Backend, error) {
	err := c.Client.Create(context.Background(), backend)
	if err != nil {
		return nil, err
	}

	backend, err = c.Get(backend.Name)
	if err != nil {
		return nil, err
	}

	return backend, nil
}

func (c *backend) Get(name string) (*v1.Backend, error) {
	backend := &v1.Backend{}
	getRetry := 5
	var err error

	for i := 0; i < getRetry; i++ {
		err = c.Client.Get(context.Background(), client.ObjectKey{Name: name, Namespace: v1.ControllerNamespace}, backend)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		} else if err == nil {
			return backend, nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return nil, err
}

func (c *backend) Update(backend *v1.Backend) (*v1.Backend, error) {
	result := &v1.Backend{}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		getErr := c.Client.Get(context.Background(), client.ObjectKey{Name: backend.Name, Namespace: v1.ControllerNamespace}, result)
		if getErr != nil {
			return fmt.Errorf("Failed to get latest version of Backend: %v", getErr)
		}

		result.Spec = backend.Spec
		result.Status = backend.Status
		updateErr := c.Client.Update(context.Background(), result)
		return updateErr
	})

	if retryErr != nil {
		return nil, retryErr
	}

	return result, nil
}

func (c *backend) Delete(name string) error {
	backend, err := c.Get(name)
	if err != nil {
		return err
	}

	err = c.Client.Delete(context.Background(), backend)
	return err
}

func (c *backend) List(opts *client.ListOptions) (*v1.BackendList, error) {
	backendList := &v1.BackendList{}
	err := c.Client.List(context.Background(), opts, backendList)

	return backendList, err
}
