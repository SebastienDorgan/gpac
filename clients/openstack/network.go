package openstack

import (
	"fmt"

	"github.com/SebastienDorgan/gpac/clients/api"
	"github.com/SebastienDorgan/gpac/clients/api/IPVersion"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/rackspace/gophercloud/openstack/networking/v2/networks"
	"github.com/rackspace/gophercloud/openstack/networking/v2/subnets"
	"github.com/rackspace/gophercloud/pagination"
)

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
		return nil, fmt.Errorf("Error getting network: %s", errorString(err))
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
		return nil, fmt.Errorf("Error listing networks: %s", errorString(err))
	}
	return netList, nil
}

//DeleteNetwork deletes the network identified by id
func (client *Client) DeleteNetwork(id string) error {
	err := networks.Delete(client.neutron, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error deleting network: %s", errorString(err))
	}
	return nil
}

func toGopherIPversion(v IPVersion.Enum) int {
	if v == IPVersion.IPv4 {
		return subnets.IPv4
	} else if v == IPVersion.IPv6 {
		return subnets.IPv6
	}
	return -1
}

func fromGopherIPversion(v int) IPVersion.Enum {
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
		IPVersion:  toGopherIPversion(request.IPVersion),
		Name:       request.Name,
		EnableDHCP: &dhcp,

		//GatewayIP:  addr.String(),
	}
	// Execute the operation and get back a subnets.Subnet struct
	subnet, err := subnets.Create(client.neutron, opts).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error creating subnet: %s", errorString(err))
	}
	router, err := client.createDefaultRouter(subnet.ID)
	_, err = routers.AddInterface(client.neutron, router.ID, routers.InterfaceOpts{
		SubnetID: subnet.ID,
	}).Extract()
	if err != nil {
		client.DeleteSubnet(subnet.ID)
		return nil, fmt.Errorf("Error creating subnet: %s", errorString(err))
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
		return nil, fmt.Errorf("Error getting subnet: %s", errorString(err))
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
			return false, fmt.Errorf("Error listing subnets: %s", errorString(err))
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
	routerList, _ := client.ListRouter()
	var router *api.Router
	for _, r := range routerList {
		if r.Name == id {
			router = &r
			break
		}
	}
	var err error
	if router != nil {
		_, err = routers.RemoveInterface(client.neutron, router.ID, routers.InterfaceOpts{
			SubnetID: id,
		}).Extract()
	}

	err2 := subnets.Delete(client.neutron, id).ExtractErr()
	if err != nil && err2 != nil {
		return fmt.Errorf("Error deleting subnets: %s", errorString(err))
	}
	if err2 != nil {
		return fmt.Errorf("Error deleting subnets: %s", errorString(err2))
	}
	err = client.DeleteRouter(id)
	if err != nil && err2 != nil {
		return fmt.Errorf("Error deleting subnets: %s", errorString(err))
	}
	return nil
}

//getProviderNetwork returns the provider network
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

//createDefaultRouter create a router on the provider network
func (client *Client) createDefaultRouter(name string) (*api.Router, error) {
	pNet, err := client.getProviderNetwork()
	if err != nil {
		return nil, fmt.Errorf("Error retriving Provider network %s: %s", client.Opts.ProviderNetwork, errorString(err))
	}

	return client.CreateRouter(api.RouterRequest{
		Name:      name,
		NetworkID: pNet.ID,
	})

}

//CreateRouter creates a router satisfying req
func (client *Client) CreateRouter(req api.RouterRequest) (*api.Router, error) {
	//Create a router to connect external Provider network
	gi := routers.GatewayInfo{
		NetworkID: req.NetworkID,
	}
	opts := routers.CreateOpts{
		Name:         req.Name,
		AdminStateUp: networks.Up,
		GatewayInfo:  &gi,
	}
	router, err := routers.Create(client.neutron, opts).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error creating Router: %s", errorString(err))
	}
	return &api.Router{
		ID:        router.ID,
		Name:      router.Name,
		NetworkID: router.GatewayInfo.NetworkID,
	}, nil

}

//GetRouter returns the router identified by id
func (client *Client) GetRouter(id string) (*api.Router, error) {
	r, err := routers.Get(client.neutron, id).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error getting Router: %s", errorString(err))
	}
	return &api.Router{
		ID:        r.ID,
		Name:      r.Name,
		NetworkID: r.GatewayInfo.NetworkID,
	}, nil

}

//ListRouter lists available routers
func (client *Client) ListRouter() ([]api.Router, error) {
	var ns []api.Router
	err := routers.List(client.neutron, routers.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		list, err := routers.ExtractRouters(page)
		if err != nil {
			return false, err
		}
		for _, r := range list {
			an := api.Router{
				ID:        r.ID,
				Name:      r.Name,
				NetworkID: r.GatewayInfo.NetworkID,
			}
			ns = append(ns, an)
		}
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error listing volume types: %s", errorString(err))
	}
	return ns, nil
}

//DeleteRouter deletes the router identified by id
func (client *Client) DeleteRouter(id string) error {
	err := routers.Delete(client.neutron, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error deleting Router: %s", errorString(err))
	}
	return nil
}
