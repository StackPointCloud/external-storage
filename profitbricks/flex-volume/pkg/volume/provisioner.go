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
	"errors"
	"fmt"
	"strconv"

	"github.com/external-storage/profitbricks/flex-volume/pkg/cloud"
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/kubernetes-incubator/external-storage/lib/gidallocator"
	"github.com/profitbricks/profitbricks-sdk-go"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/volume"
)

const (
	volumeDescription   = "Kubernetes dynamic provisioned"
	volumeNameMaxLenght = 64
	licenceType         = "LINUX"
	volumeType          = "HDD"
)

type allocatorInterface interface {
	AllocateNext(options controller.VolumeOptions) (int, error)
	Release(volume *v1.PersistentVolume) error
}

type profitbricksProvisioner struct {
	client    kubernetes.Interface
	manager   cloud.VolumeManager
	allocator allocatorInterface
	// gidallocator.Allocator
	flexDriver string
}

// NewProfitbricksProvisioner creates a Profitbricks volume provisioner
func NewProfitbricksProvisioner(client kubernetes.Interface, pb cloud.VolumeManager, flexDriver string) (controller.Provisioner, error) {

	if client == nil {
		return nil, errors.New("Provisioner needs the kubernetes client")
	}

	if pb == nil {
		return nil, errors.New("Provisioner needs the Profitbricks client")
	}

	allocator := gidallocator.New(client)
	return &profitbricksProvisioner{
		client:     client,
		manager:    pb,
		allocator:  &allocator,
		flexDriver: flexDriver,
	}, nil
}

var _ controller.Provisioner = &profitbricksProvisioner{}

// Provision creates a volume i.e. the storage asset and returns a PV object for
// the volume.
func (p *profitbricksProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
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
						"VolumeID":   vol.Id,
						"VolumeName": vol.Properties.Name,
					},
					ReadOnly: false,
				},
			},
		},
	}

	return pv, nil
}

// createVolume creates a volume at Profitbricks
func (p *profitbricksProvisioner) createVolume(options controller.VolumeOptions) (*profitbricks.Volume, error) {

	capacity := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	sizeGB := int(volume.RoundUpSize(capacity.Value(), 1024*1024*1024))

	glog.V(5).Infof("Creating Profitbricks volume %s sized %d GB", options.PVName, sizeGB)
	vol, err := p.manager.CreateVolume(options.PVName, sizeGB, licenceType, volumeType)
	if err != nil {
		return nil, err
	}

	return vol, nil
}

// Delete removes the directory that was created by Provision backing the given
// PV and removes its export from the NFS server.
func (p *profitbricksProvisioner) Delete(volume *v1.PersistentVolume) error {
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

	glog.V(5).Infof("Deleting Profitbricks volume %q ", volID)
	return p.manager.DeleteVolume(volID)
}
