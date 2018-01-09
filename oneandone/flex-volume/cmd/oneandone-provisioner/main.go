package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/kubernetes-incubator/external-storage/oneandone/flex-volume/pkg/cloud"
	"github.com/kubernetes-incubator/external-storage/oneandone/flex-volume/pkg/volume"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultProvisioner           = "oneandone/flex-volume-provisioner"
	defaultFlexDriver            = "stackpoint/oneandone"
	defaultcredentialsNamespace  = "kube-system"
	defaultCredentialsSecret     = "oneandone"
	defaultCredentialsToken      = "token"
	defaultCredentialsDatacenter = "908DC2072407C94C8054610AD5A53B8C"
)

var (
	provisioner           = flag.String("provisioner", defaultProvisioner, "Name of the provisioner. The provisioner will only provision volumes for claims that request a StorageClass with a provisioner field set equal to this name.")
	master                = flag.String("master", "", "Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.")
	kubeconfig            = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Either this or master needs to be set if the provisioner is being run out of cluster.")
	flexDriver            = flag.String("flex-driver", defaultFlexDriver, "Flex driver name for the generated volume")
	credentialsNamespace  = flag.String("credentials-namespace", defaultcredentialsNamespace, "Namespace for the 1&1 credentials secret")
	credentialsSecret     = flag.String("credentials-secret", defaultCredentialsSecret, "Secret name for the 1&1 secret")
	credentialsToken      = flag.String("credentials-token", defaultCredentialsToken, "Secret key for the 1&1 Token")
	credentilasDatacenter = flag.String("credentials-datacenter", defaultCredentialsDatacenter, "1&1 Datacenter")
)

func main() {
	var config *rest.Config
	var err error

	flag.Set("logtostderr", "true")
	flag.Parse()

	glog.Info("Starting 1&1 dynamic provisioner")

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

	// Create the 1&1 manager
	credentials, err := volume.GetCredentialsFromSecret(clientset, *credentialsNamespace, *credentialsSecret, *credentilasDatacenter, *credentialsToken)
	if err != nil {
		glog.Fatalf("Error retrieving 1&1 credentials: %v", err.Error())
	}

	glog.Info("Creating 1&1 client")
	oneandoneManager, err := cloud.NewOneandoneManager(credentials)
	if err != nil {
		glog.Fatalf("Error creating 1&1 client: %v", err.Error())
	}

	// Create the provisioner
	glog.Infof("Creating 1&1 provisioner %q", *provisioner)

	oneandoneProvisioner, err := volume.NewOneandoneProvisioner(clientset, oneandoneManager, *flexDriver)
	if err != nil {
		glog.Fatalf("Error creating 1&1 provisioner: %v", err.Error())
	}

	// Start the provision controller
	pc := controller.NewProvisionController(
		clientset,
		*provisioner,
		oneandoneProvisioner,
		serverVersion.GitVersion,
	)
	pc.Run(wait.NeverStop)

}
