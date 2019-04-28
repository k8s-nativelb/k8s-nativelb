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

func (n *nativeLB) Cluster(namespace string) ClusterInterface {
	return &cluster{n.Client, namespace}
}

type cluster struct {
	client.Client
	Namespace string
}

func (c *cluster) Create(cluster *v1.Cluster) (*v1.Cluster, error) {
	cluster.Namespace = c.Namespace
	err := c.Client.Create(context.Background(), cluster)
	if err != nil {
		return nil, err
	}

	cluster, err = c.Get(cluster.Name)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

func (c *cluster) Get(name string) (*v1.Cluster, error) {
	cluster := &v1.Cluster{}
	getRetry := 5
	var err error

	for i := 0; i < getRetry; i++ {
		err = c.Client.Get(context.Background(), client.ObjectKey{Name: name, Namespace: c.Namespace}, cluster)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		} else if err == nil {
			return cluster, nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return nil, err
}

func (c *cluster) Update(cluster *v1.Cluster) (*v1.Cluster, error) {
	result := &v1.Cluster{}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		getErr := c.Client.Get(context.Background(), client.ObjectKey{Name: cluster.Name, Namespace: c.Namespace}, result)
		if getErr != nil {
			return fmt.Errorf("Failed to get latest version of Cluster: %v", getErr)
		}

		result.Spec = cluster.Spec
		result.Status = cluster.Status
		updateErr := c.Client.Update(context.Background(), result)
		return updateErr
	})

	if retryErr != nil {
		return nil, retryErr
	}

	return result, nil
}

func (c *cluster) Delete(name string) error {
	cluster, err := c.Get(name)
	if err != nil {
		return err
	}
	cluster.Namespace = c.Namespace
	err = c.Client.Delete(context.Background(), cluster)
	return err
}

func (c *cluster) List(opts *client.ListOptions) (*v1.ClusterList, error) {
	opts.Namespace = c.Namespace
	clusterList := &v1.ClusterList{}
	err := c.Client.List(context.Background(), opts, clusterList)

	return clusterList, err
}
