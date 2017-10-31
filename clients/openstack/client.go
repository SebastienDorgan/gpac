package openstack

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sort"

	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/layer3/routers"

	api "github.com/SebastienDorgan/gpac/clients"
	gc "github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/secgroups"
	"github.com/rackspace/gophercloud/openstack/imageservice/v2/images"
	"github.com/rackspace/gophercloud/openstack/networking/v2/networks"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/flavors"
	"golang.org/x/crypto/ssh"
)

/*AuthOptions fields are the union of those recognized by each identity implementation and
provider.
*/
type AuthOptions struct {
	// IdentityEndpoint specifies the HTTP endpoint that is required to work with
	// the Identity API of the appropriate version. While it's ultimately needed by
	// all of the identity services, it will often be populated by a provider-level
	// function.
	IdentityEndpoint string

	// Username is required if using Identity V2 API. Consult with your provider's
	// control panel to discover your account's username. In Identity V3, either
	// UserID or a combination of Username and DomainID or DomainName are needed.
	Username, UserID string

	// Exactly one of Password or APIKey is required for the Identity V2 and V3
	// APIs. Consult with your provider's control panel to discover your account's
	// preferred method of authentication.
	Password, APIKey string

	// At most one of DomainID and DomainName must be provided if using Username
	// with Identity V3. Otherwise, either are optional.
	DomainID, DomainName string

	// The TenantID and TenantName fields are optional for the Identity V2 API.
	// Some providers allow you to specify a TenantName instead of the TenantId.
	// Some require both. Your provider's authentication policies will determine
	// how these fields influence authentication.
	TenantID, TenantName string

	// AllowReauth should be set to true if you grant permission for Gophercloud to
	// cache your credentials in memory, and to allow Gophercloud to attempt to
	// re-authenticate automatically if/when your token expires.  If you set it to
	// false, it will not cache these settings, but re-authentication will not be
	// possible.  This setting defaults to false.
	//
	// NOTE: The reauth function will try to re-authenticate endlessly if left unchecked.
	// The way to limit the number of attempts is to provide a custom HTTP client to the provider client
	// and provide a transport that implements the RoundTripper interface and stores the number of failed retries.
	// For an example of this, see here: https://github.com/rackspace/rack/blob/1.0.0/auth/clients.go#L311
	AllowReauth bool

	// TokenID allows users to authenticate (possibly as another user) with an
	// authentication token ID.
	TokenID string

	//Openstack region (data center) where the infrstructure will be created
	Region string

	//Name of the provider (external) network
	ProviderNetwork string

	//UseFloatingIP indicates if floating IP are used
	UseFloatingIP bool
}

//AuthenticatedClient returns an authenticated client
func AuthenticatedClient(opts AuthOptions) (*Client, error) {
	gcOpts := gc.AuthOptions{
		IdentityEndpoint: opts.IdentityEndpoint,
		Username:         opts.Username,
		UserID:           opts.UserID,
		Password:         opts.Password,
		APIKey:           opts.APIKey,
		DomainID:         opts.DomainID,
		DomainName:       opts.DomainName,
		TenantID:         opts.TenantID,
		TenantName:       opts.TenantName,
		AllowReauth:      opts.AllowReauth,
		TokenID:          opts.TokenID,
	}
	pClient, err := openstack.AuthenticatedClient(gcOpts)
	if err != nil {
		return nil, err
	}
	compute, err := openstack.NewComputeV2(pClient, gc.EndpointOpts{
		Region: opts.Region,
	})

	if err != nil {
		return nil, err
	}
	network, err := openstack.NewNetworkV2(pClient, gc.EndpointOpts{
		Region: opts.Region,
	})
	if err != nil {
		return nil, err
	}
	clt := Client{
		Opts:    &opts,
		pClient: pClient,
		nova:    compute,
		neutron: network,
	}
	err = clt.initDefaultRouter()
	if err != nil {
		return nil, err
	}
	err = clt.initDefaultSecurityGroup()
	if err != nil {
		return nil, err
	}
	return &clt, nil
}

const defaultRouter string = "d46886b1-cb8e-4e98-9b18-b60bf847dd09"
const defaultSecurityGroup string = "30ad3142-a5ec-44b5-9560-618bde3de1ef"

//Client is the implementation of the openstack driver regarding to the api.ClientAPI
type Client struct {
	Opts          *AuthOptions
	pClient       *gc.ProviderClient
	nova          *gc.ServiceClient
	neutron       *gc.ServiceClient
	router        *routers.Router
	securityGroup *secgroups.SecurityGroup
}

func (client *Client) getProviderNetwork() (*networks.Network, error) {
	var nets []networks.Network
	opts := networks.ListOpts{Name: client.Opts.ProviderNetwork}
	// Retrieve a pager (i.e. a paginated collection)
	err := networks.List(client.neutron, opts).EachPage(func(page pagination.Page) (bool, error) {
		list, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, err
		}
		nets = append(nets, list...)
		return true, nil
	})
	//nets, err := networks.ExtractNetworks(page)
	if err != nil {
		return nil, err
	}
	size := len(nets)
	if size == 0 {
		return nil, fmt.Errorf("Network %s does not exist", client.Opts.ProviderNetwork)
	}
	if size > 1 {
		return nil, fmt.Errorf("Configuration error: 2 instances of network %s exist", client.Opts.ProviderNetwork)
	}
	return &nets[0], nil
}

func (client *Client) createDefaultRouter() error {
	pNet, err := client.getProviderNetwork()
	if err != nil {
		return fmt.Errorf("Error retriving Provider network %s: %s", client.Opts.ProviderNetwork, err.Error())
	}
	//Create a router to connect external Provider network
	gi := routers.GatewayInfo{
		NetworkID: pNet.ID,
	}
	opts := routers.CreateOpts{
		Name:         defaultRouter,
		AdminStateUp: networks.Up,
		GatewayInfo:  &gi,
	}
	router, err := routers.Create(client.neutron, opts).Extract()
	if err != nil {
		return err
	}
	client.router = router
	return nil

}

func (client *Client) getDefaultSecurityGroup() (*secgroups.SecurityGroup, error) {
	var sgList []secgroups.SecurityGroup

	err := secgroups.List(client.nova).EachPage(func(page pagination.Page) (bool, error) {
		list, err := secgroups.ExtractSecurityGroups(page)
		if err != nil {
			return false, err
		}
		for _, e := range list {
			if e.Name == defaultSecurityGroup {
				sgList = append(sgList, e)
			}
		}
		return true, nil
	})
	if len(sgList) == 0 {
		return nil, err
	}
	if len(sgList) > 1 {
		return nil, fmt.Errorf("Configuration error: More than one default security groups exists")
	}

	return &sgList[0], nil
}

func (client *Client) createTCPRules(groupID string) error {
	//Open TCP Ports
	ruleOpts := secgroups.CreateRuleOpts{
		ParentGroupID: groupID,
		FromPort:      1,
		ToPort:        65535,
		IPProtocol:    "TCP",
		CIDR:          "0.0.0.0/0",
	}

	_, err := secgroups.CreateRule(client.nova, ruleOpts).Extract()
	if err != nil {
		return err
	}
	ruleOpts = secgroups.CreateRuleOpts{
		ParentGroupID: groupID,
		FromPort:      1,
		ToPort:        65535,
		IPProtocol:    "TCP",
		CIDR:          "::/0",
	}
	_, err = secgroups.CreateRule(client.nova, ruleOpts).Extract()
	return err
}
func (client *Client) createUDPRules(groupID string) error {
	//Open UDP Ports
	ruleOpts := secgroups.CreateRuleOpts{
		ParentGroupID: groupID,
		FromPort:      1,
		ToPort:        65535,
		IPProtocol:    "UDP",
		CIDR:          "0.0.0.0/0",
	}

	_, err := secgroups.CreateRule(client.nova, ruleOpts).Extract()
	if err != nil {
		return err
	}
	ruleOpts = secgroups.CreateRuleOpts{
		ParentGroupID: groupID,
		FromPort:      1,
		ToPort:        65535,
		IPProtocol:    "UDP",
		CIDR:          "::/0",
	}
	_, err = secgroups.CreateRule(client.nova, ruleOpts).Extract()
	return err
}
func (client *Client) createICMPRules(groupID string) error {
	//Open TCP Ports
	ruleOpts := secgroups.CreateRuleOpts{
		ParentGroupID: groupID,
		FromPort:      -1,
		ToPort:        -1,
		IPProtocol:    "ICMP",
		CIDR:          "0.0.0.0/0",
	}

	_, err := secgroups.CreateRule(client.nova, ruleOpts).Extract()
	if err != nil {
		return err
	}
	ruleOpts = secgroups.CreateRuleOpts{
		ParentGroupID: groupID,
		FromPort:      -1,
		ToPort:        -1,
		IPProtocol:    "ICMP",
		CIDR:          "::/0",
	}
	_, err = secgroups.CreateRule(client.nova, ruleOpts).Extract()
	return err
}

//initDefaultSecurityGroup create an open Security Group
//The default security group opens all TCP, UDP, ICMP ports
//Security is managed individually on each VM using a linux firewall
func (client *Client) initDefaultSecurityGroup() error {
	sg, err := client.getDefaultSecurityGroup()
	if err != nil {
		return err
	}
	if sg != nil {
		client.securityGroup = sg
		return nil
	}
	opts := secgroups.CreateOpts{
		Name:        defaultSecurityGroup,
		Description: "Gpac default security group",
	}

	group, err := secgroups.Create(client.nova, opts).Extract()
	if err != nil {
		return err
	}
	err = client.createTCPRules(group.ID)
	if err != nil {
		secgroups.Delete(client.nova, group.ID)
		return err
	}

	err = client.createUDPRules(group.ID)
	if err != nil {
		secgroups.Delete(client.nova, group.ID)
		return err
	}
	err = client.createICMPRules(group.ID)
	if err != nil {
		secgroups.Delete(client.nova, group.ID)
		return err
	}
	client.securityGroup = group
	return nil
}

func (client *Client) initDefaultRouter() error {
	//check if the default router exists
	var routerList []routers.Router
	err := routers.List(client.neutron, routers.ListOpts{
		Name: defaultRouter,
	}).EachPage(func(page pagination.Page) (bool, error) {
		list, err := routers.ExtractRouters(page)
		if err != nil {
			return false, err
		}
		routerList = append(routerList, list...)
		return true, nil
	})

	if err != nil {
		return err
	}
	size := len(routerList)
	//the default router exists
	if size == 1 {
		client.router = &routerList[0]
		return nil
	}
	if size > 1 {
		return fmt.Errorf("Configuration error, more than one default router is defined")
	}
	return client.createDefaultRouter()
}

//ListImages lists available OS images
func (client *Client) ListImages() ([]api.Image, error) {
	opts := images.ListOpts{}

	// Retrieve a pager (i.e. a paginated collection)
	pager := images.List(client.nova, opts)

	var imgList []api.Image

	// Define an anonymous function to be executed on each page's iteration
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		imageList, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}

		for _, img := range imageList {
			imgList = append(imgList, api.Image{ID: img.ID, Name: img.Name})

		}
		return true, nil
	})
	if len(imgList) == 0 {
		if err != nil {
			return nil, err
		}
	}
	return imgList, nil
}

//ListTemplates lists available VM templates
//VM templates are sorted using Dominant Resource Fairness Algorithm
func (client *Client) ListTemplates() ([]api.VMTemplate, error) {
	opts := flavors.ListOpts{}

	// Retrieve a pager (i.e. a paginated collection)
	pager := flavors.ListDetail(client.nova, opts)

	var flvList []api.VMTemplate

	// Define an anonymous function to be executed on each page's iteration
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		flavorList, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}

		for _, flv := range flavorList {
			flvList = append(flvList, api.VMTemplate{
				ID:       flv.ID,
				Name:     flv.Name,
				Cores:    flv.VCPUs,
				DiskSize: flv.Disk,
				RAMSize:  float32(flv.RAM) / 1000.0,
			})

		}
		return true, nil
	})
	if len(flvList) == 0 {
		if err != nil {
			return nil, err
		}
	}
	sort.Sort(api.ByRankDRF(flvList))
	return flvList, nil
}

//SelectTemplates lists VM templates compatible with sizing requirements
//VM templates are sorted using Dominant Resource Fairness Algorithm
func (client *Client) SelectTemplates(sizing api.SizingRequirements) ([]api.VMTemplate, error) {
	opts := flavors.ListOpts{
		MinDisk: sizing.MinDiskSize,
		MinRAM:  int(sizing.MinRAMSize * 1000.0),
	}

	// Retrieve a pager (i.e. a paginated collection)
	pager := flavors.ListDetail(client.nova, opts)

	var flvList []api.VMTemplate

	// Define an anonymous function to be executed on each page's iteration
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		flavorList, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}
		for _, flv := range flavorList {
			if flv.VCPUs >= sizing.MinCores {
				flvList = append(flvList, api.VMTemplate{
					ID:       flv.ID,
					Name:     flv.Name,
					Cores:    flv.VCPUs,
					DiskSize: flv.Disk,
					RAMSize:  float32(flv.RAM) / 1000.0,
				})
			}
		}
		return true, nil
	})
	if len(flvList) == 0 {
		if err != nil {
			return nil, err
		}
	}
	sort.Sort(api.ByRankDRF(flvList))
	return flvList, nil
}

//CreateKeyPair creates and import a key pair
func (client *Client) CreateKeyPair(name string) (*api.KeyPair, error) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := privateKey.PublicKey
	pub, _ := ssh.NewPublicKey(&publicKey)
	pubBytes := ssh.MarshalAuthorizedKey(pub)
	pubKey := string(pubBytes)

	priBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	priKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: priBytes,
		},
	)
	priKey := string(priKeyPem)

	kp, err := keypairs.Create(client.nova, keypairs.CreateOpts{
		Name:      name,
		PublicKey: pubKey,
	}).Extract()
	if err != nil {
		return nil, err
	}
	return &api.KeyPair{
		ID:         kp.Name,
		Name:       kp.Name,
		PublicKey:  kp.PublicKey,
		PrivateKey: priKey,
	}, nil
}

//GetKeyPair returns the key pair identified by id
func (client *Client) GetKeyPair(id string) (*api.KeyPair, error) {
	kp, err := keypairs.Get(client.nova, id).Extract()
	if err != nil {
		return nil, err
	}
	return &api.KeyPair{
		ID:         kp.Name,
		Name:       kp.Name,
		PrivateKey: kp.PrivateKey,
		PublicKey:  kp.PublicKey,
	}, nil
}

//ListKeyPairs lists available key pairs
func (client *Client) ListKeyPairs() ([]api.KeyPair, error) {

	// Retrieve a pager (i.e. a paginated collection)
	pager := keypairs.List(client.nova)

	var kpList []api.KeyPair

	// Define an anonymous function to be executed on each page's iteration
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		keyList, err := keypairs.ExtractKeyPairs(page)
		if err != nil {
			return false, err
		}

		for _, kp := range keyList {
			kpList = append(kpList, api.KeyPair{
				ID:         kp.Name,
				Name:       kp.Name,
				PublicKey:  kp.PublicKey,
				PrivateKey: kp.PrivateKey,
			})

		}
		return true, nil
	})
	if len(kpList) == 0 {
		if err != nil {
			return nil, err
		}
	}
	return kpList, nil
}

//DeleteKeyPair deletes the key pair identified by id
func (client *Client) DeleteKeyPair(id string) error {
	return keypairs.Delete(client.nova, id).ExtractErr()
}

//CreateNetwork creates a network named name
func (client *Client) CreateNetwork(name string) (*api.Network, error) {
	// We specify a name and that it should forward packets
	opts := networks.CreateOpts{
		Name:         name,
		AdminStateUp: networks.Up,
	}

	// Execute the operation and get back a networks.Network struct
	network, err := networks.Create(client.neutron, opts).Extract()
	if err != nil {
		return nil, err
	}

	return &api.Network{
		ID:   network.ID,
		Name: network.Name,
	}, nil

}

//GetNetwork returns the network identified by id
func (client *Client) GetNetwork(id string) (*api.Network, error) {
	network, err := networks.Get(client.neutron, id).Extract()
	if err != nil {
		return nil, err
	}
	return &api.Network{
		ID:      network.ID,
		Name:    network.Name,
		Subnets: network.Subnets,
	}, nil
}

//ListNetworks lists available networks
func (client *Client) ListNetworks() ([]api.Network, error) {
	// We have the option of filtering the network list. If we want the full
	// collection, leave it as an empty struct
	opts := networks.ListOpts{}

	// Retrieve a pager (i.e. a paginated collection)
	pager := networks.List(client.neutron, opts)
	var netList []api.Network
	// Define an anonymous function to be executed on each page's iteration
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		networkList, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, err
		}

		for _, n := range networkList {
			netList = append(netList, api.Network{
				ID:      n.ID,
				Name:    n.Name,
				Subnets: n.Subnets,
			})
		}
		return true, nil
	})
	if len(netList) == 0 && err != nil {
		return netList, err
	}
	return netList, nil
}

//DeleteNetwork deletes the network identified by id
func (client *Client) DeleteNetwork(id string) error {
	return networks.Delete(client.neutron, id).ExtractErr()
}

func toGCIPversion(v api.IPVersion) int {
	if v == api.IPv4 {
		return subnets.IPv4
	} else if v == api.IPv6 {
		return subnets.IPv6
	}
	return -1
}

func fromGopherIPversion(v int) api.IPVersion {
	if v == 4 {
		return subnets.IPv4
	} else if v == 6 {
		return subnets.IPv6
	}
	return -1
}

//CreateSubnet creates a sub network
//- netID ID of the parent network
//- name is the name of the sub network
//- mask is a network mask defined in CIDR notation
func (client *Client) CreateSubnet(request api.SubnetRequets) (*api.Subnet, error) {
	// You must associate a new subnet with an existing network - to do this you
	// need its UUID. You must also provide a well-formed CIDR value.
	//addr, _, err := net.ParseCIDR(mask)
	dhcp := true
	opts := subnets.CreateOpts{
		NetworkID:  request.NetworkID,
		CIDR:       request.Mask,
		IPVersion:  toGCIPversion(request.IPVersion),
		Name:       request.Name,
		EnableDHCP: &dhcp,

		//GatewayIP:  addr.String(),
	}
	// Execute the operation and get back a subnets.Subnet struct
	subnet, err := subnets.Create(client.neutron, opts).Extract()
	if err != nil {
		return nil, err
	}
	_, err = routers.AddInterface(client.neutron, client.router.ID, routers.InterfaceOpts{
		SubnetID: subnet.ID,
	}).Extract()
	if err != nil {
		client.DeleteSubnet(subnet.ID)
		return nil, err
	}
	return &api.Subnet{
		ID:        subnet.ID,
		Name:      subnet.Name,
		IPVersion: fromGopherIPversion(subnet.IPVersion),
		Mask:      subnet.CIDR,
		NetworkID: subnet.NetworkID,
	}, nil
}

//GetSubnet returns the sub network identified by id
func (client *Client) GetSubnet(id string) (*api.Subnet, error) {
	// Execute the operation and get back a subnets.Subnet struct
	subnet, err := subnets.Get(client.neutron, id).Extract()
	if err != nil {
		return nil, err
	}
	return &api.Subnet{
		ID:        subnet.ID,
		Name:      subnet.Name,
		IPVersion: fromGopherIPversion(subnet.IPVersion),
		Mask:      subnet.CIDR,
		NetworkID: subnet.NetworkID,
	}, nil
}

//ListSubnets lists available sub networks of network net
func (client *Client) ListSubnets(netID string) ([]api.Subnet, error) {
	pager := subnets.List(client.neutron, subnets.ListOpts{
		NetworkID: netID,
	})
	var subnetList []api.Subnet
	pager.EachPage(func(page pagination.Page) (bool, error) {
		list, err := subnets.ExtractSubnets(page)
		if err != nil {
			return false, err
		}

		for _, subnet := range list {
			subnetList = append(subnetList, api.Subnet{
				ID:        subnet.ID,
				Name:      subnet.Name,
				IPVersion: fromGopherIPversion(subnet.IPVersion),
				Mask:      subnet.CIDR,
				NetworkID: subnet.NetworkID,
			})
		}
		return true, nil
	})
	return subnetList, nil
}

//DeleteSubnet deletes the sub network identified by id
func (client *Client) DeleteSubnet(id string) error {
	_, err := routers.RemoveInterface(client.neutron, client.router.ID, routers.InterfaceOpts{
		SubnetID: id,
	}).Extract()
	if err != nil {
		return err
	}
	return subnets.Delete(client.neutron, id).ExtractErr()
}

//CreateVM creates a VM
func (client *Client) CreateVM(request api.VMRequest) (api.VM, error) {
	panic("Not implemented")
}

//GetVM returns the VM identified by id
func (client *Client) GetVM(id string) (api.VM, error) {
	panic("Not implemented")
}

//ListVMs lists available VMs
func (client *Client) ListVMs() ([]api.VM, error) {
	panic("Not implemented")
}

//DeleteVM deletes the VM identified by id
func (client *Client) DeleteVM(id string) error {
	panic("Not implemented")
}

//StopVM stops the VM identified by id
func (client *Client) StopVM(id string) error {
	panic("Not implemented")
}

//StartVM starts the VM identified by id
func (client *Client) StartVM(id string) error {
	panic("Not implemented")
}

//CreateVolume creates a block volume
//- name is the name of the volume
//- size is the size of the volume in GB
//- volumeType is the type of volume to create, if volumeType is empty the driver use a default type
func (client *Client) CreateVolume(name string, size float32, volumeType string) (api.Volume, error) {
	panic("Not implemented")
}

//GetVolume returns the volume identified by id
func (client *Client) GetVolume(id string) (api.Volume, error) {
	panic("Not implemented")
}

//ListVolumes list available volumes
func (client *Client) ListVolumes() ([]api.Volume, error) {
	panic("Not implemented")
}

//DeleteVolume deletes the volume identified by id
func (client *Client) DeleteVolume(id string) error {
	panic("Not implemented")
}

//CreateVolumeAttachment attaches a volume to a VM
//- name the name of the volume attachment
//- volume the volume to attach
//- vm the VM on which the volume is attached
func (client *Client) CreateVolumeAttachment(name string, volume api.Volume, vm api.VM) (api.VolumeAttachment, error) {
	panic("Not implemented")
}

//GetVolumeAttachment returns the volume attachment identified by id
func (client *Client) GetVolumeAttachment(id string) (api.VolumeAttachment, error) {
	panic("Not implemented")
}

//ListVolumeAttachments lists available volume attachment
func (client *Client) ListVolumeAttachments() ([]api.VolumeAttachment, error) {
	panic("Not implemented")
}

//DeleteVolumeAttachment deletes the volume attachment identifed by id
func (client *Client) DeleteVolumeAttachment(id string) error {
	panic("Not implemented")
}
