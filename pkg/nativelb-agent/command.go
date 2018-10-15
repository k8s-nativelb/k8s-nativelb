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
	"github.com/k8s-nativelb/pkg/log"
	. "github.com/k8s-nativelb/pkg/proto"
)

func (n *NativelbAgent) CreateServers(servers []*Server) error {
	log.Log.Infof("Get Create command with servers %+v", servers)
	return nil
}

func (n *NativelbAgent) UpdateServers(servers []*Server) error {
	log.Log.Infof("Get Update command with servers %+v", servers)
	return nil
}

func (n *NativelbAgent) DeleteServers(servers []*Server) error {
	log.Log.Infof("Get Delete command with servers %+v", servers)
	return nil
}
