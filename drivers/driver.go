package drivers

type CloudObject struct{
	Id string
}

type KeyPair struct {
	CloudObject
	Name            string
	Private_content []byte
	Public_content  []byte
}

type VMTemplate struct {
	CloudObject
	Name  string
	Cores int
	RamSize   float32
	DiskSize  float32
	CPUFrequency float32
	Price float32
}

type VMState int

const (
	STOPPED  VMState = iota
	STARTING
	STARTED
	STOPPING
	IN_ERROR
)
var vmStates = [...]string {
	"STOPPED",
	"STARTING",
	"STARTED",
	"STOPPING",
	"IN_ERROR",
}

func (state VMState) String() string {
	return vmStates[state - 1]
}

type VM struct {
	CloudObject
	Name      string
	PrivateIp string
	PublicIp  string
	Size      VMTemplate
	State     VMState
}
type Volume struct {
	CloudObject
	Name string
	Size float32
	Category string
	Available bool
}

type VolumeAttachment struct {
	CloudObject
	Name   string
	Volume Volume
	VM     VM
}

type Image struct {
	CloudObject
	Name string
}

type Network struct {
	CloudObject
	Name string
	CIDR string
}

type Driver interface {
	ListImages() ([]Image, error)
	ListVMTemplates() ([]VMTemplate, error)

	CreateKeyPair(name string) (KeyPair, error)
	GetKeyPair(id string) (KeyPair, error)
	ListKeyPairs() ([]KeyPair, error)
	DeleteKeyPair(id string) error

	CreateNetwork(name string) (Network, error)
	GetNetwork(id string) (Network, error)
	ListNetwork() ([]Network, error)
	DeleteNetwork(id string) error

	CreateVM(name string, size VMTemplate, image Image, keyPair KeyPair, network Network, public_ip bool) (VM, error)
	GetVM(id string) (VM, error)
	ListVMs() ([]VM, error)
	DeleteVM(id string) error
	StopVM(id string) error
	StartVM(id string) error

	CreateVolume(name string, size float32, category string) (Volume, error)
	GetVolume(id string) (Volume, error)
	ListVolumes() ([]Volume, error)
	DeleteVolume(id string) error

	CreateVolumeAttachment(name string, volume Volume, vm VM) (VolumeAttachment, error)
	GetVolumeAttachment(id string) (VolumeAttachment, error)
	ListVolumeAttachments() ([]VolumeAttachment, error)
	DeleteVolumeAttachment(id string) error
}
