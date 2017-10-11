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

package main

import (
	"flag"

	"github.com/external-storage/profitbricks/flex-volume/pkg/cloud"
	"github.com/external-storage/profitbricks/flex-volume/pkg/volume"
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultProvisioner           = "profitbricks/flex-volume-provisioner"
	defaultFlexDriver            = "stackpoint/profitbricks"
	defaultcredentialsDatacenter = "stackpointcloud"
	defaultcredentialsNamespace  = "kube-system"
	defaultCredentialsSecret     = "profitbricks"
	defaultCredentialsUser       = "username"
	defaultCredentialsPassword   = "password"
)

var (
	provisioner           = flag.String("provisioner", defaultProvisioner, "Name of the provisioner. The provisioner will only provision volumes for claims that request a StorageClass with a provisioner field set equal to this name.")
	master                = flag.String("master", "", "Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.")
	kubeconfig            = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Either this or master needs to be set if the provisioner is being run out of cluster.")
	flexDriver            = flag.String("flex-driver", defaultFlexDriver, "Flex driver name for the generated volume")
	credentialsDatacenter = flag.String("credentials-datacenter", defaultcredentialsDatacenter, "Datacenter for the Profitbricks credentials secret")
	credentialsNamespace  = flag.String("credentials-namespace", defaultcredentialsNamespace, "Namespace for the Profitbricks credentials secret")
	credentialsSecret     = flag.String("credentials-secret", defaultCredentialsSecret, "Secret name for the Profitbricks secret")
	credentialsUser       = flag.String("credentials-user", defaultCredentialsUser, "Secret key for the Profitbricks Username secret (base64 encoded)")
	credentialsPassword   = flag.String("credentials-password", defaultCredentialsPassword, "Secret key for the Profitbricks Password secret (base64 encoded)")
)

func main() {

	var config *rest.Config
	var err error

	flag.Set("logtostderr", "true")
	flag.Parse()

	glog.Info("Starting Profitbricks dynamic provisioner")

	if *master != "" || *kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		glog.Fatalf("Failed to create config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Failed to create kubernetes client: %s", err.Error())
	}

	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		glog.Fatalf("Error getting kubernetes server version: %s", err.Error())
	}

	// Maybe it's not necessary the datacenter here
	// Create the Profitbricks manager
	credentials, err := volume.GetCredentialsFromSecret(clientset, *credentialsNamespace, *credentialsDatacenter, *credentialsSecret, *credentialsUser, *credentialsPassword)
	if err != nil {
		glog.Fatalf("Error retrieving Profitbricks credentials: %v", err.Error())
	}

	glog.Info("Creating Profitbricks client")
	pb, err := cloud.NewProfitbricksManager(credentials)
	if err != nil {
		glog.Fatalf("Error creating Profitbricks client: %v", err.Error())
	}

	// Create the provisioner
	glog.Infof("Creating Profitbricks provisioner %q", *provisioner)
	profitbricksProvisioner, err := volume.NewProfitbricksProvisioner(clientset, pb, *flexDriver)
	if err != nil {
		glog.Fatalf("Error creating Profitbricks provisioner: %v", err.Error())
	}

	// Start the provision controller
	pc := controller.NewProvisionController(
		clientset,
		*provisioner,
		profitbricksProvisioner,
		serverVersion.GitVersion,
	)
	pc.Run(wait.NeverStop)
}
