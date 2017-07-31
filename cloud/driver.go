package cloud

type KeyPair struct {
	Id              string
	Name            string
	Private_content []byte
	Public_content  []byte
}

type VMSize struct {
	Id    string
	Cores int
	Ram   float32
	Disk  float32
}

type VMState int
/*
const (
	STOPPED  VMState = iota
	STARTING
	STARTED
	STOPPING
	IN_ERROR
)
*/

type VM struct {
	Id        string
	Name      string
	PrivateIp string
	PublicIp  string
	Size      VMSize
	Status    VMState
}
type Volume struct {
	Id   string
	Name string
	Size float32
	Available bool
}

type VolumeAttachment struct {
	Id     string
	Name   string
	Volume Volume
	VM     VM
}

type driver interface {
	CreateKeyPair(name string) (KeyPair, error)
	GetKeyPair(id string) (KeyPair, error)
	ListKeyPairs() ([]KeyPair, error)
	DeleteKeyPair(id string) error

	CreateVM(name string, pair KeyPair, size VMSize, public_ip bool) (VM, error)
	GetVM(id string) (VM, error)
	ListVMs() ([]VM, error)
	DeleteVM(id string) error
	StopVM(id string) error
	StartVM(id string) error

	CreateVolume(name string, size float32) (Volume, error)
	GetVolume(id string) (Volume, error)
	ListVolumes() ([]Volume, error)
	DeleteVolume(id string) error

	CreateVolumeAttachment(name string, volume Volume, vm VM) (VolumeAttachment, error)
	GetVolumeAttachment(id string) (VolumeAttachment, error)
	ListVolumeAttachments() ([]VolumeAttachment, error)
	DeleteVolumeAttachment(id string) error
}
