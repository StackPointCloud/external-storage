package volume

import (
	"fmt"

	"github.com/kubernetes-incubator/external-storage/oneandone/flex-volume/pkg/cloud"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetCredentialsFromSecret is a
func GetCredentialsFromSecret(client kubernetes.Interface, namespace string, secretName string, datacenterKey string, tokenKey string) (*cloud.OneandoneCredentials, error) {

	if client == nil {
		return nil, fmt.Errorf("Kubernetes client not present")
	}
	secrets, err := client.Core().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	datacenter, ok := secrets.Data[datacenterKey]
	if !ok {
		return nil, fmt.Errorf("Cannot find 1&1 DatacenterId at secret %s/%s/%s", namespace, secretName, datacenterKey)
	}

	token, ok := secrets.Data[tokenKey]
	if !ok {
		return nil, fmt.Errorf("Cannot find 1&1 token at secret %s/%s/%s", namespace, secretName, tokenKey)
	}

	return &cloud.OneandoneCredentials{
		Token:        string(token),
		DatacenterID: string(datacenter),
	}, nil
}
