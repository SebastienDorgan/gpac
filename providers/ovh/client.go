package ovh

import (
	"github.com/SebastienDorgan/gpac/providers/openstack"

	"github.com/SebastienDorgan/gpac/providers/api/VolumeSpeed"

	"github.com/ovh/go-ovh/ovh"
)

//ProviderNetwork name of ovh external network
const ProviderNetwork string = "Ext-Net"

/*AuthOptions fields are the union of those recognized by each identity implementation and
provider.
*/
type AuthOptions struct {
	// Endpoint ovh end point (ovh-eu, ovh-ca ...)
	Endpoint string
	//Application or Project Name
	ApplicationName string
	//Application Key or project ID
	ApplicationKey string
	//Consumer key
	ConsumerKey string
	//Openstack identifier
	OpenstackID string
	//OpenStack password
	OpenstackPassword string
	//Name of the data center (GRA3, BHS3 ...)
	Region string
}

// func parseOpenRC(openrc string) (*openstack.AuthOptions, error) {
// 	tokens := strings.Split(openrc, "export")
// }

//AuthenticatedClient returns an authenticated client
func AuthenticatedClient(opts AuthOptions) (*Client, error) {
	client := &Client{}
	c, err := ovh.NewClient(opts.Endpoint, opts.ApplicationName, opts.ApplicationKey, opts.ConsumerKey)
	if err != nil {
		return nil, err
	}
	client.ovh = c
	os, err := openstack.AuthenticatedClient(openstack.AuthOptions{
		IdentityEndpoint: "https://auth.cloud.ovh.net/v2.0",
		//UserID:           opts.OpenstackID,
		Username: opts.OpenstackID,
		Password: opts.OpenstackPassword,
		TenantID: opts.ApplicationKey,
		Region:   opts.Region,
	},
		openstack.CfgOptions{
			ProviderNetwork:         "Ext-Net",
			UseFloatingIP:           false,
			UseLayer3Networking:     false,
			AutoVMNetworkInterfaces: false,
			DNSList:                 []string{"8.8.8.8", "8.8.4.4"},
			VolumeSpeeds: map[string]VolumeSpeed.Enum{
				"classic":    VolumeSpeed.COLD,
				"high-speed": VolumeSpeed.HDD,
			},
		},
	)

	if err != nil {
		return nil, err
	}
	client.Client = os
	return client, nil

}

//Client is the implementation of the ovh driver regarding to the api.ClientAPI
//This client used ovh api and opensatck ovh api to maximize code reuse
type Client struct {
	*openstack.Client
	opts AuthOptions
	ovh  *ovh.Client
}
