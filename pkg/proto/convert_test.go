package proto_test

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Convert", func() {

	createFarmsAndClusterData := func() ([]v1.Farm, *v1.Cluster) {
		cluster := &v1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "test-cluster"},
			Spec: v1.ClusterSpec{Internal: true, Default: true, Subnet: "10.0.0.0/24"},
			Status: v1.ClusterStatus{AllocatedNamespaces: map[string]*v1.AllocatedNamespace{"namespace1": {RouterID: 1}, "namespace2": {RouterID: 2}},
				AllocatedIps: map[string]string{"nginx1": "10.0.0.1", "nginx2": "10.0.0.2"}}}

		farms := []v1.Farm{{Spec: v1.FarmSpec{Cluster: "test-cluster", ServiceNamespace: "namespace1", ServiceName: "nginx1"}},
			{Spec: v1.FarmSpec{Cluster: "test-cluster", ServiceNamespace: "namespace2", ServiceName: "nginx2"}},
			{Spec: v1.FarmSpec{Cluster: "test-cluster", ServiceNamespace: "namespace1", ServiceName: "nginx3"}}}

		return farms, cluster
	}

	Context("ConvertFarmsToGrpcDataList", func() {
		It("should convert the right configuration for one agent", func() {
			farms, cluster := createFarmsAndClusterData()
			data := proto.ConvertFarmsToGrpcDataList(farms, cluster, 1, 1)
			Expect(len(data)).To(Equal(3))
			Expect(data[0].RouterID).To(Equal(int32(1)))
			Expect(data[1].RouterID).To(Equal(int32(2)))
			Expect(data[2].RouterID).To(Equal(int32(1)))
			Expect(data[0].Priority).To(Equal(int32(101)))
			Expect(data[1].Priority).To(Equal(int32(101)))
			Expect(data[2].Priority).To(Equal(int32(101)))
			Expect(data[0].KeepalivedState).To(Equal("MASTER"))
			Expect(data[1].KeepalivedState).To(Equal("MASTER"))
			Expect(data[2].KeepalivedState).To(Equal("MASTER"))
		})

		It("should convert the right configuration for multiple agents", func() {
			farms, cluster := createFarmsAndClusterData()
			data := proto.ConvertFarmsToGrpcDataList(farms, cluster, 1, 2)
			Expect(len(data)).To(Equal(3))
			Expect(data[0].RouterID).To(Equal(int32(1)))
			Expect(data[1].RouterID).To(Equal(int32(2)))
			Expect(data[2].RouterID).To(Equal(int32(1)))
			Expect(data[0].Priority).To(Equal(int32(1)))
			Expect(data[1].Priority).To(Equal(int32(101)))
			Expect(data[2].Priority).To(Equal(int32(1)))
			Expect(data[0].KeepalivedState).To(Equal("MASTER"))
			Expect(data[1].KeepalivedState).To(Equal("MASTER"))
			Expect(data[2].KeepalivedState).To(Equal("MASTER"))

			data = proto.ConvertFarmsToGrpcDataList(farms, cluster, 2, 2)
			Expect(len(data)).To(Equal(3))
			Expect(data[0].RouterID).To(Equal(int32(1)))
			Expect(data[1].RouterID).To(Equal(int32(2)))
			Expect(data[2].RouterID).To(Equal(int32(1)))
			Expect(data[0].Priority).To(Equal(int32(102)))
			Expect(data[1].Priority).To(Equal(int32(2)))
			Expect(data[2].Priority).To(Equal(int32(102)))
			Expect(data[0].KeepalivedState).To(Equal("BACKUP"))
			Expect(data[1].KeepalivedState).To(Equal("BACKUP"))
			Expect(data[2].KeepalivedState).To(Equal("BACKUP"))
		})
	})
})
