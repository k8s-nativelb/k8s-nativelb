package udp_loadbalancer

import (
	"fmt"
	"github.com/k8s-nativelb/pkg/proto"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

const (
	nginxPageUrl = "http://127.0.0.1/nginx_status"
)

func (u *UdpLoadBalancer) info() (*proto.NginxStatus, error) {
	resp, err := http.Get(nginxPageUrl)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		if err != nil {
			data = []byte(err.Error())
		}
		return nil, fmt.Errorf("nginx Status %s (%d): %s", resp.Status, resp.StatusCode, data)
	}

	nginxStatus := &proto.NginxStatus{}

	// Parsing results
	lines := strings.Split(string(data), "\n")
	if len(lines) != 5 {
		return nil, fmt.Errorf("nginx unexpected number of lines in status: %v", lines)
	}

	// active connections
	parts := strings.Split(lines[0], ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("nginx unexpected first line: %s", lines[0])
	}

	v, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, err
	}
	nginxStatus.ActiveConnections = uint64(v)

	// current connections
	parts = strings.Fields(lines[3])
	if len(parts) != 6 {
		return nil, fmt.Errorf("Unexpected fourth line: %s", lines[3])
	}
	v, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, err
	}
	nginxStatus.Reading = uint64(v)

	v, err = strconv.Atoi(strings.TrimSpace(parts[3]))
	if err != nil {
		return nil, err
	}
	nginxStatus.Writing = uint64(v)

	v, err = strconv.Atoi(strings.TrimSpace(parts[5]))
	if err != nil {
		return nil, err
	}
	nginxStatus.Waiting = uint64(v)

	nginxStatus.Pid = uint64(u.GetPid())

	cmd := exec.Command("sh", "-c", "nginx -v")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	splitLine := strings.Split(string(stdoutStderr), "/")
	if len(splitLine) != 2 {
		return nil, fmt.Errorf("nginx Unexpected version output")
	}
	nginxStatus.Version = strings.Replace(splitLine[1], "\n", "", -1)

	return nginxStatus, nil
}
