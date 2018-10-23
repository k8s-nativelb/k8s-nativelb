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

func (n *nativeLB) Farm() FarmInterface {
	return &farm{n.Client}
}

type farm struct {
	client.Client
}

func (c *farm) Create(farm *v1.Farm) (*v1.Farm, error) {
	err := c.Client.Create(context.Background(), farm)
	if err != nil {
		return nil, err
	}

	farm, err = c.Get(farm.Name)
	if err != nil {
		return nil, err
	}

	return farm, nil
}

func (c *farm) Get(name string) (*v1.Farm, error) {
	farm := &v1.Farm{}
	getRetry := 5
	var err error

	for i := 0; i < getRetry; i++ {
		err = c.Client.Get(context.Background(), client.ObjectKey{Name: name, Namespace: v1.ControllerNamespace}, farm)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		} else if err == nil {
			return farm, nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return nil, err
}

func (c *farm) Update(farm *v1.Farm) (*v1.Farm, error) {
	result := &v1.Farm{}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		getErr := c.Client.Get(context.Background(), client.ObjectKey{Name: farm.Name, Namespace: v1.ControllerNamespace}, result)
		if getErr != nil {
			return fmt.Errorf("Failed to get latest version of Farm: %v", getErr)
		}

		result.Spec = farm.Spec
		result.Status = farm.Status
		updateErr := c.Client.Update(context.Background(), result)
		return updateErr
	})

	if retryErr != nil {
		return nil, retryErr
	}

	return result, nil
}

func (c *farm) Delete(name string) error {
	farm, err := c.Get(name)
	if err != nil {
		return err
	}

	err = c.Client.Delete(context.Background(), farm)
	return err
}

func (c *farm) List(opts *client.ListOptions) (*v1.FarmList, error) {
	farmList := &v1.FarmList{}
	err := c.Client.List(context.Background(), opts, farmList)

	return farmList, err
}
