package openstack

import (
	"fmt"

	"github.com/SebastienDorgan/gpac/clients/api"
	"github.com/SebastienDorgan/gpac/clients/api/VolumeState"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/volumes"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v1/volumetypes"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/rackspace/gophercloud/pagination"
)

//toVM converts a Volume status returned by the OpenStack driver into VolumeState enum
func toVolumeState(status string) VolumeState.Enum {
	switch status {
	case "creating":
		return VolumeState.CREATING
	case "available":
		return VolumeState.AVAILABLE
	case "attaching":
		return VolumeState.ATTACHING
	case "detaching":
		return VolumeState.DETACHING
	case "in-use":
		return VolumeState.USED
	case "deleting":
		return VolumeState.DELETING
	case "error", "error_deleting", "error_backing-up", "error_restoring", "error_extending":
		return VolumeState.ERROR
	default:
		return VolumeState.OTHER
	}
}

//CreateVolume creates a block volume
//- name is the name of the volume
//- size is the size of the volume in GB
//- volumeType is the type of volume to create, if volumeType is empty the driver use a default type
func (client *Client) CreateVolume(request api.VolumeRequest) (*api.Volume, error) {
	vol, err := volumes.Create(client.blocstorage, volumes.CreateOpts{
		Name:       request.Name,
		Size:       request.Size,
		VolumeType: request.Type,
	}).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error creating volume : %s", errorString(err))
	}
	v := api.Volume{
		ID:    vol.ID,
		Name:  vol.Name,
		Size:  vol.Size,
		Type:  vol.VolumeType,
		State: toVolumeState(vol.Status),
	}
	return &v, nil
}

//GetVolume returns the volume identified by id
func (client *Client) GetVolume(id string) (*api.Volume, error) {
	vol, err := volumes.Get(client.blocstorage, id).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error getting volume: %s", errorString(err))
	}
	av := api.Volume{
		ID:    vol.ID,
		Name:  vol.Name,
		Size:  vol.Size,
		Type:  vol.VolumeType,
		State: toVolumeState(vol.Status),
	}
	return &av, nil
}

//ListVolumes list available volumes
func (client *Client) ListVolumes() ([]api.Volume, error) {
	var vs []api.Volume
	err := volumes.List(client.blocstorage, volumes.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		list, err := volumes.ExtractVolumes(page)
		if err != nil {
			return false, err
		}
		for _, vol := range list {
			av := api.Volume{
				ID:    vol.ID,
				Name:  vol.Name,
				Size:  vol.Size,
				Type:  vol.VolumeType,
				State: toVolumeState(vol.Status),
			}
			vs = append(vs, av)
		}
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error listing volume types: %s", errorString(err))
	}
	return vs, nil
}

//DeleteVolume deletes the volume identified by id
func (client *Client) DeleteVolume(id string) error {
	err := volumes.Delete(client.blocstorage, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error deleting volume: %s", errorString(err))
	}
	return nil
}

//ListVolumeTypes list volume types available
func (client *Client) ListVolumeTypes() ([]api.VolumeType, error) {
	var vtypes []api.VolumeType
	err := volumetypes.List(client.blocstorage).EachPage(func(page pagination.Page) (bool, error) {
		list, err := volumetypes.ExtractVolumeTypes(page)
		if err != nil {
			return false, err
		}
		for _, vt := range list {
			avt := api.VolumeType{
				ID:   vt.ID,
				Name: vt.Name,
			}
			vtypes = append(vtypes, avt)
		}
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error listing volume types: %s", errorString(err))
	}
	return vtypes, nil

}

//CreateVolumeAttachment attaches a volume to a VM
//- name the name of the volume attachment
//- volume the volume to attach
//- vm the VM on which the volume is attached
func (client *Client) CreateVolumeAttachment(request api.VolumeAttachmentRequest) (*api.VolumeAttachment, error) {
	va, err := volumeattach.Create(client.nova, request.ServerID, volumeattach.CreateOpts{
		VolumeID: request.VolumeID,
	}).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error creating volume attachement between server %s and volume %s: %s", request.ServerID, request.VolumeID, errorString(err))
	}

	return &api.VolumeAttachment{
		ID:       va.ID,
		ServerID: va.ServerID,
		VolumeID: va.VolumeID,
		Device:   va.Device,
	}, nil
}

//GetVolumeAttachment returns the volume attachment identified by id
func (client *Client) GetVolumeAttachment(serverID, id string) (*api.VolumeAttachment, error) {
	va, err := volumeattach.Get(client.nova, serverID, id).Extract()
	if err != nil {
		return nil, fmt.Errorf("Error getting volume attachement %s: %s", id, errorString(err))
	}
	return &api.VolumeAttachment{
		ID:       va.ID,
		ServerID: va.ServerID,
		VolumeID: va.VolumeID,
		Device:   va.Device,
	}, nil
}

//ListVolumeAttachments lists available volume attachment
func (client *Client) ListVolumeAttachments(serverID string) ([]api.VolumeAttachment, error) {
	var vs []api.VolumeAttachment
	err := volumeattach.List(client.nova, serverID).EachPage(func(page pagination.Page) (bool, error) {
		list, err := volumeattach.ExtractVolumeAttachments(page)
		if err != nil {
			return false, err
		}
		for _, va := range list {
			ava := api.VolumeAttachment{
				ID:       va.ID,
				ServerID: va.ServerID,
				VolumeID: va.VolumeID,
				Device:   va.Device,
			}
			vs = append(vs, ava)
		}
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error listing volume types: %s", errorString(err))
	}
	return vs, nil
}

//DeleteVolumeAttachment deletes the volume attachment identifed by id
func (client *Client) DeleteVolumeAttachment(serverID, id string) error {
	err := volumeattach.Delete(client.nova, serverID, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("Error deleting volume attachement %s: %s", id, errorString(err))
	}
	return nil
}
