package cluster_controller

import (
	"fmt"
	"net"
	"sync"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
)

type Allocator struct {
	mutex         *sync.Mutex
	ips           []string
}

func NewAllocator(clusterObject v1.Cluster) (*Allocator, error) {
	ips, err := getsHosts(clusterObject.Spec.IpRange)
	if err != nil {
		return nil, err
	}
	log.Log.Infof("number of ip addresses to allocate from cird %d", len(ips))

	log.Log.Infof("number of allocated ip addresses %d", len(clusterObject.Status.AllocatedIps))

	log.Log.Infof("number of free ip addresses %d", len(ips)-len(clusterObject.Status.AllocatedIps))

	return &Allocator{ips: ips}, nil
}

func (a *Allocator) Allocate(farm *v1.Farm,clusterObject *v1.Cluster) (string, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	ipAddr, err := a.findFreeIpAddr(clusterObject)
	if err != nil {
		return "", err
	}

	err = a.updateAllocatedIps(ipAddr, farm, clusterObject)
	return ipAddr, err
}

func (a *Allocator) Release(ipAddr string,clusterObject *v1.Cluster) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	_, ok := clusterObject.Status.AllocatedIps[ipAddr]

	if !ok {
		log.Log.Errorf("fail to find %s in allocated ip map", ipAddr)
		return
	}

	delete(clusterObject.Status.AllocatedIps, ipAddr)
}

func (a *Allocator) updateAllocatedIps(ipAddr string, farm *v1.Farm,clusterObject *v1.Cluster) error {
	farm.Status.IpAdress = ipAddr
	servers := farm.Spec.Servers
	if clusterObject.Status.AllocatedIps == nil {
		clusterObject.Status.AllocatedIps = make(map[string]*v1.Farm)
	}

	// Update server bindings
	for idx := range servers {
		servers[idx].Spec.Bind = fmt.Sprintf("%s:%s", ipAddr, servers[idx].Spec.Bind)
	}

	// Add to allocatedIps
	clusterObject.Status.AllocatedIps[ipAddr] = farm

	return nil
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
