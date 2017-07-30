package cloud

type Ovh struct {
	
}

func (Ovh) CreateKeyPair(name string) (KeyPair, error) {
	panic("implement me")
}

func (Ovh) GetKeyPair(id string) (KeyPair, error) {
	panic("implement me")
}

func (Ovh) ListKeyPairs() ([]KeyPair, error) {
	panic("implement me")
}

func (Ovh) DeleteKeyPair(id string) error {
	panic("implement me")
}

func (Ovh) CreateVM(name string, pair KeyPair, size VMSize, public_ip bool) (VM, error) {
	panic("implement me")
}

func (Ovh) GetVM(id string) (VM, error) {
	panic("implement me")
}

func (Ovh) ListVMs() ([]VM, error) {
	panic("implement me")
}

func (Ovh) DeleteVM(id string) error {
	panic("implement me")
}

func (Ovh) CreateVolume(name string, size float32) (Volume, error) {
	panic("implement me")
}

func (Ovh) GetVolume(id string) (Volume, error) {
	panic("implement me")
}

func (Ovh) ListVolumes() ([]Volume, error) {
	panic("implement me")
}

func (Ovh) DeleteVolume(id string) error {
	panic("implement me")
}

func (Ovh) CreateVolumeAttachment(name string, volume Volume, vm VM) (VolumeAttachment, error) {
	panic("implement me")
}

func (Ovh) GetVolumeAttachment(id string) (VolumeAttachment, error) {
	panic("implement me")
}

func (Ovh) ListVolumeAttachments() ([]VolumeAttachment, error) {
	panic("implement me")
}

func (Ovh) DeleteVolumeAttachment(id string) error {
	panic("implement me")
}
