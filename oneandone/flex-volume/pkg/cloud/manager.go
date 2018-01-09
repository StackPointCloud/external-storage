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
	"regexp"

	"github.com/1and1/oneandone-cloudserver-sdk-go"
)

// OneandoneCredentials contains 1&1 credentials
type OneandoneCredentials struct {
	Token        string
	DatacenterID string
}

// OneandoneManager is a 1&1 client
type OneandoneManager struct {
	api *oneandone.API
}

// BlockStorageManager is a 1&1 block storage operations interface
type BlockStorageManager interface {
	CreateBlockStorage(name string, sizeGB int, description string, datacenterID string) (*oneandone.BlockStorage, error)
	DeleteBlockStorage(storageID string) error
}

// NewOneandoneManager returns a 1&1 manager
func NewOneandoneManager(credentials *OneandoneCredentials) (*OneandoneManager, error) {

	if credentials.Token == "" {
		return nil, errors.New("1&1 token not provided")
	}

	manager := &OneandoneManager{}
	// set auth
	manager.api = oneandone.New(credentials.Token, oneandone.BaseUrl)

	// generate client and test retrieving all datacenters
	pong, err := manager.api.PingAuth()

	if err != nil {
		return nil, fmt.Errorf("Authorization check failed. Error: %s", err.Error())
	}

	if len(pong) == 0 && pong[0] != "PONG" {
		return nil, fmt.Errorf("Invalid authorization response")
	}
	return manager, nil
}

// CreateBlockStorage creates a 1&1 block storage
func (m *OneandoneManager) CreateBlockStorage(name string, size int, description string, datacenterID string) (*oneandone.BlockStorage, error) {
	uuid := ""
	var err error

	if !m.isValidUUID(datacenterID) {
		uuid, err = m.getDatacenterID(datacenterID)
		if err != nil {
			return nil, err
		}
	} else {
		uuid = datacenterID
	}
	_, storage, err := m.api.CreateBlockStorage(&oneandone.BlockStorageRequest{
		Name: name,

		Description:  description,
		Size:         &size,
		DatacenterId: uuid,
	})

	if err != nil {
		return nil, fmt.Errorf("an error occurred creating block storag: %s", err.Error())
	}

	return storage, nil
}

// DeleteBlockStorage deletes a 1&1 block storage
func (m *OneandoneManager) DeleteBlockStorage(storageID string) error {
	_, err := m.api.DeleteBlockStorage(storageID)
	return err
}

func (m *OneandoneManager) isValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

func (m *OneandoneManager) getDatacenterID(datacenterName string) (string, error) {
	dcs, err := m.api.ListDatacenters()
	if err != nil {
		return "", fmt.Errorf("error occured while fetching datacenters %s", err.Error())
	}
	for _, d := range dcs {
		if d.CountryCode == datacenterName {
			return d.Id, nil
		}
	}

	return "", fmt.Errorf(fmt.Sprintf("error fetching datacenter %q. Error message: ", datacenterName), err)
}
