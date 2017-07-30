package cloud

type GCE struct {
	
}

func (GCE) CreateKeyPair(name string) (KeyPair, error) {
	panic("implement me")
}

func (GCE) GetKeyPair(id string) (KeyPair, error) {
	panic("implement me")
}

func (GCE) ListKeyPairs() ([]KeyPair, error) {
	panic("implement me")
}

func (GCE) DeleteKeyPair(id string) error {
	panic("implement me")
}

func (GCE) CreateVM(name string, pair KeyPair, size VMSize, public_ip bool) (VM, error) {
	panic("implement me")
}

func (GCE) GetVM(id string) (VM, error) {
	panic("implement me")
}

func (GCE) ListVMs() ([]VM, error) {
	panic("implement me")
}

func (GCE) DeleteVM(id string) error {
	panic("implement me")
}

func (GCE) CreateVolume(name string, size float32) (Volume, error) {
	panic("implement me")
}

func (GCE) GetVolume(id string) (Volume, error) {
	panic("implement me")
}

func (GCE) ListVolumes() ([]Volume, error) {
	panic("implement me")
}

func (GCE) DeleteVolume(id string) error {
	panic("implement me")
}

func (GCE) CreateVolumeAttachment(name string, volume Volume, vm VM) (VolumeAttachment, error) {
	panic("implement me")
}

func (GCE) GetVolumeAttachment(id string) (VolumeAttachment, error) {
	panic("implement me")
}

func (GCE) ListVolumeAttachments() ([]VolumeAttachment, error) {
	panic("implement me")
}

func (GCE) DeleteVolumeAttachment(id string) error {
	panic("implement me")
}

