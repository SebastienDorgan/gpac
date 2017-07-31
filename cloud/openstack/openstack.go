package openstack

import (
	cloud "github.com/SebastienDorgan/gpac/cloud"
)

type Driver struct {
	AccessData AccessData
}

func Create(auth AuthAPI) (*Driver, error) {
	access, err := auth.Authenticate()
	if err != nil {
		return nil, err
	}
	return &Driver{AccessData: access}, nil
}

func (Driver) CreateKeyPair(name string) (cloud.KeyPair, error) {
	panic("implement me")
}

func (Driver) GetKeyPair(id string) (cloud.KeyPair, error) {
	panic("implement me")
}

func (Driver) ListKeyPairs() ([]cloud.KeyPair, error) {
	panic("implement me")
}

func (Driver) DeleteKeyPair(id string) error {
	panic("implement me")
}

func (Driver) CreateVM(name string, pair cloud.KeyPair, size cloud.VMSize, public_ip bool) (cloud.VM, error) {
	panic("implement me")
}

func (Driver) GetVM(id string) (cloud.VM, error) {
	panic("implement me")
}

func (Driver) ListVMs() ([]cloud.VM, error) {
	panic("implement me")
}

func (Driver) DeleteVM(id string) error {
	panic("implement me")
}

func (Driver) StopVM(id string) error {
	panic("implement me")
}

func (Driver) StartVM(id string) error {
	panic("implement me")
}

func (Driver) CreateVolume(name string, size float32) (cloud.Volume, error) {
	panic("implement me")
}

func (Driver) GetVolume(id string) (cloud.Volume, error) {
	panic("implement me")
}

func (Driver) ListVolumes() ([]cloud.Volume, error) {
	panic("implement me")
}

func (Driver) DeleteVolume(id string) error {
	panic("implement me")
}

func (Driver) CreateVolumeAttachment(name string, volume cloud.Volume, vm cloud.VM) (cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (Driver) GetVolumeAttachment(id string) (cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (Driver) ListVolumeAttachments() ([]cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (Driver) DeleteVolumeAttachment(id string) error {
	panic("implement me")
}
