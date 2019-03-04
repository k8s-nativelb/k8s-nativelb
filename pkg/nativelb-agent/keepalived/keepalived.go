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
package keepalived

//go:generate mockgen -source $GOFILE -package=$GOPACKAGE -destination=generated_mock_$GOFILE

import (
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/nativelb-agent/handler"
	"github.com/k8s-nativelb/pkg/proto"
)

type KeepalivedInterface interface {
	GetPid() int32
	WriteConfig() error
	NewFarmForInstance(*proto.Data) error
	DeleteFarmInInstance(*proto.Data) error
	BuildIpsFromFarmsForNamespace(string)
	LoadInitData(*proto.InitAgentData) error
	StartEngine() error
	ReloadEngine() error
	StopEngine()
}

type VrrpInstance struct {
	Namespace string
	Iface     string
	MainVip   string
	SecVips   []string
	RouterID  int32
	Priority  int32
	State     string
	farms     map[string]*proto.Data
}

type Keepalived struct {
	iface         string
	pid           string
	handler       handler.HandlerInterface
	tmpl          *template.Template
	vrrpInstances map[string]*VrrpInstance
}

func NewKeepalived(iface string) (*Keepalived, error) {
	handlerInstance := &handler.Handler{}
	tmpl, err := template.ParseFiles(handler.KeepalivedTmpl)
	if err != nil {
		return nil, err
	}

	return &Keepalived{tmpl: tmpl, iface: iface, handler: handlerInstance, vrrpInstances: make(map[string]*VrrpInstance)}, nil
}

func (k *Keepalived) NewFarmForInstance(data *proto.Data) error {
	if _, ok := k.vrrpInstances[data.Namespace]; !ok {
		instance := &VrrpInstance{Namespace: data.Namespace,
			Iface:    k.iface,
			MainVip:  "",
			RouterID: data.RouterID,
			SecVips:  make([]string, 0),
			State:    data.KeepalivedState,
			Priority: data.Priority,
			farms:    map[string]*proto.Data{data.FarmName: data}}

		k.vrrpInstances[data.Namespace] = instance
	} else {
		k.vrrpInstances[data.Namespace].farms[data.FarmName] = data
	}

	k.BuildIpsFromFarmsForNamespace(data.Namespace)
	return nil
}

func (k *Keepalived) GetPid() int32 {
	if k.pid == "" {
		return 0
	}

	pid, err := strconv.Atoi(k.pid)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to convert pid %s to int", k.pid)
		return 0
	}

	return int32(pid)
}

func (k *Keepalived) DeleteFarmInInstance(data *proto.Data) error {
	if _, ok := k.vrrpInstances[data.Namespace]; !ok {
		return fmt.Errorf("failed to find namespace %s in the configuration", data.Namespace)
	}

	if _, ok := k.vrrpInstances[data.Namespace].farms[data.FarmName]; !ok {
		return fmt.Errorf("failed to find farm %s in namespace %s", data.FarmName, data.Namespace)
	}

	delete(k.vrrpInstances[data.Namespace].farms, data.FarmName)
	if len(k.vrrpInstances[data.Namespace].farms) == 0 {
		delete(k.vrrpInstances, data.Namespace)
	} else {
		k.BuildIpsFromFarmsForNamespace(data.Namespace)
	}

	return nil
}

func (k *Keepalived) BuildIpsFromFarmsForNamespace(namespace string) {
	instance := k.vrrpInstances[namespace]
	instance.MainVip = ""
	instance.SecVips = make([]string, 0)

	for _, farm := range instance.farms {
		for i := 0; i < len(farm.Servers); i++ {
			if instance.MainVip == "" {
				instance.MainVip = farm.Servers[i].Bind
			} else {
				instance.SecVips = append(instance.SecVips, farm.Servers[i].Bind)
			}
		}
	}
}

func (k *Keepalived) LoadInitData(data *proto.InitAgentData) error {
	k.vrrpInstances = make(map[string]*VrrpInstance)

	for _, farm := range data.Data {
		err := k.NewFarmForInstance(farm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *Keepalived) WriteConfig() error {
	w, err := os.Create(handler.KeepalivedCfg)
	if err != nil {
		return err
	}
	defer w.Close()

	return k.tmpl.Execute(w, k.vrrpInstances)
}

func (k *Keepalived) ReloadEngine() error {
	pid, err := k.handler.ReloadKeepalived(k.pid)
	if err != nil {
		return err
	}
	k.pid = pid

	return nil
}

func (k *Keepalived) StartEngine() error {
	if k.pid != "" {
		return k.ReloadEngine()
	}

	pid, err := k.handler.StartKeepalived()
	if err != nil {
		return err
	}
	k.pid = pid

	return nil
}

func (k *Keepalived) StopEngine() {
	err := k.handler.StopKeepalived(k.pid)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to stop keepalived process")
	}
	k.pid = ""
}
