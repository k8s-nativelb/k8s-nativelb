// +build !ignore_autogenerated

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
// Code generated by main. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Agent) DeepCopyInto(out *Agent) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Agent.
func (in *Agent) DeepCopy() *Agent {
	if in == nil {
		return nil
	}
	out := new(Agent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Agent) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AgentList) DeepCopyInto(out *AgentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Agent, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AgentList.
func (in *AgentList) DeepCopy() *AgentList {
	if in == nil {
		return nil
	}
	out := new(AgentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AgentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AgentSpec) DeepCopyInto(out *AgentSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AgentSpec.
func (in *AgentSpec) DeepCopy() *AgentSpec {
	if in == nil {
		return nil
	}
	out := new(AgentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AgentStatus) DeepCopyInto(out *AgentStatus) {
	*out = *in
	if in.LoadBalancer != nil {
		in, out := &in.LoadBalancer, &out.LoadBalancer
		*out = new(LoadBalancer)
		(*in).DeepCopyInto(*out)
	}
	in.LastUpdate.DeepCopyInto(&out.LastUpdate)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AgentStatus.
func (in *AgentStatus) DeepCopy() *AgentStatus {
	if in == nil {
		return nil
	}
	out := new(AgentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AllocatedNamespace) DeepCopyInto(out *AllocatedNamespace) {
	*out = *in
	if in.Farms != nil {
		in, out := &in.Farms, &out.Farms
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AllocatedNamespace.
func (in *AllocatedNamespace) DeepCopy() *AllocatedNamespace {
	if in == nil {
		return nil
	}
	out := new(AllocatedNamespace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackendSpec) DeepCopyInto(out *BackendSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackendSpec.
func (in *BackendSpec) DeepCopy() *BackendSpec {
	if in == nil {
		return nil
	}
	out := new(BackendSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackendStatus) DeepCopyInto(out *BackendStatus) {
	*out = *in
	out.HaproxyStatus = in.HaproxyStatus
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackendStatus.
func (in *BackendStatus) DeepCopy() *BackendStatus {
	if in == nil {
		return nil
	}
	out := new(BackendStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cluster) DeepCopyInto(out *Cluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cluster.
func (in *Cluster) DeepCopy() *Cluster {
	if in == nil {
		return nil
	}
	out := new(Cluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Cluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterList) DeepCopyInto(out *ClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Cluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterList.
func (in *ClusterList) DeepCopy() *ClusterList {
	if in == nil {
		return nil
	}
	out := new(ClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSpec) DeepCopyInto(out *ClusterSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSpec.
func (in *ClusterSpec) DeepCopy() *ClusterSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterStatus) DeepCopyInto(out *ClusterStatus) {
	*out = *in
	if in.Agents != nil {
		in, out := &in.Agents, &out.Agents
		*out = make(map[string]*Agent, len(*in))
		for key, val := range *in {
			var outVal *Agent
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Agent)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	if in.AllocatedIps != nil {
		in, out := &in.AllocatedIps, &out.AllocatedIps
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.AllocatedNamespaces != nil {
		in, out := &in.AllocatedNamespaces, &out.AllocatedNamespaces
		*out = make(map[string]*AllocatedNamespace, len(*in))
		for key, val := range *in {
			var outVal *AllocatedNamespace
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(AllocatedNamespace)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	in.LastUpdate.DeepCopyInto(&out.LastUpdate)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterStatus.
func (in *ClusterStatus) DeepCopy() *ClusterStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Farm) DeepCopyInto(out *Farm) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Farm.
func (in *Farm) DeepCopy() *Farm {
	if in == nil {
		return nil
	}
	out := new(Farm)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Farm) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FarmList) DeepCopyInto(out *FarmList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Farm, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FarmList.
func (in *FarmList) DeepCopy() *FarmList {
	if in == nil {
		return nil
	}
	out := new(FarmList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FarmList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FarmSpec) DeepCopyInto(out *FarmSpec) {
	*out = *in
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]corev1.ServicePort, len(*in))
		copy(*out, *in)
	}
	if in.Servers != nil {
		in, out := &in.Servers, &out.Servers
		*out = make(map[string]*ServerSpec, len(*in))
		for key, val := range *in {
			var outVal *ServerSpec
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(ServerSpec)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FarmSpec.
func (in *FarmSpec) DeepCopy() *FarmSpec {
	if in == nil {
		return nil
	}
	out := new(FarmSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FarmStatus) DeepCopyInto(out *FarmStatus) {
	*out = *in
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FarmStatus.
func (in *FarmStatus) DeepCopy() *FarmStatus {
	if in == nil {
		return nil
	}
	out := new(FarmStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Haproxy) DeepCopyInto(out *Haproxy) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Haproxy.
func (in *Haproxy) DeepCopy() *Haproxy {
	if in == nil {
		return nil
	}
	out := new(Haproxy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HaproxyStatus) DeepCopyInto(out *HaproxyStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HaproxyStatus.
func (in *HaproxyStatus) DeepCopy() *HaproxyStatus {
	if in == nil {
		return nil
	}
	out := new(HaproxyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HealthCheck) DeepCopyInto(out *HealthCheck) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HealthCheck.
func (in *HealthCheck) DeepCopy() *HealthCheck {
	if in == nil {
		return nil
	}
	out := new(HealthCheck)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Keepalived) DeepCopyInto(out *Keepalived) {
	*out = *in
	if in.InstancesStatus != nil {
		in, out := &in.InstancesStatus, &out.InstancesStatus
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Keepalived.
func (in *Keepalived) DeepCopy() *Keepalived {
	if in == nil {
		return nil
	}
	out := new(Keepalived)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadBalancer) DeepCopyInto(out *LoadBalancer) {
	*out = *in
	if in.Haproxy != nil {
		in, out := &in.Haproxy, &out.Haproxy
		*out = new(Haproxy)
		**out = **in
	}
	if in.Nginx != nil {
		in, out := &in.Nginx, &out.Nginx
		*out = new(Nginx)
		**out = **in
	}
	if in.Keepalived != nil {
		in, out := &in.Keepalived, &out.Keepalived
		*out = new(Keepalived)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadBalancer.
func (in *LoadBalancer) DeepCopy() *LoadBalancer {
	if in == nil {
		return nil
	}
	out := new(LoadBalancer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Nginx) DeepCopyInto(out *Nginx) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Nginx.
func (in *Nginx) DeepCopy() *Nginx {
	if in == nil {
		return nil
	}
	out := new(Nginx)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Server) DeepCopyInto(out *Server) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Server.
func (in *Server) DeepCopy() *Server {
	if in == nil {
		return nil
	}
	out := new(Server)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Server) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerData) DeepCopyInto(out *ServerData) {
	*out = *in
	if in.Server != nil {
		in, out := &in.Server, &out.Server
		*out = new(Server)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerData.
func (in *ServerData) DeepCopy() *ServerData {
	if in == nil {
		return nil
	}
	out := new(ServerData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerList) DeepCopyInto(out *ServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Server, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerList.
func (in *ServerList) DeepCopy() *ServerList {
	if in == nil {
		return nil
	}
	out := new(ServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerSpec) DeepCopyInto(out *ServerSpec) {
	*out = *in
	if in.UDP != nil {
		in, out := &in.UDP, &out.UDP
		*out = new(UDP)
		**out = **in
	}
	if in.TCP != nil {
		in, out := &in.TCP, &out.TCP
		*out = new(TCP)
		**out = **in
	}
	if in.Backends != nil {
		in, out := &in.Backends, &out.Backends
		*out = make(map[string]*BackendSpec, len(*in))
		for key, val := range *in {
			var outVal *BackendSpec
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(BackendSpec)
				**out = **in
			}
			(*out)[key] = outVal
		}
	}
	if in.HealthCheck != nil {
		in, out := &in.HealthCheck, &out.HealthCheck
		*out = new(HealthCheck)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerSpec.
func (in *ServerSpec) DeepCopy() *ServerSpec {
	if in == nil {
		return nil
	}
	out := new(ServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerStatus) DeepCopyInto(out *ServerStatus) {
	*out = *in
	if in.Frontend != nil {
		in, out := &in.Frontend, &out.Frontend
		*out = new(HaproxyStatus)
		**out = **in
	}
	if in.Backend != nil {
		in, out := &in.Backend, &out.Backend
		*out = new(HaproxyStatus)
		**out = **in
	}
	if in.Backends != nil {
		in, out := &in.Backends, &out.Backends
		*out = make([]*HaproxyStatus, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(HaproxyStatus)
				**out = **in
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerStatus.
func (in *ServerStatus) DeepCopy() *ServerStatus {
	if in == nil {
		return nil
	}
	out := new(ServerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TCP) DeepCopyInto(out *TCP) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TCP.
func (in *TCP) DeepCopy() *TCP {
	if in == nil {
		return nil
	}
	out := new(TCP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UDP) DeepCopyInto(out *UDP) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UDP.
func (in *UDP) DeepCopy() *UDP {
	if in == nil {
		return nil
	}
	out := new(UDP)
	in.DeepCopyInto(out)
	return out
}
