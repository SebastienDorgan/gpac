package clients

import (
	"fmt"
	"time"

	"github.com/SebastienDorgan/gpac/clients/api"
	"github.com/SebastienDorgan/gpac/clients/api/VMState"
)

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
			return nil, &api.TimeoutError{"Wait vm state timeout"}
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
