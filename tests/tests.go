package tests

import (
	"fmt"
	nativelb "github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"k8s.io/client-go/rest"

	"github.com/k8s-nativelb/pkg/kubecli"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

type TestClient struct {
	KubeClient     *kubernetes.Clientset
	Client         client.Client
	NativelbClient kubecli.NativelbClient
}

func NewTestClient() (*TestClient, error) {
	t := &envtest.Environment{
		UseExistingCluster: true,
	}

	err := nativelb.SchemeBuilder.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}

	var cfg *rest.Config
	var c client.Client

	if cfg, err = t.Start(); err != nil {
		return nil, err
	}

	if c, err = client.New(cfg, client.Options{Scheme: scheme.Scheme}); err != nil {
		return nil, err
	}

	KubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	nativelbClient, err := kubecli.GetNativelbClient(cfg, c)
	if err != nil {
		return nil, err
	}

	return &TestClient{KubeClient: KubeClient, Client: nativelbClient.GetClient(), NativelbClient: nativelbClient}, nil
}

func (t *TestClient) CreateTestNamespace() {
	testNamespaceObject := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: TestNamespace}}
	_, err := t.KubeClient.CoreV1().Namespaces().Create(testNamespaceObject)
	PanicOnError(err)
}

// TODO: Finnish this
func (t *TestClient) CleanNativelbNamespace() {
	err := t.deleteFarms()
	PanicOnError(err)
}

func (t *TestClient) deleteFarms() error {
	farms, err := t.NativelbClient.Farm("").List(&client.ListOptions{})
	if err != nil {
		return err
	}

	for _, farm := range farms.Items {
		err = t.NativelbClient.Farm(farm.Namespace).Delete(farm.Name)
		if err != nil {
			return err
		}
	}

	for i := 0; i < 10; i++ {
		farms, err = t.NativelbClient.Farm("").List(&client.ListOptions{})
		if err != nil {
			return err
		}

		if len(farms.Items) == 0 {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("Fail to remove all the farms")
}

func (t *TestClient) DeleteTestNamespace() error {
	return t.KubeClient.CoreV1().Namespaces().Delete(TestNamespace, &metav1.DeleteOptions{})
}

func (t *TestClient) GetTestNamespace() (*corev1.Namespace, error) {
	return t.KubeClient.CoreV1().Namespaces().Get(TestNamespace, metav1.GetOptions{})
}

func (t *TestClient) GetNativeLBNamespace() (*corev1.Namespace, error) {
	return t.KubeClient.CoreV1().Namespaces().Get(nativelb.ControllerNamespace, metav1.GetOptions{})
}
