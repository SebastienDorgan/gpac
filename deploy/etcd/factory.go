package etcd

import (
	"github.com/GeertJohan/go.rice"
	"github.com/SebastienDorgan/gpac/system"
)

//go:generate rice embed-go

//EctdFactory is used to create etcd HA cluster
type EctdFactory struct {
	LeaderSSH  *system.SSHConfig
	LeaderIP   string
	scriptsBox *rice.Box
}

//Create an EctdFactory
func Create(leaderSSH *system.SSHConfig, leaderIP string) (*EctdFactory, error) {
	box, err := rice.FindBox("scripts")
	if err != nil {
		return nil, err
	}
	return &EctdFactory{
		LeaderSSH:  leaderSSH,
		LeaderIP:   leaderIP,
		scriptsBox: box,
	}, nil
}

//Initialize install discovery service and etcd
func (leader *EctdFactory) Initialize() error {
	return nil
}

//AddMaster add a master node to etcd cluster
func (leader *EctdFactory) AddMaster(masterSSH *system.SSHConfig, masterIP string) error {
	return nil
}

//AddWorker add a work node to etcd cluster
func (leader *EctdFactory) AddWorker(workerSSH *system.SSHConfig, masterIP string) error {
	return nil
}
