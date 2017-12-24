package tests

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/SebastienDorgan/gpac/providers/api/IPVersion"
	"github.com/SebastienDorgan/gpac/providers/api/VMState"

	"github.com/SebastienDorgan/gpac/providers"

	"testing"
	"time"

	"github.com/SebastienDorgan/gpac/providers/api"
	"github.com/SebastienDorgan/gpac/providers/api/VolumeSpeed"
	"github.com/SebastienDorgan/gpac/providers/api/VolumeState"
	"github.com/stretchr/testify/assert"
)

//ClientTester helper class to test clients
type ClientTester struct {
	Service providers.Service
}

//ListImages test
func (tester *ClientTester) ListImages(t *testing.T) {

	images, err := tester.Service.ListImages()
	assert.Nil(t, err)
	assert.NotEmpty(t, images)
	for _, i := range images {
		assert.NotEqual(t, i.ID, "")
		assert.NotEqual(t, i.Name, "")
	}
}

//ListVMTemplates test
func (tester *ClientTester) ListVMTemplates(t *testing.T) {
	tpls, err := tester.Service.ListTemplates()
	assert.Nil(t, err)
	assert.NotEmpty(t, tpls)
	for _, f := range tpls {
		assert.NotEqual(t, f.ID, "")
		assert.NotEqual(t, f.Name, "")
		assert.NotEqual(t, f.Cores, 0)
		assert.NotEqual(t, f.RAMSize, 0)
		//	assert.NotEqual(t, f.DiskSize, 0)
	}
}

//CreateKeyPair test
func (tester *ClientTester) CreateKeyPair(t *testing.T) {
	kp, err := tester.Service.CreateKeyPair("kp")
	assert.Nil(t, err)
	defer tester.Service.DeleteKeyPair("kp")
	assert.NotEqual(t, kp.ID, "")
	assert.NotEqual(t, kp.Name, "")
	assert.NotEqual(t, kp.PrivateKey, "")
	assert.NotEqual(t, kp.PublicKey, "")

}

//GetKeyPair test
func (tester *ClientTester) GetKeyPair(t *testing.T) {
	kp, err := tester.Service.CreateKeyPair("kp")
	assert.Nil(t, err)
	defer tester.Service.DeleteKeyPair("kp")
	kp2, err := tester.Service.GetKeyPair(kp.ID)
	assert.Nil(t, err)
	assert.Equal(t, kp.ID, kp2.ID)
	assert.Equal(t, kp.Name, kp2.Name)
	assert.Equal(t, kp.PublicKey, kp2.PublicKey)
	assert.Equal(t, "", kp2.PrivateKey)
	_, err = tester.Service.GetKeyPair("notfound")
	assert.NotNil(t, err)
}

//ListKeyPairs test
func (tester *ClientTester) ListKeyPairs(t *testing.T) {
	lst, err := tester.Service.ListKeyPairs()
	assert.Nil(t, err)
	assert.Empty(t, lst)
	kp, err := tester.Service.CreateKeyPair("kp")
	assert.Nil(t, err)
	defer tester.Service.DeleteKeyPair("kp")
	kp2, err := tester.Service.CreateKeyPair("kp2")
	assert.Nil(t, err)
	defer tester.Service.DeleteKeyPair("kp2")
	lst, err = tester.Service.ListKeyPairs()
	assert.Nil(t, err)
	for _, kpe := range lst {
		var kpr api.KeyPair
		if kpe.ID == kp.ID {
			kpr = *kp
		} else if kpe.ID == kp2.ID {
			kpr = *kp2
		} else {
			t.Fail()
		}
		assert.Equal(t, kpe.ID, kpr.ID)
		assert.Equal(t, kpe.Name, kpr.Name)
		assert.Equal(t, kpe.PublicKey, kpr.PublicKey)
		assert.Equal(t, kpe.PrivateKey, "")
	}
}

//CreateNetwork creates a test network
func (tester *ClientTester) CreateNetwork(t *testing.T, name string) *api.Network {
	tpls, err := tester.Service.SelectTemplatesBySize(api.SizingRequirements{
		MinCores:    1,
		MinRAMSize:  1,
		MinDiskSize: 0,
	})
	assert.Nil(t, err)
	img, err := tester.Service.SearchImage("Ubuntu 16.04")
	assert.Nil(t, err)
	gwRequest := api.VMRequest{
		ImageID:    img.ID,
		Name:       "test_gw",
		TemplateID: tpls[0].ID,
	}
	network, err := tester.Service.CreateNetwork(api.NetworkRequest{
		Name:      name,
		IPVersion: IPVersion.IPv4,
		CIDR:      "192.168.1.0/24",
		GWRequest: gwRequest,
	})
	assert.NoError(t, err)
	return network
}

//CreateVM creates a test VM
func (tester *ClientTester) CreateVM(t *testing.T, name string, networkID string) *api.VM {
	tpls, err := tester.Service.SelectTemplatesBySize(api.SizingRequirements{
		MinCores:    1,
		MinRAMSize:  4,
		MinDiskSize: 10,
	})
	assert.Nil(t, err)
	img, err := tester.Service.SearchImage("Ubuntu 16.04")
	assert.Nil(t, err)
	vmRequest := api.VMRequest{
		ImageID:    img.ID,
		Name:       "test_gw",
		TemplateID: tpls[0].ID,
		NetworkIDs: []string{networkID},
		PublicIP:   false,
		IsGateway:  false,
	}
	vm, err := tester.Service.CreateVM(vmRequest)
	assert.Nil(t, err)
	return vm
}

//Networks test
func (tester *ClientTester) Networks(t *testing.T) {
	network1 := tester.CreateNetwork(t, "test_network_1")
	defer tester.Service.DeleteNetwork(network1.ID)

	vm, err := tester.Service.GetVM(network1.GatewayID)
	assert.Nil(t, err)
	assert.True(t, vm.AccessIPv4 != "" || vm.AccessIPv6 != "")
	assert.NotEmpty(t, vm.PrivateKey)
	assert.Empty(t, vm.GatewayID)
	fmt.Println(vm.AccessIPv4)
	fmt.Println(vm.PrivateKey)
	ssh, err := tester.Service.GetSSHConfig(vm.ID)
	assert.Nil(t, err)

	//Waits sshd deamon is up
	time.Sleep(30 * time.Second)
	cmd, err := ssh.Command("whoami")
	assert.Nil(t, err)
	out, err := cmd.Output()
	assert.Nil(t, err)
	content := strings.Trim(string(out), "\n")
	assert.Equal(t, api.DefaultUser, content)

	network2 := tester.CreateNetwork(t, "test_network_2")

	defer tester.Service.DeleteNetwork(network2.ID)

	nets, err := tester.Service.ListNetworks()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(nets))
	found := 0
	for _, n := range nets {
		if n.ID == network1.ID {
			found++
		} else if n.ID == network2.ID {
			found++
		} else {
			t.Fail()
		}
	}
	assert.Equal(t, 2, found)

	n1, err := tester.Service.GetNetwork(network1.ID)
	assert.Nil(t, err)
	assert.Equal(t, n1.CIDR, network1.CIDR)
	assert.Equal(t, n1.GatewayID, network1.GatewayID)
	assert.Equal(t, n1.ID, network1.ID)
	assert.Equal(t, n1.IPVersion, network1.IPVersion)
	assert.Equal(t, n1.Name, network1.Name)

	//network := tester.CreateNetwork(t, "test_network_1")

	// network, err := tester.Service.CreateNetwork("test_network")
	// assert.Nil(t, err)
	// defer tester.Service.DeleteNetwork(network.ID)
	// network2, err := tester.Service.GetNetwork(network.ID)
	// assert.Nil(t, err)
	// assert.Equal(t, network2.ID, network.ID)
	// assert.Equal(t, network2.Name, network2.Name)
	// assert.Empty(t, network.Subnets)
	// assert.Empty(t, network2.Subnets)
}

//VMs test
func (tester *ClientTester) VMs(t *testing.T) {
	network := tester.CreateNetwork(t, "test_network")
	defer tester.Service.DeleteNetwork(network.ID)
	vm := tester.CreateVM(t, "vm1", network.ID)
	time.Sleep(30 * time.Second)
	ssh, err := tester.Service.GetSSHConfig(vm.ID)
	cmd, err := ssh.Command("whoami")
	assert.Nil(t, err)
	out, err := cmd.Output()
	assert.Nil(t, err)
	content := strings.Trim(string(out), "\n")
	assert.Equal(t, api.DefaultUser, content)

	cmd, err = ssh.Command("ping -c1 8.8.8.8")
	fmt.Println(ssh.PrivateKey)
	assert.Nil(t, err)
	err = cmd.Run()
	assert.Nil(t, err)

	cmd, err = ssh.Command("ping -c1 www.google.fr")
	fmt.Println(ssh.PrivateKey)
	assert.Nil(t, err)
	err = cmd.Run()
	assert.Nil(t, err)

	vms, err := tester.Service.ListVMs()
	assert.Equal(t, 2, len(vms))
	found := 0
	for _, v := range vms {
		if v.ID == network.GatewayID {
			found++
		} else if v.ID == vm.ID {
			found++
		} else {
			t.Fail()
		}
	}
	assert.Equal(t, 2, found)

	v, err := tester.Service.GetVM(vm.ID)
	assert.NoError(t, err)
	assert.Equal(t, v.AccessIPv4, vm.AccessIPv4)
	assert.Equal(t, v.AccessIPv6, vm.AccessIPv6)
	assert.Equal(t, v.GatewayID, vm.GatewayID)
	assert.Equal(t, v.ID, vm.ID)
	assert.Equal(t, v.Name, vm.Name)
	assert.Equal(t, v.PrivateKey, vm.PrivateKey)
	assert.Equal(t, v.Size, vm.Size)
	assert.Equal(t, v.State, vm.State)

	for i, addr := range v.PrivateIPsV4 {
		assert.Equal(t, addr, vm.PrivateIPsV4[i])
	}

	for i, addr := range v.PrivateIPsV6 {
		assert.Equal(t, addr, vm.PrivateIPsV6[i])
	}

	tester.Service.DeleteVM(vm.ID)

}

//StartStopVM test
func (tester *ClientTester) StartStopVM(t *testing.T) {

	net := tester.CreateNetwork(t, "test_network")
	defer tester.Service.DeleteNetwork(net.ID)
	vm, err := tester.Service.GetVM(net.GatewayID)
	assert.NoError(t, err)
	{
		err := tester.Service.StopVM(vm.ID)
		assert.Nil(t, err)
		start := time.Now()
		vm, err = tester.Service.WaitVMState(vm.ID, VMState.STOPPED, 40*time.Second)
		tt := time.Now()
		fmt.Println(tt.Sub(start))
		assert.Nil(t, err)
		assert.Equal(t, vm.State, VMState.STOPPED)
	}
	{
		err := tester.Service.StartVM(vm.ID)
		assert.Nil(t, err)
		start := time.Now()
		vm, err = tester.Service.WaitVMState(vm.ID, VMState.STARTED, 40*time.Second)
		tt := time.Now()
		fmt.Println(tt.Sub(start))
		assert.Nil(t, err)
		assert.Equal(t, vm.State, VMState.STARTED)
	}

}

//Volume test
func (tester *ClientTester) Volume(t *testing.T) {
	v, err := tester.Service.CreateVolume(api.VolumeRequest{
		Name:  "test_volume",
		Size:  500,
		Speed: VolumeSpeed.HDD,
	})
	assert.Nil(t, err)
	assert.Equal(t, "test_volume", v.Name)
	assert.Equal(t, 500, v.Size)
	assert.Equal(t, VolumeSpeed.HDD, v.Speed)

	tester.Service.WaitVolumeState(v.ID, VolumeState.AVAILABLE, 40*time.Second)
	defer tester.Service.DeleteVolume(v.ID)
	v2, err := tester.Service.CreateVolume(api.VolumeRequest{
		Name:  "test_volume",
		Size:  500,
		Speed: VolumeSpeed.HDD,
	})
	assert.Nil(t, err)
	tester.Service.WaitVolumeState(v2.ID, VolumeState.AVAILABLE, 40*time.Second)
	defer tester.Service.DeleteVolume(v2.ID)
	lst, err := tester.Service.ListVolumes()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(lst))
	for _, vl := range lst {
		if vl.ID == v.ID {
			assert.Equal(t, v.Name, vl.Name)
			assert.Equal(t, v.Size, vl.Size)
			assert.Equal(t, v.Speed, vl.Speed)
		} else if vl.ID == v2.ID {
			assert.Equal(t, v2.Name, vl.Name)
			assert.Equal(t, v2.Size, vl.Size)
			assert.Equal(t, v2.Speed, vl.Speed)
		} else {
			t.Fail()
		}
	}

}

//VolumeAttachment test
func (tester *ClientTester) VolumeAttachment(t *testing.T) {
	net := tester.CreateNetwork(t, "test_network")
	defer tester.Service.DeleteNetwork(net.ID)
	vm, err := tester.Service.GetVM(net.GatewayID)
	assert.NoError(t, err)

	v, err := tester.Service.CreateVolume(api.VolumeRequest{
		Name:  "test_volume",
		Size:  500,
		Speed: VolumeSpeed.HDD,
	})
	assert.Nil(t, err)
	v2, err := tester.Service.CreateVolume(api.VolumeRequest{
		Name:  "test_volume2",
		Size:  500,
		Speed: VolumeSpeed.HDD,
	})
	assert.Nil(t, err)
	//defer clt.DeleteVolume(v.ID)
	tester.Service.WaitVolumeState(v2.ID, VolumeState.AVAILABLE, 40*time.Second)
	va, err := tester.Service.CreateVolumeAttachment(api.VolumeAttachmentRequest{
		Name:     "Attachment",
		ServerID: vm.ID,
		VolumeID: v.ID,
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, va.Device)
	va2, err := tester.Service.CreateVolumeAttachment(api.VolumeAttachmentRequest{
		Name:     "Attachment2",
		ServerID: vm.ID,
		VolumeID: v2.ID,
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, va2.Device)
	val, err := tester.Service.GetVolumeAttachment(vm.ID, v.ID)
	assert.Nil(t, err)
	assert.Equal(t, va.ID, val.ID)
	assert.Equal(t, va.Name, val.Name)
	assert.Equal(t, va.Device, val.Device)
	assert.Equal(t, va.ServerID, val.ServerID)
	assert.Equal(t, va.VolumeID, val.VolumeID)
	assert.Nil(t, err)
	lst, err := tester.Service.ListVolumeAttachments(vm.ID)
	assert.Equal(t, 2, len(lst))
	for _, val := range lst {
		if val.ID == va.ID {
			assert.Equal(t, va.ID, val.ID)
			assert.Equal(t, va.Name, val.Name)
			assert.Equal(t, va.Device, val.Device)
			assert.Equal(t, va.ServerID, val.ServerID)
			assert.Equal(t, va.VolumeID, val.VolumeID)
		} else if val.ID == va2.ID {
			assert.Equal(t, va2.ID, val.ID)
			assert.Equal(t, va2.Name, val.Name)
			assert.Equal(t, va2.Device, val.Device)
			assert.Equal(t, va2.ServerID, val.ServerID)
			assert.Equal(t, va2.VolumeID, val.VolumeID)
		} else {
			t.Fail()
		}
	}
	err = tester.Service.DeleteVolumeAttachment(vm.ID, va.ID)
	assert.Nil(t, err)
	err = tester.Service.DeleteVolumeAttachment(vm.ID, va2.ID)
	assert.Nil(t, err)
	tester.Service.DeleteVM(vm.ID)
	assert.Nil(t, err)
	tester.Service.DeleteVolume(v.ID)
	assert.Nil(t, err)
	tester.Service.DeleteVolume(v2.ID)
	assert.Nil(t, err)

}

//Containers test
func (tester *ClientTester) Containers(t *testing.T) {
	err := tester.Service.CreateContainer("testC")
	assert.Nil(t, err)
	err = tester.Service.CreateContainer("testC2")
	assert.Nil(t, err)

	cl, err := tester.Service.ListContainers()
	assert.Contains(t, cl, "testC", "testC2")
	err = tester.Service.DeleteContainer("testC")
	assert.Nil(t, err)
	err = tester.Service.DeleteContainer("testC2")
	assert.Nil(t, err)
	cl, err = tester.Service.ListContainers()
	assert.NotContains(t, cl, "testC", "testC2")
}

//Objects test
func (tester *ClientTester) Objects(t *testing.T) {
	err := tester.Service.CreateContainer("testC")
	assert.Nil(t, err)
	err = tester.Service.PutObject("testC", api.Object{
		Content:  strings.NewReader("123456789"),
		DeleteAt: time.Now().Add(5 * time.Second),
		Metadata: map[string]string{"A": "B"},
		Name:     "object1",
	})
	assert.Nil(t, err)

	o, err := tester.Service.GetObject("testC", "object1", nil)
	var buff bytes.Buffer
	_, err = buff.ReadFrom(o.Content)
	assert.Nil(t, err)
	sc := buff.String()
	assert.Equal(t, "123456789", sc)
	assert.Equal(t, 1, len(o.Metadata))
	assert.Equal(t, "B", o.Metadata["A"])

	o, err = tester.Service.GetObjectMetadata("testC", "object1")
	assert.Empty(t, o.Content)
	assert.Equal(t, 1, len(o.Metadata))
	assert.Equal(t, "B", o.Metadata["A"])
	o, err = tester.Service.GetObject("testC", "object1", []api.Range{
		api.NewRange(0, 2),
		api.NewRange(4, 7),
	})
	assert.Nil(t, err)
	if err == nil {
		buff.Reset()
		_, err = buff.ReadFrom(o.Content)
		assert.Nil(t, err)
		sc = buff.String()
		assert.Equal(t, "1235678", sc)
	}

	assert.Nil(t, err)
	time.Sleep(5 * time.Second)
	_, err = tester.Service.GetObject("testC", "object1", nil)
	assert.NotNil(t, err)

	err = tester.Service.DeleteObject("testC", "object1")
	assert.NotNil(t, err)
	err = tester.Service.DeleteContainer("testC")
	assert.Nil(t, err)
}
