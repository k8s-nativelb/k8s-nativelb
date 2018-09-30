package cluster_controller

import (
	"fmt"
	"net"
	"sync"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
)

type Allocator struct {
	mutex sync.Mutex
	ips   []string
}

func NewAllocator(clusterObject *v1.Cluster) (*Allocator, error) {
	ips, err := getsHosts(clusterObject.Spec.IpRange)
	if err != nil {
		return nil, err
	}
	log.Log.Infof("number of ip addresses to allocate from cird %d", len(ips))

	log.Log.Infof("number of allocated ip addresses %d", len(clusterObject.Status.AllocatedIps))

	log.Log.Infof("number of free ip addresses %d", len(ips)-len(clusterObject.Status.AllocatedIps))

	return &Allocator{ips: ips, mutex: sync.Mutex{}}, nil
}

func (a *Allocator) Allocate(farm *v1.Farm, clusterObject *v1.Cluster) (string, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	ipAddr, isExist := a.AllocatedFarm(farm, clusterObject)
	if isExist {
		return ipAddr, nil
	}

	ipAddr, err := a.findFreeIpAddr(clusterObject)
	if err != nil {
		return "", err
	}

	farm.Status.IpAdress = ipAddr
	if clusterObject.Status.AllocatedIps == nil {
		clusterObject.Status.AllocatedIps = make(map[string]string)
	}

	// Add to allocatedIps
	clusterObject.Status.AllocatedIps[ipAddr] = farm.Name

	return ipAddr, nil
}

func (a *Allocator) Release(ipAddr string, clusterObject *v1.Cluster) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	_, ok := clusterObject.Status.AllocatedIps[ipAddr]

	if !ok {
		log.Log.Errorf("fail to find %s in allocated ip map", ipAddr)
		return
	}

	delete(clusterObject.Status.AllocatedIps, ipAddr)
}

func (a *Allocator) findFreeIpAddr(clusterObject *v1.Cluster) (string, error) {
	if clusterObject.Status.AllocatedIps == nil {
		return a.ips[0], nil
	}

	for _, value := range a.ips {
		if _, ok := clusterObject.Status.AllocatedIps[value]; !ok {
			return value, nil
		}
	}

	return "", fmt.Errorf("fail to find any free address")
}

func (a *Allocator) AllocatedFarm(farm *v1.Farm, clusterObject *v1.Cluster) (string, bool) {
	for allocatedIp := range clusterObject.Status.AllocatedIps {
		if clusterObject.Status.AllocatedIps[allocatedIp] == farm.Name {
			return allocatedIp, true
		}
	}

	return "", false
}

func getsHosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
