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

//https://github.com/profitbricks/profitbricks-sdk-go#create-a-volume
//https://github.com/profitbricks/profitbricks-sdk-go#delete-a-volume

type ProfitbricksManager struct {
	datacenter string
}

// VolumeManager is a Profitbricks volumes operations interface
type VolumeManager interface {
	CreateVolume(name, datacenter, sizeGB int, volumeType string) (*profitbricks.Volume, error)
	DeleteVolume(datacenter string, volumeID string) error
}

// // TokenSource represents and oauth2 token source
// type tokenSource struct {
// 	AccessToken string
// }
//
// // Token returns an oauth2 token
// func (t *tokenSource) Token() (*oauth2.Token, error) {
// 	token := &oauth2.Token{
// 		AccessToken: t.AccessToken,
// 	}
// 	return token, nil
// }

// NewProfitbricksManager returns a Profitbricks manager
func NewProfitbricksManager(datacenter string, user string, password string) (*ProfitbricksManager, error) {

	if user == "" || password == "" {
		return nil, errors.New("Profitbricks credentials must be informed")
	}

	manager := &ProfitbricksManager{
		datacenter: datacenter,
	}
	profitbricks.SetAuth(user, password)

	// generate client and test retrieving all datacenters
	datacenters := profitbricks.ListDatacenters()

	if datacenters.StatusCode != 200 {
		return nil, fmt.Errorf("an error occurred listing datacenters: %s", datacenters.Response)
	}

	for _, dc := range datacenters.Items {
		if dc.Properties.Name == datacenter {
			manager.datacenter = dc.Id
			return manager, nil
		}
	}
	return nil, fmt.Errorf("datacenter %s not found", datacenter)
}

// Fetch and return the UUID of a resource regardless of whether the name orUUID is passed.
// func get_resource_id(resource_list, identity):
// 	for r: resource_list; r != nil; r = r.Next(){
// 		if identity
// 	}
//     for resource in resource_list['items']:
//         if identity in (resource['properties']['name'], resource['id']):
//             return resource['id']
//     return None

// CreateVolume creates a Profitbricks volume
func (m *ProfitbricksManager) CreateVolume(name, volumeType, licenceType string, size int) (*profitbricks.Volume, error) {
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
