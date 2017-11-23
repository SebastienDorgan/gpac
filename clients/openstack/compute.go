package openstack

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/SebastienDorgan/gpac/clients"
	"github.com/SebastienDorgan/gpac/clients/api"
	"github.com/SebastienDorgan/gpac/clients/api/IPVersion"
	"github.com/SebastienDorgan/gpac/clients/api/VMState"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/floatingip"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/openstack/imageservice/v2/images"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/flavors"
	"golang.org/x/crypto/ssh"
)

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
			return nil, fmt.Errorf("Error listing images: %s", errorString(err))
		}
	}
	return imgList, nil
}

//GetImage returns the Image referenced by id
func (client *Client) GetImage(id string) (*api.Image, error) {
	img, err := images.Get(client.nova, id).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error getting image: %s", errorString(err))
	}
	return &api.Image{ID: img.ID, Name: img.Name}, nil
}

//GetTemplate returns the Template referenced by id
func (client *Client) GetTemplate(id string) (*api.VMTemplate, error) {
	flv, err := flavors.Get(client.nova, id).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error getting template: %s", errorString(err))
	}
	return &api.VMTemplate{
		VMSize: api.VMSize{
			Cores:    flv.VCPUs,
			RAMSize:  float32(flv.RAM) / 1000.0,
			DiskSize: flv.Disk,
		},
		ID:   flv.ID,
		Name: flv.Name,
	}, nil
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
				VMSize: api.VMSize{
					Cores:    flv.VCPUs,
					RAMSize:  float32(flv.RAM) / 1000.0,
					DiskSize: flv.Disk,
				},
				ID:   flv.ID,
				Name: flv.Name,
			})

		}
		return true, nil
	})
	if len(flvList) == 0 {
		if err != nil {
			return nil, err
		}
	}
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
	err := keypairs.Delete(client.nova, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error deleting key pair: %s", errorString(err))
	}
	return nil
}

//toVMSize converts flavor attributes returned by OpenStack driver into api.VM
func (client *Client) toVMSize(flavor map[string]interface{}) api.VMSize {
	if i, ok := flavor["id"]; ok {
		fid := i.(string)
		tpl, _ := client.GetTemplate(fid)
		return tpl.VMSize
	}
	if _, ok := flavor["vcpus"]; ok {
		return api.VMSize{
			Cores:    flavor["vcpus"].(int),
			DiskSize: flavor["disk"].(int),
			RAMSize:  flavor["ram"].(float32) / 1000.0,
		}
	}
	return api.VMSize{}
}

//toVMState converts VM status returned by OpenStack driver into VMState enum
func toVMState(status string) VMState.Enum {
	switch status {
	case "BUILD", "build", "BUILDING", "building":
		return VMState.STARTING
	case "ACTIVE", "active":
		return VMState.STARTED
	case "RESCUED", "rescued":
		return VMState.STOPPING
	case "STOPPED", "stopped", "SHUTOFF", "shutoff":
		return VMState.STOPPED
	default:
		return VMState.ERROR
	}
}

//convertAdresses converts adresses returned by the OpenStack driver arrange them by version in a map
func (client *Client) convertAdresses(addresses map[string]interface{}) map[IPVersion.Enum][]string {
	addrs := make(map[IPVersion.Enum][]string)
	for n, obj := range addresses {
		if n == client.Opts.ProviderNetwork {
			break
		}
		for _, networkAddresses := range obj.([]interface{}) {
			address := networkAddresses.(map[string]interface{})
			version := address["version"].(float64)
			fixedIP := address["addr"].(string)
			switch version {
			case 4:
				addrs[IPVersion.IPv4] = append(addrs[IPVersion.IPv4], fixedIP)
			case 6:
				addrs[IPVersion.IPv6] = append(addrs[IPVersion.IPv4], fixedIP)
			}
		}
	}
	return addrs
}

//toVM converts an OpenStack server into api VM
func (client *Client) toVM(server *servers.Server) *api.VM {
	adresses := client.convertAdresses(server.Addresses)
	vm := api.VM{
		ID:           server.ID,
		Name:         server.Name,
		PrivateIPsV4: adresses[IPVersion.IPv4],
		PrivateIPsV6: adresses[IPVersion.IPv6],
		AccessIPv4:   server.AccessIPv4,
		AccessIPv6:   server.AccessIPv6,
		Size:         client.toVMSize(server.Flavor),
		State:        toVMState(server.Status),
	}
	return &vm
}

//CreateVM creates a VM satisfying request
func (client *Client) CreateVM(request api.VMRequest) (*api.VM, error) {
	//Prepare network list
	var nets []servers.Network
	for _, n := range request.NetworkIDs {
		nets = append(nets, servers.Network{
			UUID: n,
		})
	}

	//Prepare user data
	dataBuffer := bytes.NewBufferString("")
	type Data struct {
		User, Key string
	}
	data := Data{
		User: api.DefaultUser,
		Key:  strings.Trim(request.KeyPair.PublicKey, "\n"),
	}
	fmt.Println(data.Key)
	err := client.userDataTpl.Execute(dataBuffer, data)
	if err != nil {
		return nil, err
	}
	fmt.Println(dataBuffer.String())
	//Create VM
	srvOpts := servers.CreateOpts{
		Name:           request.Name,
		SecurityGroups: []string{client.securityGroup.Name},
		Networks:       nets,
		FlavorRef:      request.TemplateID,
		ImageRef:       request.ImageID,
		UserData:       dataBuffer.Bytes(),
	}
	server, err := servers.Create(client.nova, keypairs.CreateOptsExt{
		CreateOptsBuilder: srvOpts,
		KeyName:           request.KeyPair.ID,
	}).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error creating VM: %s", errorString(err))
	}

	//Wait that VM is started
	vm, err := clients.WaitVMState(client, server.ID, VMState.STARTED, 120*time.Second)
	if err != nil {
		return nil, fmt.Errorf("Timeout creating VM: %s", errorString(err))
	}
	//Not use Floating IP or no public address requested
	if !client.Opts.UseFloatingIP || !request.PublicIP {
		return vm, nil
	}

	//Create the floating IP
	ip, err := floatingip.Create(client.nova, floatingip.CreateOpts{
		Pool: client.Opts.FloatingIPPool,
	}).Extract()
	if err != nil {
		servers.Delete(client.nova, vm.ID)
		return nil, fmt.Errorf("Error creating VM: %s", errorString(err))
	}

	//Associate floating IP to VM
	err = floatingip.AssociateInstance(client.nova, floatingip.AssociateOpts{
		FloatingIP: ip.IP,
		ServerID:   vm.ID,
	}).ExtractErr()
	if err != nil {
		floatingip.Delete(client.nova, ip.ID)
		servers.Delete(client.nova, vm.ID)
		return nil, fmt.Errorf("Error creating VM: %s", errorString(err))
	}

	if IPVersion.IPv4.Is(ip.IP) {
		vm.AccessIPv4 = ip.IP
	} else if IPVersion.IPv6.Is(ip.IP) {
		vm.AccessIPv6 = ip.IP
	}
	return vm, nil

}

//GetVM returns the VM identified by id
func (client *Client) GetVM(id string) (*api.VM, error) {
	server, err := servers.Get(client.nova, id).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error getting VM: %s", errorString(err))
	}
	fmt.Println(server.Status)
	return client.toVM(server), nil
}

//ListVMs lists available VMs
func (client *Client) ListVMs() ([]api.VM, error) {
	pager := servers.List(client.nova, servers.ListOpts{})
	var vms []api.VM
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		list, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}

		for _, srv := range list {
			vms = append(vms, *client.toVM(&srv))
		}
		return true, nil
	})
	if len(vms) == 0 && err != nil {
		return nil, fmt.Errorf("Error listing vms : %s", errorString(err))
	}
	return vms, nil
}

//getFloatingIP returns the floating IP associated with the VM identified by vmID
//By convention only one floating IP is allocated to a VM
func (client *Client) getFloatingIP(vmID string) (*floatingip.FloatingIP, error) {
	pager := floatingip.List(client.nova)
	var fips []floatingip.FloatingIP
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		list, err := floatingip.ExtractFloatingIPs(page)
		if err != nil {
			return false, err
		}

		for _, fip := range list {
			if fip.InstanceID == vmID {
				fips = append(fips, fip)
			}
		}
		return true, nil
	})
	if len(fips) == 0 {
		if err != nil {
			return nil, fmt.Errorf("No floating IP found for VM %s: %s", vmID, errorString(err))
		}
		return nil, fmt.Errorf("No floating IP found for VM %s", vmID)

	}
	if len(fips) > 1 {
		return nil, fmt.Errorf("Configuration error, more than one Floating IP associated to VM %s", vmID)
	}
	return &fips[0], nil
}

//DeleteVM deletes the VM identified by id
func (client *Client) DeleteVM(id string) error {
	if client.Opts.UseFloatingIP {
		fip, err := client.getFloatingIP(id)
		if err == nil {
			if fip != nil {
				err = floatingip.DisassociateInstance(client.nova, floatingip.AssociateOpts{
					ServerID:   id,
					FloatingIP: fip.IP,
				}).ExtractErr()
				if err != nil {
					return fmt.Errorf("Error deleting VM %s : %s", id, errorString(err))
				}
				err = floatingip.Delete(client.nova, fip.ID).ExtractErr()
				if err != nil {
					return fmt.Errorf("Error deleting VM %s : %s", id, errorString(err))
				}
			}
		}
	}
	err := servers.Delete(client.nova, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error deleting VM %s : %s", id, errorString(err))
	}
	return nil
}

//StopVM stops the VM identified by id
func (client *Client) StopVM(id string) error {
	err := startstop.Stop(client.nova, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error stoping VM : %s", errorString(err))
	}
	return nil
}

//StartVM starts the VM identified by id
func (client *Client) StartVM(id string) error {
	err := startstop.Start(client.nova, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error stoping VM : %s", errorString(err))
	}
	return nil
}
