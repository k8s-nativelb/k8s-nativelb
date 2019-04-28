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

package loadbalancer

//go:generate mockgen -source $GOFILE -package=$GOPACKAGE -destination=generated_mock_$GOFILE

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/nativelb-agent/handler"
	"github.com/k8s-nativelb/pkg/proto"
)

const (
	haproxyStatusSocketFile = "/run/haproxy.sock"
)

type LoadBalancerInterface interface {
	GetPid() int32
	GetStatus() (*proto.HaproxyStatus, error)
	GetStats() (*proto.ServersStats, error)
	UpdateFarm(*proto.FarmSpec) error
	RemoveFarm(*proto.FarmSpec) error
	LoadInitData(*proto.InitAgentData) error
	StartEngine() error
	ReloadEngine() error
	StopEngine()
	WriteConfig() error
}

type LoadBalancer struct {
	tmpl    *template.Template
	handler handler.HandlerInterface
	pid     string
	servers map[string]*proto.Server
}

func NewLoadBalancer() (*LoadBalancer, error) {
	handlerInstance := &handler.Handler{}
	tmpl, err := template.ParseFiles(handler.HaproxyTemplateFile)
	if err != nil {
		return nil, err
	}

	return &LoadBalancer{tmpl: tmpl, handler: handlerInstance, servers: map[string]*proto.Server{}}, nil
}

func (l *LoadBalancer) GetPid() int32 {
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

func (l *LoadBalancer) UpdateFarm(farm *proto.FarmSpec) error {
	for serverName, server := range farm.Servers {
		if !proto.IsTCPServer(server) {
			break
		}

		l.servers[fmt.Sprintf("%s-%s", farm.Namespace, serverName)] = server
	}

	return nil
}

func (l *LoadBalancer) RemoveFarm(farm *proto.FarmSpec) error {
	for serverName, server := range farm.Servers {
		if !proto.IsTCPServer(server) {
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

func (l *LoadBalancer) LoadInitData(data *proto.InitAgentData) error {
	l.servers = map[string]*proto.Server{}

	for _, farm := range data.Farms {
		for serverName, server := range farm.Servers {
			serverName = fmt.Sprintf("%s-%s", farm.Namespace, serverName)
			if proto.IsTCPServer(server) {
				l.servers[serverName] = server
			}
		}
	}

	return nil
}

func (l *LoadBalancer) WriteConfig() error {
	w, err := os.Create(handler.HaproxyConfigFile)
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

func (l *LoadBalancer) StartEngine() error {
	if l.GetPid() != 0 {
		return l.ReloadEngine()
	}

	pid, err := l.handler.StartHaproxy()
	if err != nil {
		return err
	}
	l.pid = pid

	return nil
}

func (l *LoadBalancer) ReloadEngine() error {
	pid, err := l.handler.ReloadHaproxy(l.pid)
	if err != nil {
		return err
	}
	l.pid = pid

	return nil
}

func (l *LoadBalancer) StopEngine() {
	err := l.handler.StopHaproxy(l.pid)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to stop Haproxy")
		return
	}
	l.pid = ""
}

// RunCommand is the entrypoint to the client. Sends an arbitray command string to HAProxy.
func (l *LoadBalancer) RunCommand(cmd string) (*bytes.Buffer, error) {
	var err error
	for retry := 0; retry <= 5; retry++ {
		_, err = os.Stat(haproxyStatusSocketFile)
		if err == nil {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}

	if err != nil {
		return nil, err
	}

	timeout := time.Duration(30) * time.Second
	conn, err := net.DialTimeout("unix", haproxyStatusSocketFile, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	result := bytes.NewBuffer(nil)

	_, err = conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(result, conn)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(result.String(), "Unknown command") {
		return nil, fmt.Errorf("Unknown command: %s", cmd)
	}

	return result, nil
}

func (l *LoadBalancer) GetStatus() (*proto.HaproxyStatus, error) {
	infoData, err := l.info()
	if err != nil {
		return nil, err
	}
	log.Log.Infof("%v", *infoData)

	haproxyStatus := &proto.HaproxyStatus{
		CompressBpsIn:              infoData.CompressBpsIn,
		CompressBpsOut:             infoData.CompressBpsOut,
		CompressBpsRateLim:         infoData.CompressBpsRateLim,
		ConnRate:                   infoData.ConnRate,
		ConnRateLimit:              infoData.ConnRateLimit,
		CumConns:                   infoData.CumConns,
		CumReq:                     infoData.CumReq,
		CumSslConns:                infoData.CumSslConns,
		CurrConns:                  infoData.CurrConns,
		CurrSslConns:               infoData.CurrSslConns,
		HardMaxconn:                infoData.HardMaxconn,
		IdlePct:                    infoData.IdlePct,
		Maxconn:                    infoData.Maxconn,
		MaxConnRate:                infoData.MaxConnRate,
		Maxpipes:                   infoData.Maxpipes,
		MaxSessRate:                infoData.MaxSessRate,
		Maxsock:                    infoData.Maxsock,
		MaxSslConns:                infoData.MaxSslConns,
		MaxSslRate:                 infoData.MaxSslRate,
		MemMaxMB:                   infoData.MemMaxMB,
		Nbproc:                     infoData.Nbproc,
		Pid:                        infoData.Pid,
		PipesFree:                  infoData.PipesFree,
		PipesUsed:                  infoData.PipesUsed,
		ProcessNum:                 infoData.ProcessNum,
		ReleaseDate:                infoData.ReleaseDate,
		RunQueue:                   infoData.RunQueue,
		SessRate:                   infoData.SessRate,
		SessRateLimit:              infoData.SessRateLimit,
		SslBackendKeyRate:          infoData.SslBackendKeyRate,
		SslBackendMaxKeyRate:       infoData.SslBackendMaxKeyRate,
		SslCacheLookups:            infoData.SslCacheLookups,
		SslCacheMisses:             infoData.SslCacheMisses,
		SslFrontendKeyRate:         infoData.SslFrontendKeyRate,
		SslFrontendMaxKeyRate:      infoData.SslFrontendMaxKeyRate,
		SslFrontendSessionReusePct: infoData.SslFrontendSessionReusePct,
		SslRate:                    infoData.SslRate,
		SslRateLimit:               infoData.SslRateLimit,
		Tasks:                      infoData.Tasks,
		UlimitN:                    infoData.UlimitN,
		Uptime:                     infoData.Uptime,
		UptimeSec:                  infoData.UptimeSec,
		Version:                    infoData.Version,
	}

	return haproxyStatus, nil
}
