package clients

import (
	"fmt"
	"sort"
	"time"

	"github.com/satori/go.uuid"

	"github.com/xrash/smetrics"

	"github.com/SebastienDorgan/gpac/clients/api"
	"github.com/SebastienDorgan/gpac/clients/api/IPVersion"
	"github.com/SebastienDorgan/gpac/clients/api/VMState"
)

const (
	//CoreDRFWeight is the Dominant Resource Fairness weight of a core
	CoreDRFWeight float32 = 1.0
	//RAMDRFWeight is the Dominant Resource Fairness weight of 1 GB of RAM
	RAMDRFWeight float32 = 1.0 / 8.0
	//DiskDRFWeight is the Dominant Resource Fairness weight of 1 GB of Disk
	DiskDRFWeight float32 = 1.0 / 16.0
)

//RankDRF computes the Dominant Resource Fairness Rank of a VM template
func RankDRF(t *api.VMTemplate) float32 {
	fc := float32(t.Cores)
	fr := t.RAMSize
	fd := float32(t.DiskSize)
	return fc*CoreDRFWeight + fr*RAMDRFWeight + fd*DiskDRFWeight
}

// ByRankDRF implements sort.Interface for []VMTemplate based on
// the Dominant Resource Fairness
type ByRankDRF []api.VMTemplate

func (a ByRankDRF) Len() int           { return len(a) }
func (a ByRankDRF) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRankDRF) Less(i, j int) bool { return RankDRF(&a[i]) < RankDRF(&a[j]) }

//ServerAccess a VM and the SSH Key Pair
type ServerAccess struct {
	VM      *api.VM
	Key     *api.KeyPair
	Gateway *ServerAccess
}

//ServerRequest used to create a server
type ServerRequest struct {
	Name string `json:"name,omitempty"`
	//NetworksIDs list of the network IDs the VM must be connected
	Networks []api.Network `json:"networks,omitempty"`
	//PublicIP a flg telling if the VM must have a public IP is
	PublicIP bool `json:"public_ip,omitempty"`
	//TemplateID the UUID of the template used to size the VM (see SelectTemplates)
	Template api.VMTemplate `json:"sizing,omitempty"`
	//ImageID  is the UUID of the image that contains the server's OS and initial state.
	OSName string `json:"os_name,omitempty"`
	//Gateway through which the server can be connected
	Gateway *ServerAccess
}

//WaitVMState waits a vm achieve state
func WaitVMState(client api.ClientAPI, vmID string, state VMState.Enum, timeout time.Duration) (*api.VM, error) {
	cout := make(chan int)
	next := make(chan bool)
	vmc := make(chan *api.VM)
	go poll(client, vmID, state, cout, next, vmc)
	for {
		select {
		case res := <-cout:
			if res == 0 {
				//next <- false
				return nil, fmt.Errorf("Error getting vm state")
			}
			if res == 1 {
				//next <- false
				fmt.Println("Found Found")
				return <-vmc, nil
			}
			if res == 2 {
				fmt.Println("Continue")
				next <- true
			}
		case <-time.After(time.Second * timeout):
			next <- false
			return nil, &api.TimeoutError{Message: "Wait vm state timeout"}
		}
	}
}

func poll(client api.ClientAPI, vmID string, state VMState.Enum, cout chan int, next chan bool, vmc chan *api.VM) {
	for {
		vm, err := client.GetVM(vmID)
		fmt.Println(vm.State.String())
		if err != nil {
			cout <- 0
			return
		}
		if vm.State == state {
			cout <- 1
			vmc <- vm
			fmt.Println("Found")
			return
		}
		cout <- 2
		if !<-next {
			return
		}
	}
}

//NetworkRequest defines a request to create a network
type NetworkRequest struct {
	Name string `json:"name,omitempty"`
	//IPVersion must be IPv4 or IPv6 (see IPVersion)
	IPVersion IPVersion.Enum `json:"ip_version,omitempty"`
	//Mask mask in CIDR notation
	Mask string `json:"mask,omitempty"`
	//NetworkID id of the parent network
	NetworkID string `json:"network_id,omitempty"`
}

//CreateNetwork create a network with a default subnet
func CreateNetwork(client api.ClientAPI, request NetworkRequest) (*api.Network, error) {
	n, err := client.CreateNetwork(request.Name)
	if err != nil {
		return nil, err
	}
	_, err = client.CreateSubnet(api.SubnetRequets{
		Name:      n.Name,
		NetworkID: n.ID,
		IPVersion: request.IPVersion,
		Mask:      request.Mask,
	})
	if err != nil {
		defer client.DeleteNetwork(n.ID)
		return nil, err
	}
	return n, nil
}

//DeleteNetwork delete a network and the subnet associated
func DeleteNetwork(client api.ClientAPI, networkID string) error {
	net, err := client.GetNetwork(networkID)
	if err != nil {
		return err
	}
	subnets, err := client.ListSubnets(networkID)
	if err != nil {
		return err
	}
	for _, sn := range subnets {
		client.DeleteSubnet(sn.ID)
	}

	return client.DeleteNetwork(net.ID)
}

//SelectTemplatesBySize select templates satisfying sizing requirements
//returned list is ordered by size fitting
func SelectTemplatesBySize(client api.ClientAPI, sizing api.SizingRequirements) ([]api.VMTemplate, error) {
	tpls, err := client.ListTemplates()
	var selectedTpls []api.VMTemplate
	if err != nil {
		return nil, err
	}
	for _, tpl := range tpls {
		if tpl.Cores >= sizing.MinCores && tpl.DiskSize >= sizing.MinDiskSize && tpl.RAMSize >= sizing.MinRAMSize {
			selectedTpls = append(selectedTpls, tpl)
		}
	}
	sort.Sort(ByRankDRF(selectedTpls))
	return selectedTpls, nil
}

//CreateServer creates a sever fitting request
func CreateServer(client api.ClientAPI, request ServerRequest) (*ServerAccess, error) {
	imgs, err := client.ListImages()
	if err != nil {
		return nil, err
	}
	maxscore := 0.0
	maxi := 0
	for i, img := range imgs {
		score := smetrics.JaroWinkler(img.Name, request.OSName, 0.7, 4)
		if score > maxscore {
			maxscore = score
			maxi = i
		}
	}
	if maxscore < 0.8 {
		return nil, fmt.Errorf("Unable to found and image matching %s", request.OSName)
	}
	kpName := uuid.NewV4().String()
	kp, err := client.CreateKeyPair(kpName)
	if err != nil {
		return nil, fmt.Errorf("Error creating key pair")
	}
	defer client.DeleteKeyPair(kpName)
	var netIds []string
	for _, n := range request.Networks {
		netIds = append(netIds, n.ID)

	}
	vmReq := api.VMRequest{
		Name:       request.Name,
		ImageID:    imgs[maxi].ID,
		KeyPairID:  kp.ID,
		PublicIP:   request.PublicIP,
		NetworkIDs: netIds,
		TemplateID: request.Template.ID,
	}
	vm, err := client.CreateVM(vmReq)
	if err != nil {
		return nil, err
	}
	return &ServerAccess{
		VM:      vm,
		Key:     kp,
		Gateway: request.Gateway,
	}, nil
}