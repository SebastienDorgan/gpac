package openstack

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/gob"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/SebastienDorgan/gpac/providers"
	"github.com/SebastienDorgan/gpac/system"

	"github.com/SebastienDorgan/gpac/providers/api"
	"github.com/SebastienDorgan/gpac/providers/api/IPVersion"
	"github.com/SebastienDorgan/gpac/providers/api/VMState"
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
	pager := images.List(client.Compute, opts)

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
	img, err := images.Get(client.Compute, id).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error getting image: %s", errorString(err))
	}
	return &api.Image{ID: img.ID, Name: img.Name}, nil
}

//GetTemplate returns the Template referenced by id
func (client *Client) GetTemplate(id string) (*api.VMTemplate, error) {
	flv, err := flavors.Get(client.Compute, id).Extract()
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
	pager := flavors.ListDetail(client.Compute, opts)

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

	kp, err := keypairs.Create(client.Compute, keypairs.CreateOpts{
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
	kp, err := keypairs.Get(client.Compute, id).Extract()
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
	pager := keypairs.List(client.Compute)

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
	err := keypairs.Delete(client.Compute, id).ExtractErr()
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
		if n == client.Cfg.ProviderNetwork {
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
	vmDef, err := client.readVMDefinition(server.ID)
	if err == nil {
		vm.GatewayID = vmDef.GatewayID
		vm.PrivateKey = vmDef.PrivateKey
		//Floating IP management
		if vm.AccessIPv4 == "" {
			vm.AccessIPv4 = vmDef.AccessIPv4
		}
		if vm.AccessIPv6 == "" {
			vm.AccessIPv6 = vmDef.AccessIPv6
		}

	}
	return &vm
}

//Data structure to apply to userdata.sh template
type userData struct {
	//Name of the default user (api.DefaultUser)
	User string
	//Private key used to create the VM
	Key string
	//If true configure all interfaces to DHCP
	ConfIF bool
	//If true activate IP frowarding
	IsGateway bool
	//If true configure default gateway
	AddGateway bool
	//Content of the /etc/resolve.conf of the Gateway
	//Used only if IsGateway is true
	ResolveConf string
	//IP of the gateway
	GatewayIP string
}

func (client *Client) prepareUserData(request api.VMRequest, kp *api.KeyPair, gw *servers.Server) ([]byte, error) {
	dataBuffer := bytes.NewBufferString("")
	var gwIP string
	var ResolveConf string
	var err error
	if !request.PublicIP {
		gwIP = gw.AccessIPv4
		if gwIP == "" {
			gwIP = gw.AccessIPv6
		}
		var buffer bytes.Buffer
		for _, dns := range client.Cfg.DNSList {
			buffer.WriteString(fmt.Sprintf("nameserver %s\n", dns))
		}
		ResolveConf = buffer.String()
	}

	data := userData{
		User:        api.DefaultUser,
		Key:         strings.Trim(kp.PublicKey, "\n"),
		ConfIF:      !client.Cfg.AutoVMNetworkInterfaces,
		IsGateway:   request.IsGateway && !client.Cfg.UseLayer3Networking,
		AddGateway:  request.PublicIP && !client.Cfg.UseLayer3Networking,
		ResolveConf: ResolveConf,
		GatewayIP:   gwIP,
	}
	err = client.UserDataTpl.Execute(dataBuffer, data)
	if err != nil {
		return nil, err
	}
	return dataBuffer.Bytes(), nil
}

func (client *Client) readGateway(networkID string) (*servers.Server, error) {
	gwID, err := client.getGateway(networkID)
	if err != nil {
		return nil, fmt.Errorf("Error creating VM: Enable to found Gateway %s", errorString(err))
	}
	gw, err := servers.Get(client.Compute, gwID).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error creating VM: Enable to found Gateway %s", errorString(err))
	}
	return gw, nil
}

func (client *Client) saveVMDefinition(vm api.VM) error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(vm)
	if err != nil {
		return err
	}
	return client.PutObject("__vms__", api.Object{
		Name:    vm.ID,
		Content: bytes.NewReader(buffer.Bytes()),
	})
}
func (client *Client) removeVMDefinition(vmID string) error {
	return client.DeleteObject("__vms__", vmID)
}
func (client *Client) readVMDefinition(vmID string) (*api.VM, error) {
	o, err := client.GetObject("__vms__", vmID, nil)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	buffer.ReadFrom(o.Content)
	enc := gob.NewDecoder(&buffer)
	var vm api.VM
	err = enc.Decode(&vm)
	if err != nil {
		return nil, err
	}
	return &vm, nil
}

//CreateVM creates a VM satisfying request
func (client *Client) CreateVM(request api.VMRequest) (*api.VM, error) {
	//Prepare network list
	mainNetID := request.NetworkIDs[0]

	var nets []servers.Network
	//If floating IPs are not used and VM is public
	//then add provider network to VM networks
	if !client.Cfg.UseFloatingIP && request.PublicIP {
		nets = append(nets, servers.Network{
			UUID: client.ProviderNetworkID,
		})
	}
	//Add private networks
	for _, n := range request.NetworkIDs {
		nets = append(nets, servers.Network{
			UUID: n,
		})
	}

	//Prepare key pair
	kp := request.KeyPair
	var err error
	if kp == nil {
		id := uuid.NewV4()
		name := fmt.Sprintf("%s_%s", request.Name, id)
		kp, err = client.CreateKeyPair(name)
		if err != nil {
			return nil, fmt.Errorf("Error creating VM: %s", errorString(err))
		}
		defer client.DeleteKeyPair(kp.ID)
	}

	if err != nil {
		return nil, err
	}
	var gw *servers.Server
	if !request.PublicIP {
		gw, err = client.readGateway(mainNetID)
		if err != nil {
			return nil, fmt.Errorf("Error creating VM: %s", errorString(err))
		}
	}
	userData, err := client.prepareUserData(request, kp, gw)

	//Create VM
	srvOpts := servers.CreateOpts{
		Name:           request.Name,
		SecurityGroups: []string{client.SecurityGroup.Name},
		Networks:       nets,
		FlavorRef:      request.TemplateID,
		ImageRef:       request.ImageID,
		UserData:       userData,
	}
	server, err := servers.Create(client.Compute, keypairs.CreateOptsExt{
		CreateOptsBuilder: srvOpts,
		KeyName:           kp.ID,
	}).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error creating VM: %s", errorString(err))
	}
	//Wait that VM is started
	service := providers.Service{
		ClientAPI: client,
	}
	vm, err := service.WaitVMState(server.ID, VMState.STARTED, 120*time.Second)
	if err != nil {
		return nil, fmt.Errorf("Timeout creating VM: %s", errorString(err))
	}
	//Add gateway ID to VM definition
	var gwID string
	if gw != nil {
		gwID = gw.ID
	}
	vm.GatewayID = gwID
	vm.PrivateKey = kp.PrivateKey
	//if Not use Floating IP or no public address requested
	if !client.Cfg.UseFloatingIP || !request.PublicIP {
		err = client.saveVMDefinition(*vm)
		if err != nil {
			client.DeleteVM(vm.ID)
			return nil, fmt.Errorf("Error creating VM: %s", errorString(err))
		}
		return vm, nil
	}

	//Create the floating IP
	ip, err := floatingip.Create(client.Compute, floatingip.CreateOpts{
		Pool: client.Opts.FloatingIPPool,
	}).Extract()
	if err != nil {
		servers.Delete(client.Compute, vm.ID)
		return nil, fmt.Errorf("Error creating VM: %s", errorString(err))
	}

	//Associate floating IP to VM
	err = floatingip.AssociateInstance(client.Compute, floatingip.AssociateOpts{
		FloatingIP: ip.IP,
		ServerID:   vm.ID,
	}).ExtractErr()
	if err != nil {
		floatingip.Delete(client.Compute, ip.ID)
		servers.Delete(client.Compute, vm.ID)
		return nil, fmt.Errorf("Error creating VM: %s", errorString(err))
	}

	if IPVersion.IPv4.Is(ip.IP) {
		vm.AccessIPv4 = ip.IP
	} else if IPVersion.IPv6.Is(ip.IP) {
		vm.AccessIPv6 = ip.IP
	}
	err = client.saveVMDefinition(*vm)
	if err != nil {
		client.DeleteVM(vm.ID)
		return nil, fmt.Errorf("Error creating VM: %s", errorString(err))
	}

	return vm, nil

}

//GetVM returns the VM identified by id
func (client *Client) GetVM(id string) (*api.VM, error) {
	server, err := servers.Get(client.Compute, id).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error getting VM: %s", errorString(err))
	}
	return client.toVM(server), nil
}

//ListVMs lists available VMs
func (client *Client) ListVMs() ([]api.VM, error) {
	pager := servers.List(client.Compute, servers.ListOpts{})
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
	pager := floatingip.List(client.Compute)
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
	client.readVMDefinition(id)
	if client.Cfg.UseFloatingIP {
		fip, err := client.getFloatingIP(id)
		if err == nil {
			if fip != nil {
				err = floatingip.DisassociateInstance(client.Compute, floatingip.AssociateOpts{
					ServerID:   id,
					FloatingIP: fip.IP,
				}).ExtractErr()
				if err != nil {
					return fmt.Errorf("Error deleting VM %s : %s", id, errorString(err))
				}
				err = floatingip.Delete(client.Compute, fip.ID).ExtractErr()
				if err != nil {
					return fmt.Errorf("Error deleting VM %s : %s", id, errorString(err))
				}
			}
		}
	}
	err := servers.Delete(client.Compute, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error deleting VM %s : %s", id, errorString(err))
	}
	return nil
}

//StopVM stops the VM identified by id
func (client *Client) StopVM(id string) error {
	err := startstop.Stop(client.Compute, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error stoping VM : %s", errorString(err))
	}
	return nil
}

//StartVM starts the VM identified by id
func (client *Client) StartVM(id string) error {
	err := startstop.Start(client.Compute, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error stoping VM : %s", errorString(err))
	}
	return nil
}

func (client *Client) getSSHConfig(vm *api.VM) (*system.SSHConfig, error) {

	ip := vm.GetAccessIP()
	sshConfig := system.SSHConfig{
		PrivateKey:         vm.PrivateKey,
		Port:               22,
		Host:               ip,
		ConnecttionTimeout: 30 * time.Second,
		User:               api.DefaultUser,
	}
	if vm.GatewayID != "" {
		gw, err := client.GetVM(vm.GatewayID)
		if err != nil {
			return nil, err
		}
		ip := gw.GetAccessIP()
		GatewayConfig := system.SSHConfig{
			PrivateKey:         gw.PrivateKey,
			Port:               22,
			ConnecttionTimeout: 30 * time.Second,
			User:               api.DefaultUser,
			Host:               ip,
		}
		sshConfig.GatewayConfig = &GatewayConfig
	}

	return &sshConfig, nil

}

//GetSSHConfig creates SSHConfig to connect a VM
func (client *Client) GetSSHConfig(id string) (*system.SSHConfig, error) {
	vm, err := client.GetVM(id)
	if err != nil {
		return nil, err
	}
	return client.getSSHConfig(vm)
}
