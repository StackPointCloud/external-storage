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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ProfitbricksCredentials struct {
	datacenter string
	user       string
	password   string
}

// GetCredentialsFromSecret locates and returns credentials from kubernetes secret
func GetCredentialsFromSecret(client kubernetes.Interface, credentialsNamespace, credentialsDatacenter, credentialsSecret, credentialsUser string, credentialsPassword string) (*ProfitbricksCredentials, error) {

	credentials := &ProfitbricksCredentials{
		datacenter: credentialsDatacenter,
		user:       credentialsUser,
		password:   credentialsPassword,
	}

	if client == nil {
		return nil, fmt.Errorf("Kubernetes client not present")
	}
	secrets, err := client.Core().Secrets(credentialsNamespace).Get(credentialsSecret, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	datacenter, ok := secrets.Data[credentialsDatacenter]
	if !ok {
		return nil, fmt.Errorf("Cannot find Profitbricks Datacenter at secret %s/%s/%s", credentialsNamespace, credentialsSecret, credentialsDatacenter)
	}

	user, ok := secrets.Data[credentialsUser]
	if !ok {
		return nil, fmt.Errorf("Cannot find Profitbricks User at secret %s/%s/%s", credentialsNamespace, credentialsSecret, credentialsUser)
	}

	password, ok := secrets.Data[credentialsPassword]
	if !ok {
		return nil, fmt.Errorf("Cannot find Profitbricks Password at secret %s/%s/%s", credentialsNamespace, credentialsSecret, credentialsPassword)
	}

	credentials.datacenter = string(datacenter)
	credentials.user = string(user)
	credentials.password = string(password)

	return credentials, nil
}
