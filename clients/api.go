package drivers

//KeyPair represents a SSH key pair
type KeyPair struct {
	ID         string
	Name       string
	PrivateKey string
	PublicKey  string
}

//VMTemplate represents a VM template
type VMTemplate struct {
	ID       string
	Name     string
	Cores    int
	RAMSize  float32
	DiskSize int
}

const (
	//CoreDRFWeight is the Dominant Resource Fairness weight of a core
	CoreDRFWeight float32 = 1.0
	//RAMDRFWeight is the Dominant Resource Fairness weight of 1 GB of RAM
	RAMDRFWeight float32 = 1.0 / 8.0
	//DiskDRFWeight is the Dominant Resource Fairness weight of 1 GB of Disk
	DiskDRFWeight float32 = 1.0 / 16.0
)

//RankDRF computes the Dominant Resource Fairness Rank of a VM template
func (t *VMTemplate) RankDRF() float32 {
	fc := float32(t.Cores)
	fr := t.RAMSize
	fd := float32(t.DiskSize)
	return fc*CoreDRFWeight + fr*RAMDRFWeight + fd*DiskDRFWeight
}

// ByRankDRF implements sort.Interface for []VMTemplate based on
// the Dominant Resource Fairness
type ByRankDRF []VMTemplate

func (a ByRankDRF) Len() int           { return len(a) }
func (a ByRankDRF) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRankDRF) Less(i, j int) bool { return a[i].RankDRF() < a[j].RankDRF() }

//SizingRequirements represents VM sizing requirements to fulfil
type SizingRequirements struct {
	MinCores    int
	MinRAMSize  float32
	MinDiskSize int
}

//VMState represents the state of a VM
type VMState int

const (
	/*STOPPED VM is stopped*/
	STOPPED VMState = iota
	/*STARTING VM is starting*/
	STARTING
	/*STARTED VM is started*/
	STARTED
	/*STOPPING VM is stopping*/
	STOPPING
	/*ERROR VM is in error state*/
	ERROR
)

var vmStates = [...]string{
	"STOPPED",
	"STARTING",
	"STARTED",
	"STOPPING",
	"IN_ERROR",
}

func (state VMState) String() string {
	return vmStates[state]
}

//IPVersion is an enum defining IP versions
type IPVersion int

const (
	//IPv4 is IP v4 version
	IPv4 IPVersion = 4
	//IPv6 is IP v6 version
	IPv6 IPVersion = 6
)

var ipVersion = [...]string{
	"V4",
	"V6",
}

func (version IPVersion) String() string {
	return ipVersion[version]
}

//VM represents a virtual machine properties
type VM struct {
	ID           string
	Name         string
	PrivateIPsV4 []string
	PrivateIPsV6 []string
	AccessIPv4   string
	AccessIPv6   string
	Size         VMTemplate
	State        VMState
}

//VMRequest represents requirements to create virtual machine properties
type VMRequest struct {
	Name string
	//KeyPairID ID of the key pair use to secure SSH connections with the VM
	KeyPairID string
	//NetworksIDs list of the network IDs the VM must be connected
	NetworkIDs []string
	//PublicIP a flg telling if the VM must have a public IP is
	PublicIP bool
	//TemplateID the UUID of the template used to size the VM (see SelectTemplates)
	TemplateID string
	//ImageID  is the UUID of the image that contains the server's OS and initial state.
	ImageID string
}

//Volume represents an block volume
type Volume struct {
	ID        string
	Name      string
	Size      float32
	Type      string
	Available bool
}

//VolumeAttachment represents an volume attachment
type VolumeAttachment struct {
	ID     string
	Name   string
	Volume Volume
	VM     VM
}

//Image representes an OS image
type Image struct {
	ID   string
	Name string
}

//Network representes a virtual network
type Network struct {
	ID   string
	Name string
	// IDs of the Subnets associated with this network.
	Subnets []string
}

//Subnet represents a sub network where Mask is defined in CIDR notation
//like "192.0.2.0/24" or "2001:db8::/32", as defined in RFC 4632 and RFC 4291.
type Subnet struct {
	ID   string
	Name string
	//IPVersion is IPv4 or IPv6 (see IPVersion)
	IPVersion IPVersion
	//Mask mask in CIDR notation
	Mask string
	//NetworkID id of the parent network
	NetworkID string
}

//SubnetRequets represents sub network requirements to create a subnet where Mask is defined in CIDR notation
//like "192.0.2.0/24" or "2001:db8::/32", as defined in RFC 4632 and RFC 4291.
type SubnetRequets struct {
	Name string
	//IPVersion must be IPv4 or IPv6 (see IPVersion)
	IPVersion IPVersion
	//Mask mask in CIDR notation
	Mask string
	//NetworkID id of the parent network
	NetworkID string
}

//ClientAPI is an API defining an IaaS driver
type ClientAPI interface {
	//ListImages lists available OS images
	ListImages() ([]Image, error)
	//ListTemplates lists available VM templates
	//VM templates are sorted using Dominant Resource Fairness Algorithm
	ListTemplates() ([]VMTemplate, error)
	//SelectTemplates lists VM templates compatible with sizing requirements
	//VM templates are sorted using Dominant Resource Fairness Algorithm
	SelectTemplates(sizing SizingRequirements) ([]VMTemplate, error)

	//CreateKeyPair creates and import a key pair
	CreateKeyPair(name string) (*KeyPair, error)
	//GetKeyPair returns the key pair identified by id
	GetKeyPair(id string) (*KeyPair, error)
	//ListKeyPairs lists available key pairs
	ListKeyPairs() ([]KeyPair, error)
	//DeleteKeyPair deletes the key pair identified by id
	DeleteKeyPair(id string) error

	//CreateNetwork creates a network named name
	CreateNetwork(name string) (*Network, error)
	//GetNetwork returns the network identified by id
	GetNetwork(id string) (*Network, error)
	//ListNetworks lists available networks
	ListNetworks() ([]Network, error)
	//DeleteNetwork deletes the network identified by id
	DeleteNetwork(id string) error

	//CreateSubnet creates a sub network
	//- netID ID of the parent network
	//- name is the name of the sub network
	//- mask is a network mask defined in CIDR notation
	CreateSubnet(request SubnetRequets) (*Subnet, error)
	//GetSubnet returns the sub network identified by id
	GetSubnet(id string) (*Subnet, error)
	//ListSubnets lists available sub networks of network net
	ListSubnets(netID string) ([]Subnet, error)
	//DeleteSubnet deletes the sub network identified by id
	DeleteSubnet(id string) error

	//CreateVM creates a VM that fulfils the request
	CreateVM(request VMRequest) (*VM, error)
	//GetVM returns the VM identified by id
	GetVM(id string) (*VM, error)
	//ListVMs lists available VMs
	ListVMs() ([]VM, error)
	//DeleteVM deletes the VM identified by id
	DeleteVM(id string) error
	//StopVM stops the VM identified by id
	StopVM(id string) error
	//StartVM starts the VM identified by id
	StartVM(id string) error

	//CreateVolume creates a block volume
	//- name is the name of the volume
	//- size is the size of the volume in GB
	//- volumeType is the type of volume to create, if volumeType is empty the driver use a default type
	CreateVolume(name string, size float32, volumeType string) (Volume, error)
	//GetVolume returns the volume identified by id
	GetVolume(id string) (Volume, error)
	//ListVolumes list available volumes
	ListVolumes() ([]Volume, error)
	//DeleteVolume deletes the volume identified by id
	DeleteVolume(id string) error

	//CreateVolumeAttachment attaches a volume to a VM
	//- name the name of the volume attachment
	//- volume the volume to attach
	//- vm the VM on which the volume is attached
	CreateVolumeAttachment(name string, volume Volume, vm VM) (VolumeAttachment, error)
	//GetVolumeAttachment returns the volume attachment identified by id
	GetVolumeAttachment(id string) (VolumeAttachment, error)
	//ListVolumeAttachments lists available volume attachment
	ListVolumeAttachments() ([]VolumeAttachment, error)
	//DeleteVolumeAttachment deletes the volume attachment identifed by id
	DeleteVolumeAttachment(id string) error
}
