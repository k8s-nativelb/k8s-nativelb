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

package udp_loadbalancer

//go:generate mockgen -source $GOFILE -package=$GOPACKAGE -destination=generated_mock_$GOFILE

import (
	"fmt"
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/nativelb-agent/handler"
	"github.com/k8s-nativelb/pkg/proto"
	"os"
	"strconv"
	"text/template"
)

type UdpLoadBalancerInterface interface {
	GetPid() int32
	GetStatus() (*proto.NginxStatus, error)
	UpdateFarm(*proto.FarmSpec) error
	RemoveFarm(*proto.FarmSpec) error
	LoadInitData(*proto.InitAgentData) error
	StartEngine() error
	ReloadEngine() error
	StopEngine()
	WriteConfig() error
}

type UdpLoadBalancer struct {
	tmpl    *template.Template
	handler handler.HandlerInterface
	pid     string
	servers map[string]*proto.Server
}

func NewUdpLoadBalancer() (*UdpLoadBalancer, error) {
	handlerInstance := &handler.Handler{}
	tmpl, err := template.ParseFiles(handler.NginxTemplateFile)
	if err != nil {
		return nil, err
	}

	return &UdpLoadBalancer{tmpl: tmpl, handler: handlerInstance, servers: map[string]*proto.Server{}}, nil
}

func (l *UdpLoadBalancer) GetPid() int32 {
	if l.pid == "" {
		return 0
	}

	pid, err := strconv.Atoi(l.pid)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to convert pid %s to int", l.pid)
		return 0
	}

	return int32(pid)
}

func (l *UdpLoadBalancer) UpdateFarm(farm *proto.FarmSpec) error {
	for serverName, server := range farm.Servers {
		if proto.IsTCPServer(server) {
			break
		}

		l.servers[fmt.Sprintf("%s-%s", farm.Namespace, serverName)] = server
	}

	return nil
}

func (l *UdpLoadBalancer) RemoveFarm(farm *proto.FarmSpec) error {
	for serverName, server := range farm.Servers {
		if proto.IsTCPServer(server) {
			break
		}
		serverName = fmt.Sprintf("%s-%s", farm.Namespace, serverName)
		if _, ok := l.servers[serverName]; !ok {
			return fmt.Errorf("failed to find server %s farm %s in the configuration", serverName, farm.FarmName)
		}

		delete(l.servers, serverName)
	}

	return nil
}

func (l *UdpLoadBalancer) LoadInitData(data *proto.InitAgentData) error {
	l.servers = map[string]*proto.Server{}

	for _, farm := range data.Farms {
		for serverName, server := range farm.Servers {
			serverName = fmt.Sprintf("%s-%s", farm.Namespace, serverName)
			if !proto.IsTCPServer(server) {
				l.servers[serverName] = server
			}
		}
	}

	return nil
}

func (l *UdpLoadBalancer) WriteConfig() error {
	w, err := os.Create(handler.NginxConfigFile)
	if err != nil {
		return err
	}
	defer w.Close()

	if l.tmpl.Execute(w, l.servers) != nil {
		log.Log.Reason(err).Errorf("failed to execute template error %v", err)
		return err
	}
	return nil
}

func (l *UdpLoadBalancer) StartEngine() error {
	if l.GetPid() != 0 {
		return l.ReloadEngine()
	}

	pid, err := l.handler.StartNginx()
	if err != nil {
		return err
	}
	l.pid = pid

	return nil
}

func (l *UdpLoadBalancer) ReloadEngine() error {
	pid, err := l.handler.ReloadNginx(l.pid)
	if err != nil {
		return err
	}
	l.pid = pid

	return nil
}

func (l *UdpLoadBalancer) StopEngine() {
	err := l.handler.StopNginx(l.pid)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to stop Haproxy")
		return
	}
	l.pid = ""
}

func (l *UdpLoadBalancer) GetStatus() (*proto.NginxStatus, error) {
	return l.info()
}
