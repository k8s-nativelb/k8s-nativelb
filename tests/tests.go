package tests

import (
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestClient struct {
	kubeClient *kubernetes.Clientset
	client client.Client
}


func NewTestClient()(*TestClient, error) {
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		return nil,err
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		return nil,err
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil,err
	}

	return &TestClient{kubeClient:kubeClient,client:mgr.GetClient()}, nil
}

func (t *TestClient)CreateTestNamespace(){
	testNamespaceObject := &corev1.Namespace{ObjectMeta:metav1.ObjectMeta{Name:TestNamespace}}
	_,err := t.kubeClient.CoreV1().Namespaces().Create(testNamespaceObject)
	PanicOnError(err)
}

func (t *TestClient)DeleteTestNamespace() (error){
	return t.kubeClient.CoreV1().Namespaces().Delete(TestNamespace,&metav1.DeleteOptions{})
}

func (t *TestClient)GetTestNamespace() (*corev1.Namespace, error) {
	return t.kubeClient.CoreV1().Namespaces().Get(TestNamespace,metav1.GetOptions{})
}