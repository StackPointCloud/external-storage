/*
Copyright 2017 The Kubernetes Authors.

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

package volume

import (
	"fmt"

	"github.com/external-storage/profitbricks/flex-volume/pkg/cloud"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetCredentialsFromSecret locates and returns credentials from kubernetes secret
func GetCredentialsFromSecret(client kubernetes.Interface, namespace string, secretName string, datacenterKey string, userKey string, passwordKey string) (*cloud.ProfitbricksCredentials, error) {

	if client == nil {
		return nil, fmt.Errorf("Kubernetes client not present")
	}
	secrets, err := client.Core().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	datacenter, ok := secrets.Data[datacenterKey]
	if !ok {
		return nil, fmt.Errorf("Cannot find Profitbricks Datacenter at secret %s/%s/%s", namespace, secretName, datacenterKey)
	}

	user, ok := secrets.Data[userKey]
	if !ok {
		return nil, fmt.Errorf("Cannot find Profitbricks User at secret %s/%s/%s", namespace, secretName, userKey)
	}

	password, ok := secrets.Data[passwordKey]
	if !ok {
		return nil, fmt.Errorf("Cannot find Profitbricks Password at secret %s/%s/%s", namespace, secretName, passwordKey)
	}

	return &cloud.ProfitbricksCredentials{
		Datacenter: string(datacenter),
		User:       string(user),
		Password:   string(password),
	}, nil
}
