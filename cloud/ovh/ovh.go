package ovh

import (
	cloud "github.com/SebastienDorgan/gpac/cloud"
)


type OVH struct {

}

func (OVH) CreateKeyPair(name string) (cloud.KeyPair, error) {
	panic("implement me")
}

func (OVH) GetKeyPair(id string) (cloud.KeyPair, error) {
	panic("implement me")
}

func (OVH) ListKeyPairs() ([]cloud.KeyPair, error) {
	panic("implement me")
}

func (OVH) DeleteKeyPair(id string) error {
	panic("implement me")
}

func (OVH) CreateVM(name string, pair cloud.KeyPair, size cloud.VMSize, public_ip bool) (cloud.VM, error) {
	panic("implement me")
}

func (OVH) GetVM(id string) (cloud.VM, error) {
	panic("implement me")
}

func (OVH) ListVMs() ([]cloud.VM, error) {
	panic("implement me")
}

func (OVH) DeleteVM(id string) error {
	panic("implement me")
}

func (OVH) CreateVolume(name string, size float32) (cloud.Volume, error) {
	panic("implement me")
}

func (OVH) GetVolume(id string) (cloud.Volume, error) {
	panic("implement me")
}

func (OVH) ListVolumes() ([]cloud.Volume, error) {
	panic("implement me")
}

func (OVH) DeleteVolume(id string) error {
	panic("implement me")
}

func (OVH) CreateVolumeAttachment(name string, volume cloud.Volume, vm cloud.VM) (cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (OVH) GetVolumeAttachment(id string) (cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (OVH) ListVolumeAttachments() ([]cloud.VolumeAttachment, error) {
	panic("implement me")
}

func (OVH) DeleteVolumeAttachment(id string) error {
	panic("implement me")
}

