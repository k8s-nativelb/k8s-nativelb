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

package handler

import (
	"github.com/k8s-nativelb/pkg/log"
	"os"
	"os/exec"
	"strconv"
)

//go:generate mockgen -source $GOFILE -package=$GOPACKAGE -destination=generated_mock_$GOFILE

const (
	HaproxyConfigFile   = "/etc/haproxy/haproxy.cfg"
	HaproxyTemplateFile = "/templates/haproxy-template.tmpl"
	HaproxyPID          = "/run/haproxy.pid"

	KeepalivedCfg  = "/etc/keepalived/keepalived.conf"
	KeepalivedTmpl = "/templates/keepalived-template.tmpl"
	KeepalivedPID  = "/run/keepalived.pid"

	NginxConfigFile   = "/etc/nginx/nginx.conf"
	NginxTemplateFile = "/templates/nginx-template.tmpl"
	NginxPID          = "/run/nginx.pid"
)

type HandlerInterface interface {
	GetPid(string) (string, error)
	CheckHaproxyConfig() (string, error)
	CheckNginxConfig() (string, error)
	CheckKeepalivedConfig() (string, error)
	StartHaproxy() (string, error)
	StartNginx() (string, error)
	StartKeepalived() (string, error)
	ReloadHaproxy(string) (string, error)
	ReloadNginx(string) (string, error)
	ReloadKeepalived(string) (string, error)
	StopHaproxy(string) error
	StopNginx(string) error
	StopKeepalived(string) error
}

type Handler struct{}

func (h *Handler) GetPid(fileName string) (string, error) {
	cmd := exec.Command("cat", fileName)
	pid, err := cmd.Output()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get pid")
		return "", err
	}
	pidStr := string(pid[:len(pid)-1])

	if _, err := strconv.Atoi(pidStr); err != nil {
		return "", err
	}

	return pidStr, nil
}

func (h *Handler) CheckHaproxyConfig() (string, error) {
	cmd := exec.Command("haproxy", "-c", "-V", "-f", HaproxyConfigFile)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Log.Reason(err).Errorf("haproxy configuration test failed")
		return "", err
	}

	return string(stdoutStderr), nil
}

func (h *Handler) CheckNginxConfig() (string, error) {
	cmd := exec.Command("nginx", "-t")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Log.Reason(err).Errorf("nginx configuration test failed")
		return "", err
	}

	return string(stdoutStderr), nil
}

// TODO: Need to check this
func (h *Handler) CheckKeepalivedConfig() (string, error) {
	return "", nil
}

func (h *Handler) StartHaproxy() (string, error) {
	cmd := exec.Command("haproxy", "-f", HaproxyConfigFile, "-p", HaproxyPID)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to start haproxy")
		return "", err
	}

	log.Log.Infof("haproxy process started output %s", stdoutStderr)

	return h.GetPid(HaproxyPID)
}

func (h *Handler) StartNginx() (string, error) {
	cmd := exec.Command("nginx")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to start nginx")
		return "", err
	}

	log.Log.Infof("nginx process started output %s", stdoutStderr)

	return h.GetPid(NginxPID)
}

func (h *Handler) StartKeepalived() (string, error) {
	cmd := exec.Command("keepalived", "--log-console", "--log-detail")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	log.Log.Infof("keepalived process started output %s", stdoutStderr)

	return h.GetPid(KeepalivedPID)
}

func (h *Handler) ReloadHaproxy(pid string) (string, error) {
	cmd := exec.Command("haproxy", "-f", HaproxyConfigFile, "-p", HaproxyPID, "-sf", pid)
	stdoutStderr, err := cmd.Output()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to reaload haproxy engine")
		return "", err
	}

	log.Log.Infof("haproxy engine reloaded output %s", string(stdoutStderr))

	return h.GetPid(HaproxyPID)
}

func (h *Handler) ReloadNginx(pid string) (string, error) {
	cmd := exec.Command("nginx", "-s", "reload")
	stdoutStderr, err := cmd.Output()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to reaload nginx engine")
		return "", err
	}

	log.Log.Infof("nginx engine reloaded output %s", string(stdoutStderr))

	return h.GetPid(NginxPID)
}

func (h *Handler) ReloadKeepalived(pid string) (string, error) {
	cmd := exec.Command("kill", "-HUP", pid)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to reaload keepalived engine")
		return "", err
	}

	log.Log.Infof("keepalived process configuration reloaded output %s", stdoutStderr)

	return h.GetPid(KeepalivedPID)
}

func (h *Handler) StopHaproxy(pid string) error {
	log.Log.Infof("stoping haproxy process with pid %s", pid)
	cmd := exec.Command("kill", pid)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	log.Log.Infof("haproxy stopped output %s", string(stdoutStderr))

	err = os.Remove(HaproxyPID)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to remove %s file", HaproxyPID)
	}

	return nil
}

func (h *Handler) StopNginx(pid string) error {
	log.Log.Infof("stoping nginx process with pid %s", pid)
	cmd := exec.Command("nginx", "-s", "stop")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	log.Log.Infof("nginx stopped output %s", string(stdoutStderr))

	err = os.Remove(NginxPID)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to remove %s file", HaproxyPID)
	}

	return nil
}

func (h *Handler) StopKeepalived(pid string) error {
	cmd := exec.Command("kill", pid)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to stop keepalived error %v", err)
		return err
	}

	log.Log.Infof("keepalived stoped output %s", string(stdoutStderr))

	err = os.Remove(KeepalivedPID)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to remove %s file", KeepalivedPID)
	}

	return nil
}
