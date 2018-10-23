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
	"context"
	"github.com/k8s-nativelb/pkg/log"
	. "github.com/k8s-nativelb/pkg/proto"
)

func (n *NativelbAgent) CreateServers(ctx context.Context, data *Data) (*Result, error) {
	log.Log.Infof("%v", data)
	return &Result{}, nil
}
func (n *NativelbAgent) UpdateServers(ctx context.Context, data *Data) (*Result, error) {
	log.Log.Infof("%v", data)
	return &Result{}, nil
}
func (n *NativelbAgent) DeleteServers(ctx context.Context, data *Data) (*Result, error) {
	log.Log.Infof("%v", data)
	return &Result{}, nil
}
func (n *NativelbAgent) GetAgentStatus(ctx context.Context, cmd *Command) (*AgentStatus, error) {
	log.Log.Infof("%v", cmd)
	return &AgentStatus{}, nil
}
func (n *NativelbAgent) GetServerStats(ctx context.Context, cmd *Command) (*ServerStats, error) {
	log.Log.Infof("%v", cmd)
	return &ServerStats{}, nil
}
