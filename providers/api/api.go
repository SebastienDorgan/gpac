package api

import (
	"fmt"
	"io"
	"time"

	"github.com/SebastienDorgan/gpac/system"

	"github.com/SebastienDorgan/gpac/providers/api/IPVersion"
	"github.com/SebastienDorgan/gpac/providers/api/VMState"
	"github.com/SebastienDorgan/gpac/providers/api/VolumeSpeed"
	"github.com/SebastienDorgan/gpac/providers/api/VolumeState"
)

//DefaultUser Default VM user
const DefaultUser = "gpac"

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
	PrivateIPsV4 []string     `json:"private_ips_v4,omitempty"`
	PrivateIPsV6 []string     `json:"private_ips_v6,omitempty"`
	AccessIPv4   string       `json:"access_ip_v4,omitempty"`
	AccessIPv6   string       `json:"access_ip_v6,omitempty"`
	Size         VMSize       `json:"size,omitempty"`
	State        VMState.Enum `json:"state,omitempty"`
	PrivateKey   string       `json:"private_key,omitempty"`
	GatewayID    string       `json:"gateway_id,omitempty"`
}

//GetAccessIP compues access IP of the VM
func (vm *VM) GetAccessIP() string {
	ip := vm.AccessIPv4
	if ip == "" {
		ip = vm.AccessIPv6
	}
	if ip == "" {
		if len(vm.PrivateIPsV4) > 0 {
			ip = vm.PrivateIPsV4[0]
		} else {
			ip = vm.PrivateIPsV6[0]
		}
	}
	return ip
}

//VMRequest represents requirements to create virtual machine properties
type VMRequest struct {
	Name string `json:"name,omitempty"`
	//NetworksIDs list of the network IDs the VM must be connected
	NetworkIDs []string `json:"network_i_ds,omitempty"`
	//PublicIP a flg telling if the VM must have a public IP is
	PublicIP bool `json:"public_ip,omitempty"`
	//TemplateID the UUID of the template used to size the VM (see SelectTemplates)
	TemplateID string `json:"template_id,omitempty"`
	//ImageID  is the UUID of the image that contains the server's OS and initial state.
	ImageID   string   `json:"image_id,omitempty"`
	KeyPair   *KeyPair `json:"key_pair,omitempty"`
	IsGateway bool     `json:"is_gateway,omitempty"`
}

//Volume represents an block volume
type Volume struct {
	ID    string           `json:"id,omitempty"`
	Name  string           `json:"name,omitempty"`
	Size  int              `json:"size,omitempty"`
	Speed VolumeSpeed.Enum `json:"speed,omitempty"`
	State VolumeState.Enum `json:"state,omitempty"`
}

//VolumeRequest represents a volume request
type VolumeRequest struct {
	Name  string           `json:"name,omitempty"`
	Size  int              `json:"size,omitempty"`
	Speed VolumeSpeed.Enum `json:"speed,omitempty"`
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

/*
//RouterRequest represents a router request
type RouterRequest struct {
	Name string `json:"name,omitempty"`
	//NetworkID is the Network ID which the router gateway is connected to.
	NetworkID string `json:"network_id,omitempty"`
}

//Router represents a router
type Router struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	//NetworkID is the Network ID which the router gateway is connected to.
	NetworkID string `json:"network_id,omitempty"`
}
*/

//Network representes a virtual network
type Network struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	//IPVersion is IPv4 or IPv6 (see IPVersion)
	IPVersion IPVersion.Enum `json:"ip_version,omitempty"`
	//Mask mask in CIDR notation
	CIDR string `json:"mask,omitempty"`
	//Gateway network gateway
	GatewayID string
}

/*
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
*/

//NetworkRequest represents network requirements to create a subnet where Mask is defined in CIDR notation
//like "192.0.2.0/24" or "2001:db8::/32", as defined in RFC 4632 and RFC 4291.
type NetworkRequest struct {
	Name string `json:"name,omitempty"`
	//IPVersion must be IPv4 or IPv6 (see IPVersion)
	IPVersion IPVersion.Enum `json:"ip_version,omitempty"`
	//CIDR mask
	CIDR string `json:"cidr,omitempty"`
	//gwDefinition gateway of this netwok
	GWRequest VMRequest
}

//Object object to put in a container
type Object struct {
	Name          string            `json:"name,omitempty"`
	Content       io.ReadSeeker     `json:"content,omitempty"`
	DeleteAt      time.Time         `json:"delete_at,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Date          time.Time         `json:"date,omitempty"`
	LastModified  time.Time         `json:"last_modified,omitempty"`
	ContentType   string            `json:"content_type,omitempty"`
	ContentLength int64             `json:"content_length,omitempty"`
}

//ObjectFilter filter object
type ObjectFilter struct {
	Prefix string `json:"prefix,omitempty"`
	Path   string `json:"path,omitempty"`
}

//Range Defines a range of bytes
type Range struct {
	From *int `json:"from,omitempty"`
	To   *int `json:"to,omitempty"`
}

//NewRange creates a range
func NewRange(from, to int) Range {
	return Range{&from, &to}
}

func (r Range) String() string {
	if r.From != nil && r.To != nil {
		return fmt.Sprintf("%d-%d", *r.From, *r.To)
	}
	if r.From != nil {
		return fmt.Sprintf("%d-", *r.From)
	}
	if r.To != nil {
		return fmt.Sprintf("%d", *r.To)
	}
	return ""
}

//ClientAPI is an API defining an IaaS driver
type ClientAPI interface {

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
	CreateNetwork(req NetworkRequest) (*Network, error)
	//GetNetwork returns the network identified by id
	GetNetwork(id string) (*Network, error)
	//ListNetworks lists available networks
	ListNetworks() ([]Network, error)
	//DeleteNetwork deletes the network identified by id
	DeleteNetwork(id string) error

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
	//GetSSHConfig creates SSHConfig from VM
	GetSSHConfig(id string) (*system.SSHConfig, error)

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

	//CreateContainer creates an object container
	CreateContainer(name string) error
	//DeleteContainer deletes an object container
	DeleteContainer(name string) error
	//ListContainers list object containers
	ListContainers() ([]string, error)

	//PutObject put an object into an object container
	PutObject(container string, obj Object) error
	//UpdateObjectMetadata update an object into  object container
	UpdateObjectMetadata(container string, obj Object) error
	//GetObject get  object content from an object container
	GetObject(container string, name string, ranges []Range) (*Object, error)
	//GetObjectMetadata get  object metadata from an object container
	GetObjectMetadata(container string, name string) (*Object, error)
	//ListObjects list objects of a container
	ListObjects(container string, filter ObjectFilter) ([]string, error)
	//CopyObject copies an object
	CopyObject(containerSrc, objectSrc, objectDst string) error
	//DeleteObject deleta an object from a container
	DeleteObject(container, object string) error
}
