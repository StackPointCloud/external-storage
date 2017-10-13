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

package cloud

import (
	"errors"
	"fmt"

	"github.com/profitbricks/profitbricks-sdk-go"
)

type ProfitbricksCredentials struct {
	Datacenter string
	User       string
	Password   string
}

type ProfitbricksManager struct {
	datacenter string
}

// VolumeManager is a Profitbricks volumes operations interface
type VolumeManager interface {
	CreateVolume(name string, sizeGB int, volumeType string, licenceType string) (*profitbricks.Volume, error)
	DeleteVolume(volumeID string) error
}

// NewProfitbricksManager returns a Profitbricks manager
func NewProfitbricksManager(credentials *ProfitbricksCredentials) (*ProfitbricksManager, error) {

	if credentials.User == "" || credentials.Password == "" {
		return nil, errors.New("Profitbricks credentials must be informed")
	}

	manager := &ProfitbricksManager{
		datacenter: credentials.Datacenter,
	}
	// set auth
	profitbricks.SetAuth(credentials.User, credentials.Password)

	// generate client and test retrieving all datacenters
	datacenters := profitbricks.ListDatacenters()

	if datacenters.StatusCode != 200 {
		return nil, fmt.Errorf("an error occurred listing datacenters: %s", datacenters.Response)
	}

	for _, dc := range datacenters.Items {
		if dc.Properties.Name == credentials.Datacenter {
			manager.datacenter = dc.Id
			return manager, nil
		}
	}
	return nil, fmt.Errorf("datacenter %s not found", credentials.Datacenter)
}

// CreateVolume creates a Profitbricks volume
func (m *ProfitbricksManager) CreateVolume(name string, size int, volumeType string, licenceType string) (*profitbricks.Volume, error) {
	req := profitbricks.Volume{
		Properties: profitbricks.VolumeProperties{
			Size:        size,
			Name:        name,
			LicenceType: licenceType,
			Type:        volumeType,
		},
	}

	vol := profitbricks.CreateVolume(m.datacenter, req)

	if vol.StatusCode != 202 {
		return nil, fmt.Errorf("an error occurred creating volume: %s", vol.Id)
	}

	return &vol, nil
}

// DeleteVolume deletes a Digital Ocean volume
func (m *ProfitbricksManager) DeleteVolume(volumeID string) error {
	profitbricks.DeleteVolume(m.datacenter, volumeID)
	return nil
}
