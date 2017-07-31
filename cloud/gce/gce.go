package gce

import (
	cloud "github.com/SebastienDorgan/gpac/cloud"
)


type GCE struct {

}

func (GCE) CreateKeyPair(name string) (cloud.KeyPair, error) {
	panic("implement me")
}

func (GCE) GetKeyPair(id string) (cloud.KeyPair, error) {
	panic("implement me")
}

func (GCE) ListKeyPairs() ([]cloud.KeyPair, error) {
	panic("implement me")
}

func (GCE) DeleteKeyPair(id string) error {
	panic("implement me")
}

func (GCE) CreateVM(name string, pair cloud.KeyPair, size cloud.VMSize, public_ip bool) (cloud.VM, error) {
	panic("implement me")
}

func (GCE) GetVM(id string) (cloud.VM, error) {
	panic("implement me")
}

func (GCE) ListVMs() ([]cloud.VM, error) {
	panic("implement me")
}

func (GCE) DeleteVM(id string) error {
	panic("implement me")
}

func (GCE) CreateVolume(name string, size float32) (cloud.Volume, error) {
	panic("implement me")
}

func (GCE) GetVolume(id string) (cloud.Volume, error) {
	panic("implement me")
}

func (GCE) ListVolumes() ([]cloud.Volume, error) {
	panic("implement me")
}

func (GCE) DeleteVolume(id string) error {
	panic("implement me")
}

func (GCE) CreateVolumeAttachment(name string, volume cloud.Volume, vm cloud.VM) (cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (GCE) GetVolumeAttachment(id string) (cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (GCE) ListVolumeAttachments() ([]cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (GCE) DeleteVolumeAttachment(id string) error {
	panic("implement me")
}

