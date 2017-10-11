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
	"reflect"
	"strconv"
	"testing"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/kubernetes-incubator/external-storage/lib/gidallocator"
	"github.com/kubernetes-incubator/external-storage/profitbricks/flex-volume/pkg/cloud"
	"github.com/profitbricks/profitbricks-sdk-go"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	defaultVolID       = "abcd-1234"
	defaultGid         = 4321
	defaultPVNamespace = "my-namespace"
	defaultPVName      = "my-name"
)

type fakeManager struct {
	createVolumeFn func(name, description string, sizeGB int) (*profitbricks.Volume, error)
	deleteVolumeFn func(volumeID string) error
}

func (f *fakeManager) CreateVolume(name, description string, sizeGB int) (*profitbricks.Volume, error) {
	if f.createVolumeFn != nil {
		return f.createVolumeFn(name, description, sizeGB)
	}
	// default fake implementation
	return &profitbricks.Volume{
		ID:            defaultVolID,
		Name:          name,
		Description:   description,
		SizeGigaBytes: int64(sizeGB),
	}, nil
}

func (f *fakeManager) DeleteVolume(volumeID string) error {
	if f.deleteVolumeFn != nil {
		return f.deleteVolumeFn(volumeID)
	}
	// default fake implementation
	return nil
}

type fakeAllocator struct {
	allocateNextFn func(options controller.VolumeOptions) (int, error)
	releaseFn      func(volume *v1.PersistentVolume) error
}

func (f *fakeAllocator) AllocateNext(options controller.VolumeOptions) (int, error) {
	if f.allocateNextFn != nil {
		return f.AllocateNext(options)
	}
	// default fake implementation
	return defaultGid, nil
}

func (f *fakeAllocator) Release(volume *v1.PersistentVolume) error {
	if f.releaseFn != nil {
		return f.releaseFn(volume)
	}
	// default fake implementation
	return nil
}

func TestNewProfitbricksProvisioner(t *testing.T) {
	testcases := []struct {
		name    string
		client  kubernetes.Interface
		manager cloud.VolumeManager
		// expected
		provisioner *profitbricksProvisioner
		err         error
	}{
		{
			"kubernetes client nil",
			nil,
			&fakeManager{},
			nil,
			errors.New("Provisioner needs the kubernetes client"),
		},
		{
			"do manager nil",
			&fake.Clientset{},
			nil,
			nil,
			errors.New("Provisioner needs the Profitbricks client"),
		},
		{
			"new provisioner",
			&fake.Clientset{},
			&fakeManager{},
			&profitbricksProvisioner{},
			nil,
		},
	}
	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			p, err := NewprofitbricksProvisioner(test.client, test.manager, "driver-name")
			if (p == nil || test.provisioner == nil) &&
				(p != nil || test.provisioner != nil) {
				t.Error("unexpected provisioner")
				t.Logf("expected: %v", test.provisioner)
				t.Logf("actual: %v", p)
			}

			if !reflect.DeepEqual(err, test.err) {
				t.Error("unexpected error")
				t.Logf("expected: %v", test.err)
				t.Logf("actual: %v", err)
			}

		})
	}
}

func TestProvision(t *testing.T) {
	testcases := []struct {
		name      string
		manager   cloud.VolumeManager
		allocator allocatorInterface
		// expected
		flexVolume *v1.FlexVolumeSource
		err        error
	}{
		{
			"provision volume",
			&fakeManager{},
			&fakeAllocator{},
			&v1.FlexVolumeSource{
				Driver: "flex-profitbricks-driver",
				Options: map[string]string{
					"VolumeID":   defaultVolID,
					"VolumeName": "test-volume",
				},
				ReadOnly: false,
			},
			nil,
		},
	}
	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			provisioner := profitbricksProvisioner{
				flexDriver: test.flexVolume.Driver,
				manager:    test.manager,
				allocator:  test.allocator,
			}
			opts := createVolumeOptions(10, test.flexVolume.Options["VolumeName"])
			pv, err := provisioner.Provision(opts)

			if pv.Annotations[gidallocator.VolumeGidAnnotationKey] != strconv.Itoa(defaultGid) {
				t.Error("unexpected gid")
				t.Logf("expected: %v", defaultGid)
				t.Logf("actual: %v", pv.Annotations[gidallocator.VolumeGidAnnotationKey])
			}
			if !reflect.DeepEqual(pv.Spec.PersistentVolumeSource.FlexVolume, test.flexVolume) {
				t.Error("unexpected flexVolume")
				t.Logf("expected: %v", test.flexVolume)
				t.Logf("actual: %v", pv.Spec.PersistentVolumeSource.FlexVolume)
			}
			if !reflect.DeepEqual(err, test.err) {
				t.Error("unexpected error")
				t.Logf("expected: %v", test.err)
				t.Logf("actual: %v", err)
			}

		})
	}
}

func TestDelete(t *testing.T) {
	testcases := []struct {
		name      string
		manager   cloud.VolumeManager
		allocator allocatorInterface
		pv        *v1.PersistentVolume
		// expected
		err error
	}{
		{
			"provision volume",
			&fakeManager{},
			&fakeAllocator{},
			createPersistentVolume(defaultPVNamespace, defaultPVName, defaultVolID),
			nil,
		},
	}
	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			provisioner := profitbricksProvisioner{
				manager:   test.manager,
				allocator: test.allocator,
			}
			err := provisioner.Delete(test.pv)

			if !reflect.DeepEqual(err, test.err) {
				t.Error("unexpected error")
				t.Logf("expected: %v", test.err)
				t.Logf("actual: %v", err)
			}

		})
	}
}

func createVolumeOptions(capacity int64, name string) controller.VolumeOptions {
	opts := controller.VolumeOptions{
		PVC: &v1.PersistentVolumeClaim{
			Spec: v1.PersistentVolumeClaimSpec{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{},
				},
			},
		},
		PVName: name,
	}

	c := resource.NewQuantity(capacity, "Gi")
	opts.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)] = *c

	return opts
}

func createPersistentVolume(namespace, name, volID string) *v1.PersistentVolume {
	v := &v1.PersistentVolume{
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FlexVolume: &v1.FlexVolumeSource{
					Options: map[string]string{
						"VolumeID": volID,
					},
				},
			},
		},
	}
	v.Name = name
	v.Namespace = namespace
	return v
}

// func createPersistentVolume(options controller.VolumeOptions) *v1.PersistentVolume {
// 	return &v1.PersistentVolume{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: options.PVName,
// 			Annotations: map[string]string{
// 				gidallocator.VolumeGidAnnotationKey: strconv.FormatInt(int64(gid), 10),
// 			},
// 		},
// 		Spec: v1.PersistentVolumeSpec{
// 			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
// 			AccessModes:                   options.PVC.Spec.AccessModes,
// 			Capacity: v1.ResourceList{
// 				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
// 			},
// 			PersistentVolumeSource: v1.PersistentVolumeSource{
// 				FlexVolume: &v1.FlexVolumeSource{
// 					Driver: p.flexDriver,
// 					Options: map[string]string{
// 						"VolumeID":   vol.ID,
// 						"VolumeName": vol.Name,
// 					},
// 					ReadOnly: false,
// 				},
// 			},
// 		},
// 	}
// }
