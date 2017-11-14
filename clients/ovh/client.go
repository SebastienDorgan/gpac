package ovh

//ProviderNetwork name of ovh external network
const ProviderNetwork string = "Ext-Net"

/*AuthOptions fields are the union of those recognized by each identity implementation and
provider.
*/
type AuthOptions struct {
	// Endpoint ovh end point (ovh-eu, ovh-ca ...)
	Endpoint string

	ApplicationKey    string
	ApplicaionName    string
	ConsumerKey       string
	OpenstackPassword string
	//Openstack region (data center) where the infrstructure will be created
	Region string

	//Name of the provider (external) network
	ProviderNetwork string
}
