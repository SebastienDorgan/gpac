package cloud

type AWS struct {
	
}

func (AWS) CreateKeyPair(name string) (KeyPair, error) {
	panic("implement me")
}

func (AWS) GetKeyPair(id string) (KeyPair, error) {
	panic("implement me")
}

func (AWS) ListKeyPairs() ([]KeyPair, error) {
	panic("implement me")
}

func (AWS) DeleteKeyPair(id string) error {
	panic("implement me")
}

func (AWS) CreateVM(name string, pair KeyPair, size VMSize, public_ip bool) (VM, error) {
	panic("implement me")
}

func (AWS) GetVM(id string) (VM, error) {
	panic("implement me")
}

func (AWS) ListVMs() ([]VM, error) {
	panic("implement me")
}

func (AWS) DeleteVM(id string) error {
	panic("implement me")
}

func (AWS) CreateVolume(name string, size float32) (Volume, error) {
	panic("implement me")
}

func (AWS) GetVolume(id string) (Volume, error) {
	panic("implement me")
}

func (AWS) ListVolumes() ([]Volume, error) {
	panic("implement me")
}

func (AWS) DeleteVolume(id string) error {
	panic("implement me")
}

func (AWS) CreateVolumeAttachment(name string, volume Volume, vm VM) (VolumeAttachment, error) {
	panic("implement me")
}

func (AWS) GetVolumeAttachment(id string) (VolumeAttachment, error) {
	panic("implement me")
}

func (AWS) ListVolumeAttachments() ([]VolumeAttachment, error) {
	panic("implement me")
}

func (AWS) DeleteVolumeAttachment(id string) error {
	panic("implement me")
}

