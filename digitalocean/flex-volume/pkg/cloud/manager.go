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

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

// VolumeManager is a Digital Ocean cloud volumes operations interface
type VolumeManager interface {
	CreateVolume(name, description string, sizeGB int) (*godo.Volume, error)
	DeleteVolume(volumeID string) error
}

// DigitalOceanManager communicates with the DO API
type DigitalOceanManager struct {
	client *godo.Client
	region string
}

// TokenSource represents and oauth2 token source
type tokenSource struct {
	AccessToken string
}

// Token returns an oauth2 token
func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// NewDigitalOceanManager returns a Digitial Ocean manager
func NewDigitalOceanManager(token string) (*DigitalOceanManager, error) {

	if token == "" {
		return nil, errors.New("Digital Ocean token must be informed")
	}

	tokenSource := &tokenSource{AccessToken: token}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)
	region, err := dropletRegion()
	if err != nil {
		return nil, errors.New("failed to get region from droplet metadata")
	}

	do := &DigitalOceanManager{
		client: client,
		region: region,
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
