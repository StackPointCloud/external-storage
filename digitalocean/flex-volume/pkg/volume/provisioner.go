package volume

import (
	"fmt"
	"strconv"

	"github.com/digitalocean/godo"
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/digitalocean/flex-volume/pkg/cloud"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/kubernetes-incubator/external-storage/lib/gidallocator"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/volume"
)

const (
	volumeDescription   = "Kubernetes dynamic provisioned"
	volumeNameMaxLenght = 64
)

type digitalOceanProvisioner struct {
	client       kubernetes.Interface
	digitalocean *cloud.DigitalOceanManager
	token        string
	allocator    gidallocator.Allocator
	flexDriver   string
}

// NewDigitalOceanProvisioner creates a Digital Ocean volume provisioner
func NewDigitalOceanProvisioner(client kubernetes.Interface, do *cloud.DigitalOceanManager, flexDriver string) (controller.Provisioner, error) {

	return newDigitalOceanProvisionerInternal(client, do, flexDriver)
}

func newDigitalOceanProvisionerInternal(client kubernetes.Interface, do *cloud.DigitalOceanManager, flexDriver string) (*digitalOceanProvisioner, error) {

	if client == nil {
		return nil, fmt.Errorf("Provisioner needs the kubernetes client")
	}

	if do == nil {
		return nil, fmt.Errorf("Provisioner needs the Digital Ocean client")
	}

	return &digitalOceanProvisioner{
		client:       client,
		digitalocean: do,
		allocator:    gidallocator.New(client),
		flexDriver:   flexDriver,
	}, nil
}

var _ controller.Provisioner = &digitalOceanProvisioner{}

// Provision creates a volume i.e. the storage asset and returns a PV object for
// the volume.
func (p *digitalOceanProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	if options.PVC.Spec.Selector != nil {
		return nil, fmt.Errorf("claim.Spec.Selector is not supported")
	}

	gid, err := p.allocator.AllocateNext(options)
	if err != nil {
		return nil, err
	}

	vol, err := p.createVolume(options)
	if err != nil {
		return nil, err
	}

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName,
			Annotations: map[string]string{
				gidallocator.VolumeGidAnnotationKey: strconv.FormatInt(int64(gid), 10),
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FlexVolume: &v1.FlexVolumeSource{
					Driver: p.flexDriver,
					Options: map[string]string{
						"VolumeID":   vol.ID,
						"VolumeName": vol.Name,
					},
					ReadOnly: false,
				},
			},
		},
	}

	return pv, nil
}

// createVolume creates a volume at Digital Ocean
func (p *digitalOceanProvisioner) createVolume(options controller.VolumeOptions) (*godo.Volume, error) {

	name := generateVolumeName(options.PVName)
	capacity := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	sizeGB := int(volume.RoundUpSize(capacity.Value(), 1024*1024*1024))

	glog.V(5).Infof("Creating Digital Ocean volume %s sized %d GB", name, sizeGB)
	vol, err := p.digitalocean.CreateVolume(name, volumeDescription, sizeGB)
	if err != nil {
		return nil, err
	}

	return vol, nil
}

func generateVolumeName(name string) string {
	prefix := "kubernetes-dynamic"
	nameLen := len(name)

	if nameLen+1+len(prefix) > volumeNameMaxLenght {
		prefix = prefix[:volumeNameMaxLenght-nameLen-1]
	}
	return prefix + "-" + name
}

// Delete removes the directory that was created by Provision backing the given
// PV and removes its export from the NFS server.
func (p *digitalOceanProvisioner) Delete(volume *v1.PersistentVolume) error {
	err := p.allocator.Release(volume)
	if err != nil {
		return err
	}

	flx := volume.Spec.FlexVolume
	if flx == nil {
		return fmt.Errorf("Volume %s/%s is not a FlexVolume", volume.Namespace, volume.Name)
	}

	volID, ok := flx.Options["VolumeID"]
	if !ok {
		return fmt.Errorf("Volume %s/%s does not contain VolumeID attribute", volume.Namespace, volume.Name)
	}

	glog.V(5).Infof("Deleting Digital Ocean volume %q ", volID)
	return p.digitalocean.DeleteVolume(volID)
}
