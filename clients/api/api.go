package api

import (
	"github.com/SebastienDorgan/gpac/clients/api/IPVersion"
	"github.com/SebastienDorgan/gpac/clients/api/VMState"
	"github.com/SebastienDorgan/gpac/clients/api/VolumeState"
)

//TimeoutError defines a Timeout erroe
type TimeoutError struct {
	Message string
}

func (e *TimeoutError) Error() string {
	return e.Message
}

//KeyPair represents a SSH key pair
type KeyPair struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	PublicKey  string `json:"public_key,omitempty"`
}

//VMSize represent Sizing elements of a VM
type VMSize struct {
	Cores    int     `json:"cores,omitempty"`
	RAMSize  float32 `json:"ram_size,omitempty"`
	DiskSize int     `json:"disk_size,omitempty"`
}

//VMTemplate represents a VM template
type VMTemplate struct {
	VMSize `json:"vm_size,omitempty"`
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
}

//SizingRequirements represents VM sizing requirements to fulfil
type SizingRequirements struct {
	MinCores    int     `json:"min_cores,omitempty"`
	MinRAMSize  float32 `json:"min_ram_size,omitempty"`
	MinDiskSize int     `json:"min_disk_size,omitempty"`
}

//VM represents a virtual machine properties
type VM struct {
	ID           string       `json:"id,omitempty"`
	Name         string       `json:"name,omitempty"`
	PrivateIPsV4 []string     `json:"private_i_ps_v_4,omitempty"`
	PrivateIPsV6 []string     `json:"private_i_ps_v_6,omitempty"`
	AccessIPv4   string       `json:"access_i_pv_4,omitempty"`
	AccessIPv6   string       `json:"access_i_pv_6,omitempty"`
	Size         VMSize       `json:"size,omitempty"`
	State        VMState.Enum `json:"state,omitempty"`
}

//VMRequest represents requirements to create virtual machine properties
type VMRequest struct {
	Name string `json:"name,omitempty"`
	//KeyPairID ID of the key pair use to secure SSH connections with the VM
	KeyPairID string `json:"key_pair_id,omitempty"`
	//NetworksIDs list of the network IDs the VM must be connected
	NetworkIDs []string `json:"network_i_ds,omitempty"`
	//PublicIP a flg telling if the VM must have a public IP is
	PublicIP bool `json:"public_ip,omitempty"`
	//TemplateID the UUID of the template used to size the VM (see SelectTemplates)
	TemplateID string `json:"template_id,omitempty"`
	//ImageID  is the UUID of the image that contains the server's OS and initial state.
	ImageID string `json:"image_id,omitempty"`
}

//Volume represents an block volume
type Volume struct {
	ID    string           `json:"id,omitempty"`
	Name  string           `json:"name,omitempty"`
	Size  int              `json:"size,omitempty"`
	Type  string           `json:"type,omitempty"`
	State VolumeState.Enum `json:"state,omitempty"`
}

//VolumeRequest represents a volume request
type VolumeRequest struct {
	Name string `json:"name,omitempty"`
	Size int    `json:"size,omitempty"`
	Type string `json:"type,omitempty"`
}

//VolumeType represents a volume type
type VolumeType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

//VolumeAttachment represents an volume attachment
type VolumeAttachment struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	VolumeID string `json:"volume,omitempty"`
	ServerID string `json:"vm,omitempty"`
	Device   string `json:"device,omitempty"`
}

//VolumeAttachmentRequest represents an volume attachment request
type VolumeAttachmentRequest struct {
	Name     string `json:"name,omitempty"`
	VolumeID string `json:"volume,omitempty"`
	ServerID string `json:"vm,omitempty"`
}

//Image representes an OS image
type Image struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

//Network representes a virtual network
type Network struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// IDs of the Subnets associated with this network.
	Subnets []string `json:"subnets,omitempty"`
}

//Subnet represents a sub network where Mask is defined in CIDR notation
//like "192.0.2.0/24" or "2001:db8::/32", as defined in RFC 4632 and RFC 4291.
type Subnet struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	//IPVersion is IPv4 or IPv6 (see IPVersion)
	IPVersion IPVersion.Enum `json:"ip_version,omitempty"`
	//Mask mask in CIDR notation
	Mask string `json:"mask,omitempty"`
	//NetworkID id of the parent network
	NetworkID string `json:"network_id,omitempty"`
}

//SubnetRequets represents sub network requirements to create a subnet where Mask is defined in CIDR notation
//like "192.0.2.0/24" or "2001:db8::/32", as defined in RFC 4632 and RFC 4291.
type SubnetRequets struct {
	Name string `json:"name,omitempty"`
	//IPVersion must be IPv4 or IPv6 (see IPVersion)
	IPVersion IPVersion.Enum `json:"ip_version,omitempty"`
	//Mask mask in CIDR notation
	Mask string `json:"mask,omitempty"`
	//NetworkID id of the parent network
	NetworkID string `json:"network_id,omitempty"`
}

//ClientAPI is an API defining an IaaS driver
type ClientAPI interface {
	//SetDefaultUser set the default user
	SetDefaultUser(user string)
	//GetDefaultUser returns server default user
	GetDefaultUser() string
	//ListImages lists available OS images
	ListImages() ([]Image, error)
	//GetImage returns the Image referenced by id
	GetImage(id string) (*Image, error)
	//GetTemplate returns the Template referenced by id
	GetTemplate(id string) (*VMTemplate, error)
	//ListTemplates lists available VM templates
	//VM templates are sorted using Dominant Resource Fairness Algorithm
	ListTemplates() ([]VMTemplate, error)

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
	CreateVolume(request VolumeRequest) (*Volume, error)
	//GetVolume returns the volume identified by id
	GetVolume(id string) (*Volume, error)
	//ListVolumes list available volumes
	ListVolumes() ([]Volume, error)
	//DeleteVolume deletes the volume identified by id
	DeleteVolume(id string) error
	//ListVolumeTypes list volume types available
	ListVolumeTypes() ([]VolumeType, error)

	//CreateVolumeAttachment attaches a volume to a VM
	//- name the name of the volume attachment
	//- volume the volume to attach
	//- vm the VM on which the volume is attached
	CreateVolumeAttachment(request VolumeAttachmentRequest) (*VolumeAttachment, error)
	//GetVolumeAttachment returns the volume attachment identified by id
	GetVolumeAttachment(serverID, id string) (*VolumeAttachment, error)
	//ListVolumeAttachments lists available volume attachment
	ListVolumeAttachments(serverID string) ([]VolumeAttachment, error)
	//DeleteVolumeAttachment deletes the volume attachment identifed by id
	DeleteVolumeAttachment(serverID, id string) error
}
