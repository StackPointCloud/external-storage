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
	"context"
	"errors"

	"github.com/profitbricks/profitbricks-sdk-go"
)

//https://github.com/profitbricks/profitbricks-sdk-go#create-a-volume
//https://github.com/profitbricks/profitbricks-sdk-go#delete-a-volume

// VolumeManager is a Digital Ocean cloud volumes operations interface
type VolumeManager interface {
	CreateVolume(name, datacenter, description string, sizeGB int) (*godo.Volume, error)
	DeleteVolume(volumeID string, datacenter) error
}

// DigitalOceanManager communicates with the DO API
type DigitalOceanManager struct {
	client *godo.Client
	region string
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
		return nil, errors.New("Digital Ocean credentials must be informed")
	}

	client := profitbricks.SetAuth(user, password)

	pb := &ProfitbricksManager{
		client: client,
		datacenter: datacenter,
	}

	// generate client and test retrieving account info
	_, err = do.GetAccount()
	if err != nil {
		return nil, err
	}

	return do, nil
}

// GetAccount returns the token related account
func (m *DigitalOceanManager) GetAccount() (*godo.Account, error) {
	account, _, err := m.client.Account.Get(context.TODO())
	if err != nil {
		return nil, err
	}
	return account, nil
}

// CreateVolume creates a Digital Ocean volume
func (m *DigitalOceanManager) CreateVolume(name, description string, sizeGB int) (*godo.Volume, error) {
	req := &godo.VolumeCreateRequest{
		Region:        m.region,
		Name:          name,
		Description:   description,
		SizeGigaBytes: int64(sizeGB),
	}

	vol, _, err := m.client.Storage.CreateVolume(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	return vol, nil
}

// DeleteVolume deletes a Digital Ocean volume
func (m *DigitalOceanManager) DeleteVolume(volumeID string) error {
	_, err := m.client.Storage.DeleteVolume(context.TODO(), volumeID)
	return err
}
