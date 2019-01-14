package cluster_controller

import (
	"fmt"
	"math/big"
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
	ips, err := getsHosts(&clusterObject.Spec)
	if err != nil {
		return nil, err
	}

	log.Log.Infof("cluster %s number of ip addresses to allocate from cird %d", clusterObject.Name, len(ips))
	log.Log.Infof("cluster %s number of allocated ip addresses %d", clusterObject.Name, len(clusterObject.Status.AllocatedIps))
	log.Log.Infof("cluster %s number of free ip addresses %d", clusterObject.Name, len(ips)-len(clusterObject.Status.AllocatedIps))

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
	} else if !isFarmInList(clusterObject.Status.AllocatedNamespaces[farm.Spec.ServiceNamespace].Farms, farm.Name) {
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

func getsHosts(clusterSpec *v1.ClusterSpec) ([]string, error) {
	var ips []string
	ipAddr, ipnet, err := net.ParseCIDR(clusterSpec.Subnet)
	if err != nil {
		return nil, err
	}

	// Can't create an allocator for a network with no addresses, eg
	// a /32 or /31
	ones, masklen := ipnet.Mask.Size()
	if ones > masklen-2 {
		return ips, fmt.Errorf("Network %s too small to allocate from", (*ipnet).String())
	}

	if err := canonicalizeIP(&ipAddr); err != nil {
		return ips, err
	}

	if len(ipAddr) != len(ipnet.Mask) {
		return ips, fmt.Errorf("IPNet IP and Mask version mismatch")
	}

	// Ensure Subnet IP is the network address, not some other address
	networkIP := ipAddr.Mask(ipnet.Mask)
	if !ipAddr.Equal(networkIP) {
		return ips, fmt.Errorf("Network has host bits set. For a subnet mask of length %d the network address is %s", ones, networkIP.String())
	}

	// RangeStart: If specified, make sure it's sane (inside the subnet),
	// otherwise use the first free IP (i.e. .1) - this will conflict with the
	// gateway but we skip it in the iterator
	var rangeStartIP net.IP
	if clusterSpec.RangeStart != "" {
		rangeStartIP = net.ParseIP(clusterSpec.RangeStart)
		if err := canonicalizeIP(&rangeStartIP); err != nil {
			return ips, err
		}

		if !ipnet.Contains(rangeStartIP) {
			return ips, fmt.Errorf("RangeStart %s not in network %s", clusterSpec.RangeStart, (*ipnet).String())
		}
	} else {
		rangeStartIP = nextIP(ipAddr)
		clusterSpec.RangeStart = rangeStartIP.String()
	}

	// RangeEnd: If specified, verify sanity. Otherwise, add a sensible default
	// (e.g. for a /24: .254 if IPv4, ::255 if IPv6)
	var rangeEndIP net.IP
	if clusterSpec.RangeEnd != "" {
		rangeEndIP = net.ParseIP(clusterSpec.RangeEnd)
		if err := canonicalizeIP(&rangeEndIP); err != nil {
			return ips, err
		}

		if !ipnet.Contains(rangeEndIP) {
			return ips, fmt.Errorf("RangeEnd %s not in network %s", clusterSpec.RangeEnd, (*ipnet).String())
		}
	} else {
		rangeEndIP = lastIP(ipnet)
		clusterSpec.RangeEnd = rangeEndIP.String()
	}

	for ipIdx := rangeStartIP; ipIdx.String() != rangeEndIP.String(); inc(ipIdx) {
		ips = append(ips, ipIdx.String())
	}

	ips = append(ips, rangeEndIP.String())
	// remove network address and broadcast address
	return ips, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// nextIP returns IP incremented by 1
func nextIP(ip net.IP) net.IP {
	i := ipToInt(ip)
	return intToIP(i.Add(i, big.NewInt(1)))
}

func lastIP(subnet *net.IPNet) net.IP {
	var end net.IP
	for i := 0; i < len(subnet.IP); i++ {
		end = append(end, subnet.IP[i]|^subnet.Mask[i])
	}
	if subnet.IP.To4() != nil {
		end[3]--
	}

	return end
}

func ipToInt(ip net.IP) *big.Int {
	if v := ip.To4(); v != nil {
		return big.NewInt(0).SetBytes(v)
	}
	return big.NewInt(0).SetBytes(ip.To16())
}

func intToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes())
}

// canonicalizeIP makes sure a provided ip is in standard form
func canonicalizeIP(ip *net.IP) error {
	if ip.To4() != nil {
		*ip = ip.To4()
		return nil
	} else if ip.To16() != nil {
		*ip = ip.To16()
		return nil
	}
	return fmt.Errorf("IP %s not v4 nor v6", *ip)
}

func isFarmInList(farms []string, farmName string) bool {
	for _, farm := range farms {
		if farm == farmName {
			return true
		}
	}

	return false
}
