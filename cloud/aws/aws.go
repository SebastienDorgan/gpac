package aws

import (
	cloud "github.com/SebastienDorgan/gpac/cloud"
)


type AWS struct {
	
}

func (AWS) CreateKeyPair(name string) (cloud.KeyPair, error) {
	panic("implement me")
}

func (AWS) GetKeyPair(id string) (cloud.KeyPair, error) {
	panic("implement me")
}

func (AWS) ListKeyPairs() ([]cloud.KeyPair, error) {
	panic("implement me")
}

func (AWS) DeleteKeyPair(id string) error {
	panic("implement me")
}

func (AWS) CreateVM(name string, pair cloud.KeyPair, size cloud.VMSize, public_ip bool) (cloud.VM, error) {
	panic("implement me")
}

func (AWS) GetVM(id string) (cloud.VM, error) {
	panic("implement me")
}

func (AWS) ListVMs() ([]cloud.VM, error) {
	panic("implement me")
}

func (AWS) DeleteVM(id string) error {
	panic("implement me")
}

func (AWS) CreateVolume(name string, size float32) (cloud.Volume, error) {
	panic("implement me")
}

func (AWS) GetVolume(id string) (cloud.Volume, error) {
	panic("implement me")
}

func (AWS) ListVolumes() ([]cloud.Volume, error) {
	panic("implement me")
}

func (AWS) DeleteVolume(id string) error {
	panic("implement me")
}

func (AWS) CreateVolumeAttachment(name string, volume cloud.Volume, vm cloud.VM) (cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (AWS) GetVolumeAttachment(id string) (cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (AWS) ListVolumeAttachments() ([]cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (AWS) DeleteVolumeAttachment(id string) error {
	panic("implement me")
}

