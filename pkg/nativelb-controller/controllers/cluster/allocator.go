package cluster_controller

import (
	"fmt"
	"net"
	"sync"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
)

func (c *ClusterController) allocateIpAddrAndRouterID(farm *v1.Farm, cluster *v1.Cluster) error {
	_, isExist := c.allocator[cluster.Name]
	if !isExist {
		allocator, err := NewAllocator(cluster)
		if err != nil {
			return fmt.Errorf("failed to create allocator for cluster %s error %v", cluster.Name, err)
		}

		c.allocator[cluster.Name] = allocator
	}

	return c.allocator[cluster.Name].Allocate(farm, cluster)
}

func (c *ClusterController) releaseIpAddrAndRouterID(farm *v1.Farm, cluster *v1.Cluster) error {
	_, isExist := c.allocator[cluster.Name]
	if !isExist {
		allocator, err := NewAllocator(cluster)
		if err != nil {
			return fmt.Errorf("failed to create allocator for cluster %s error %v", cluster.Name, err)
		}

		c.allocator[cluster.Name] = allocator
	}

	c.allocator[cluster.Name].Release(farm, cluster)
	return nil
}

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

func (a *Allocator) Allocate(farm *v1.Farm, clusterObject *v1.Cluster) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	ipAddr, err := a.allocateIpAddr(farm, clusterObject)
	if err != nil {
		return err
	}

	routerID, err := a.allocateRouterID(farm, clusterObject)
	if err != nil {
		return err
	}

	farm.Status.IpAdress = ipAddr
	if clusterObject.Status.AllocatedIps == nil {
		clusterObject.Status.AllocatedIps = make(map[string]string)
	}

	// Add to allocatedIps
	clusterObject.Status.AllocatedIps[ipAddr] = farm.Name


	if clusterObject.Status.AllocatedNamespaces == nil {
		clusterObject.Status.AllocatedNamespaces = make(map[string]*v1.AllocatedNamespace)
	}

	// add to allocatedNamespaces
	if _, ok := clusterObject.Status.AllocatedNamespaces[farm.Spec.ServiceNamespace]; !ok {
		clusterObject.Status.AllocatedNamespaces[farm.Spec.ServiceNamespace] = &v1.AllocatedNamespace{RouterID: routerID, Farms: []string{farm.Name}}
	} else {
		clusterObject.Status.AllocatedNamespaces[farm.Spec.ServiceNamespace].Farms = append(clusterObject.Status.AllocatedNamespaces[farm.Spec.ServiceNamespace].Farms, farm.Name)
	}

	return nil
}

func (a *Allocator) Release(farm *v1.Farm, clusterObject *v1.Cluster) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	ipAddr := farm.Status.IpAdress
	_, ok := clusterObject.Status.AllocatedIps[ipAddr]

	if !ok {
		log.Log.Errorf("fail to find %s in allocated ip map", ipAddr)
	}

	delete(clusterObject.Status.AllocatedIps, ipAddr)

	a.removeFarmFromAllocatedRouterID(clusterObject, farm.Name, farm.Spec.ServiceNamespace)
}

func (a *Allocator) allocateIpAddr(farm *v1.Farm, clusterObject *v1.Cluster) (string, error) {
	ipAddr, isExist := a.AllocatedFarm(farm, clusterObject)
	if isExist {
		return ipAddr, nil
	}

	ipAddr, err := a.findFreeIpAddr(clusterObject)
	if err != nil {
		return "", err
	}

	return ipAddr, nil
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

func (a *Allocator) allocateRouterID(farm *v1.Farm, clusterObject *v1.Cluster) (int32, error) {
	if clusterObject.Status.AllocatedNamespaces != nil {
		if value, ok := clusterObject.Status.AllocatedNamespaces[farm.Spec.ServiceNamespace]; ok {
			return value.RouterID, nil
		}
	}

	routerID, err := a.findFreeRouterID(clusterObject)
	if err != nil {
		return 0, err
	}

	return routerID, nil
}

func (a *Allocator) findFreeRouterID(clusterObject *v1.Cluster) (int32, error) {
	routerIDs := make(map[int32]bool)
	for _, allocatedNamespace := range clusterObject.Status.AllocatedNamespaces {
		routerIDs[allocatedNamespace.RouterID] = true
	}

	for i := 1; i < 255; i++ {
		if _, ok := routerIDs[int32(i)]; !ok {
			return int32(i), nil
		}
	}

	return 0, fmt.Errorf("failed to allocated routerID")
}

func (a *Allocator) AllocatedFarm(farm *v1.Farm, clusterObject *v1.Cluster) (string, bool) {
	for allocatedIp := range clusterObject.Status.AllocatedIps {
		if clusterObject.Status.AllocatedIps[allocatedIp] == farm.Name {
			return allocatedIp, true
		}
	}

	return "", false
}

func (a *Allocator) removeFarmFromAllocatedRouterID(clusterObject *v1.Cluster, farmName, namespace string) {
	if len(clusterObject.Status.AllocatedNamespaces[namespace].Farms) == 1 {
		delete(clusterObject.Status.AllocatedNamespaces, namespace)
		return
	}

	for idx, AllocatedfarmName := range clusterObject.Status.AllocatedNamespaces[namespace].Farms {
		if AllocatedfarmName == farmName {
			clusterObject.Status.AllocatedNamespaces[namespace].Farms = append(clusterObject.Status.AllocatedNamespaces[namespace].Farms[:idx],
				clusterObject.Status.AllocatedNamespaces[namespace].Farms[idx+1:]...)
			return
		}
	}
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
