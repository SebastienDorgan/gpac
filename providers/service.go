package providers

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/SebastienDorgan/gpac/providers/api"
	"github.com/SebastienDorgan/gpac/providers/api/VMState"
	"github.com/SebastienDorgan/gpac/providers/api/VolumeState"
	uuid "github.com/satori/go.uuid"

	"github.com/xrash/smetrics"
)

//ResourceError resource error
type ResourceError struct {
	Name         string
	ResourceType string
}

//ResourceNotFound resource not found error
type ResourceNotFound struct {
	ResourceError
}

//ResourceNotFoundError creates a ResourceNotFound error
func ResourceNotFoundError(resource string, name string) ResourceNotFound {
	return ResourceNotFound{
		ResourceError{
			Name:         name,
			ResourceType: resource,
		},
	}
}
func (e ResourceNotFound) Error() string {
	return fmt.Sprintf("Unable to find %s %s", e.ResourceType, e.Name)
}

//ResourceAlreadyExists resource already exists error
type ResourceAlreadyExists struct {
	ResourceError
}

//ResourceAlreadyExistsError creates a ResourceAlreadyExists error
func ResourceAlreadyExistsError(resource string, name string) ResourceAlreadyExists {
	return ResourceAlreadyExists{
		ResourceError{
			Name:         name,
			ResourceType: resource,
		},
	}
}

func (e ResourceAlreadyExists) Error() string {
	return fmt.Sprintf("%s %s alredy exists", e.ResourceType, e.Name)
}

//Service Client High level service
type Service struct {
	api.ClientAPI
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

//VMAccess a VM and the SSH Key Pair
type VMAccess struct {
	VM      *api.VM
	Key     *api.KeyPair
	User    string
	Gateway *VMAccess
}

//GetAccessIP returns the access IP
func (access *VMAccess) GetAccessIP() string {
	ip := access.VM.AccessIPv4
	if len(ip) == 0 {
		ip = access.VM.AccessIPv6
	}
	return ip
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
	Gateway *VMAccess `json:"gateway,omitempty"`
}

//WaitVMState waits a vm achieve state
func (srv *Service) WaitVMState(vmID string, state VMState.Enum, timeout time.Duration) (*api.VM, error) {
	cout := make(chan int)
	next := make(chan bool)
	vmc := make(chan *api.VM)

	go pollVM(srv, vmID, state, cout, next, vmc)
	for {
		select {
		case res := <-cout:
			if res == 0 {
				//next <- false
				return nil, fmt.Errorf("Error getting vm state")
			}
			if res == 1 {
				//next <- false
				return <-vmc, nil
			}
			if res == 2 {
				next <- true
			}
		case <-time.After(timeout):
			next <- false
			return nil, &api.TimeoutError{Message: "Wait vm state timeout"}
		}
	}
}

func pollVM(client api.ClientAPI, vmID string, state VMState.Enum, cout chan int, next chan bool, vmc chan *api.VM) {
	for {

		vm, err := client.GetVM(vmID)
		if err != nil {

			cout <- 0
			return
		}
		if vm.State == state {
			cout <- 1
			vmc <- vm
			return
		}
		cout <- 2
		if !<-next {
			return
		}
	}
}

//WaitVolumeState waits a vm achieve state
func (srv *Service) WaitVolumeState(volumeID string, state VolumeState.Enum, timeout time.Duration) (*api.Volume, error) {
	cout := make(chan int)
	next := make(chan bool)
	vc := make(chan *api.Volume)

	go pollVolume(srv, volumeID, state, cout, next, vc)
	for {
		select {
		case res := <-cout:
			if res == 0 {
				//next <- false
				return nil, fmt.Errorf("Error getting vm state")
			}
			if res == 1 {
				//next <- false
				return <-vc, nil
			}
			if res == 2 {
				next <- true
			}
		case <-time.After(timeout):
			next <- false
			return nil, &api.TimeoutError{Message: "Wait vm state timeout"}
		}
	}
}

func pollVolume(client api.ClientAPI, volumeID string, state VolumeState.Enum, cout chan int, next chan bool, vmc chan *api.Volume) {
	for {

		v, err := client.GetVolume(volumeID)
		if err != nil {

			cout <- 0
			return
		}
		if v.State == state {
			cout <- 1
			vmc <- v
			return
		}
		cout <- 2
		if !<-next {
			return
		}
	}
}

//SelectTemplatesBySize select templates satisfying sizing requirements
//returned list is ordered by size fitting
func (srv *Service) SelectTemplatesBySize(sizing api.SizingRequirements) ([]api.VMTemplate, error) {
	tpls, err := srv.ListTemplates()
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

//SearchImage search an image corresponding to OS Name
//use the Jaro Winkler algorithm to match os name and image name
func (srv *Service) SearchImage(osname string) (*api.Image, error) {
	imgs, err := srv.ListImages()
	if err != nil {
		return nil, err
	}
	maxscore := 0.0
	maxi := 0
	for i, img := range imgs {
		score := smetrics.JaroWinkler(strings.ToUpper(img.Name), strings.ToUpper(osname), 0.7, 4)
		if score > maxscore {
			maxscore = score
			maxi = i
		}
	}
	if maxscore < 0.8 {
		return nil, fmt.Errorf("Unable to found and image matching %s", osname)
	}
	return &imgs[maxi], nil
}

//GetNetworkByName returns the network named name
func (srv *Service) GetNetworkByName(name string) (*api.Network, error) {
	nets, err := srv.ListNetworksByName()
	if err != nil {
		return nil, err
	}
	n, ok := nets[name]
	if !ok {
		return nil, ResourceNotFoundError("Network", name)
	}
	return &n, nil

}

//ListNetworksByName returns network list
func (srv *Service) ListNetworksByName() (map[string]api.Network, error) {
	nets, err := srv.ListNetworks()
	if err != nil {
		return nil, err
	}
	netMap := make(map[string]api.Network)
	for _, n := range nets {
		netMap[n.Name] = n
	}
	return netMap, nil

}

//CreateVMWithKeyPair creates a VM
func (srv *Service) CreateVMWithKeyPair(request api.VMRequest) (*api.VM, *api.KeyPair, error) {
	_, err := srv.GetVMByName(request.Name)
	if err == nil {
		return nil, nil, ResourceAlreadyExistsError("VM", request.Name)
	}

	//Create temporary key pair
	kpName := uuid.NewV4().String()
	kp, err := srv.CreateKeyPair(kpName)
	if err != nil {
		return nil, nil, err
	}
	defer srv.DeleteKeyPair(kpName)

	//Create VM
	vmReq := api.VMRequest{
		Name:       request.Name,
		ImageID:    request.ImageID,
		KeyPair:    kp,
		PublicIP:   request.PublicIP,
		NetworkIDs: request.NetworkIDs,
		TemplateID: request.TemplateID,
	}
	vm, err := srv.CreateVM(vmReq)
	if err != nil {
		return nil, nil, err
	}
	return vm, kp, nil
}

//ListVMsByName list VMs by name
func (srv *Service) ListVMsByName() (map[string]api.VM, error) {
	vms, err := srv.ListVMs()
	if err != nil {
		return nil, err
	}
	vmMap := make(map[string]api.VM)
	for _, vm := range vms {
		vmMap[vm.Name] = vm
	}
	return vmMap, nil
}

//GetVMByName returns VM corresponding to name
func (srv *Service) GetVMByName(name string) (*api.VM, error) {
	vms, err := srv.ListVMsByName()
	if err != nil {
		return nil, err
	}
	vm, ok := vms[name]
	if !ok {
		return nil, ResourceNotFoundError("VM", name)
	}
	return &vm, nil
}
