package volume

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetTokenFromSecret locates and returns token from kubernetes secret
func GetTokenFromSecret(client kubernetes.Interface, tokenNamespace, tokenSecret, tokenKey string) (string, error) {

	if client == nil {
		return "", fmt.Errorf("Kubernetes client not present")
	}
	secrets, err := client.Core().Secrets(tokenNamespace).Get(tokenSecret, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	token, ok := secrets.Data[tokenKey]
	if !ok {
		return "", fmt.Errorf("Cannot find Digital Ocean token at secret %s/%s/%s", tokenNamespace, tokenSecret, tokenKey)
	}

	return string(token), nil
}
