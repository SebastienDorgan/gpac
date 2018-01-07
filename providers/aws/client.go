package aws

//go:generate rice embed-go
import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/SebastienDorgan/gpac/providers"
	"github.com/SebastienDorgan/gpac/providers/api/VolumeState"

	"github.com/SebastienDorgan/gpac/providers/api/VolumeSpeed"

	rice "github.com/GeertJohan/go.rice"
	"github.com/SebastienDorgan/gpac/providers/api/VMState"

	"github.com/SebastienDorgan/gpac/providers/api"
	"github.com/SebastienDorgan/gpac/system"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/s3"
)

//Config AWS configurations
type Config struct {
	ImageOwners    []string
	DefaultNetwork string
}

//AuthOpts AWS credentials
type AuthOpts struct {
	// AWS Access key ID
	AccessKeyID string

	// AWS Secret Access Key
	SecretAccessKey string
	// The region to send requests to. This parameter is required and must
	// be configured globally or on a per-client basis unless otherwise
	// noted. A full list of regions is found in the "Regions and Endpoints"
	// document.
	//
	// @see http://docs.aws.amazon.com/general/latest/gr/rande.html
	//   AWS Regions and Endpoints
	Region string
	Config *Config
}

// Retrieve returns nil if it successfully retrieved the value.
// Error is returned if the value were not obtainable, or empty.
func (o AuthOpts) Retrieve() (credentials.Value, error) {
	return credentials.Value{
		AccessKeyID:     o.AccessKeyID,
		SecretAccessKey: o.SecretAccessKey,
		ProviderName:    "internal",
	}, nil
}

// IsExpired returns if the credentials are no longer valid, and need
// to be retrieved.
func (o AuthOpts) IsExpired() bool {
	return false
}

//AuthenticatedClient returns an authenticated client
func AuthenticatedClient(opts AuthOpts) (*Client, error) {
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(opts.Region),
		Credentials: credentials.NewCredentials(opts),
	})
	if err != nil {
		return nil, err
	}
	sPricing, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewCredentials(opts),
	})
	if err != nil {
		return nil, err
	}
	box, err := rice.FindBox("scripts")
	if err != nil {
		return nil, err
	}
	userDataStr, err := box.String("userdata.sh")
	if err != nil {
		return nil, err
	}
	tpl, err := template.New("user_data").Parse(userDataStr)
	if err != nil {
		return nil, err
	}
	c := Client{
		Session:     s,
		EC2:         ec2.New(s),
		Pricing:     pricing.New(sPricing),
		AuthOpts:    opts,
		UserDataTpl: tpl,
	}
	c.CreateContainer("gpac.aws.networks")
	c.CreateContainer("gpac.aws.wms")
	c.CreateContainer("gpac.aws.volumes")

	return &c, nil
}

func wrapError(msg string, err error) error {
	if err == nil {
		return nil
	}
	if aerr, ok := err.(awserr.Error); ok {
		return fmt.Errorf("%s: cause by %s", msg, aerr.Message())
	}
	return err
}

//Client a AWS provider client
type Client struct {
	Session     *session.Session
	EC2         *ec2.EC2
	Pricing     *pricing.Pricing
	AuthOpts    AuthOpts
	UserDataTpl *template.Template
	ImageOwners []string
}

func createFilters() []*ec2.Filter {
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("state"),
			Values: []*string{aws.String("available")},
		},
		&ec2.Filter{
			Name:   aws.String("architecture"),
			Values: []*string{aws.String("x86_64")},
		},
		&ec2.Filter{
			Name:   aws.String("virtualization-type"),
			Values: []*string{aws.String("hvm")},
		},
		&ec2.Filter{
			Name:   aws.String("root-device-type"),
			Values: []*string{aws.String("ebs")},
		},
	}
	// Ubuntu 099720109477
	// Fedora 013116697141
	// Debian 379101102735
	// CentOS 057448758665
	// CoreOS 595879546273
	// Gentoo 341857463381
	owners := []*string{
		aws.String("099720109477"),
		aws.String("013116697141"),
		aws.String("379101102735"),
		aws.String("057448758665"),
		aws.String("595879546273"),
		aws.String("902460189751"),
	}
	filters = append(filters, &ec2.Filter{
		Name:   aws.String("owner-id"),
		Values: owners,
	})
	return filters
}

//ListImages lists available OS images
func (c *Client) ListImages() ([]api.Image, error) {

	images, err := c.EC2.DescribeImages(&ec2.DescribeImagesInput{
		//Owners: []*string{aws.String("aws-marketplace"), aws.String("self")},
		Filters: createFilters(),
	})
	if err != nil {
		return nil, err
	}
	var list []api.Image
	for _, img := range images.Images {
		if img.Description == nil || strings.Contains(strings.ToUpper(*img.Name), "TEST") {
			continue
		}
		list = append(list, api.Image{
			ID:   *img.ImageId,
			Name: *img.Name,
		})
	}

	return list, nil
}

//Attributes attributes of a compute instance
type Attributes struct {
	ClockSpeed                  string `json:"clockSpeed,omitempty"`
	CurrentGeneration           string `json:"currentGeneration,omitempty"`
	DedicatedEbsThroughput      string `json:"dedicatedEbsThroughput,omitempty"`
	Ecu                         string `json:"ecu,omitempty"`
	EnhancedNetworkingSupported string `json:"enhancedNetworkingSupported,omitempty"`
	InstanceFamily              string `json:"instanceFamily,omitempty"`
	InstanceType                string `json:"instanceType,omitempty"`
	LicenseModel                string `json:"licenseModel,omitempty"`
	Location                    string `json:"location,omitempty"`
	LocationType                string `json:"locationType,omitempty"`
	Memory                      string `json:"memory,omitempty"`
	MetworkPerformance          string `json:"metworkPerformance,omitempty"`
	NormalizationSizeFactor     string `json:"normalizationSizeFactor,omitempty"`
	OperatingSystem             string `json:"operatingSystem,omitempty"`
	Operation                   string `json:"operation,omitempty"`
	PhysicalProcessor           string `json:"physicalProcessor,omitempty"`
	PreInstalledSw              string `json:"preInstalled_sw,omitempty"`
	ProcessorArchitecture       string `json:"processorArchitecture,omitempty"`
	ProcessorFeatures           string `json:"processorFeatures,omitempty"`
	Servicecode                 string `json:"servicecode,omitempty"`
	Servicename                 string `json:"servicename,omitempty"`
	Storage                     string `json:"storage,omitempty"`
	Tenancy                     string `json:"tenancy,omitempty"`
	Usagetype                   string `json:"usagetype,omitempty"`
	Vcpu                        string `json:"vcpu,omitempty"`
}

//Product compute instance product
type Product struct {
	Attributes    Attributes `json:"attributes,omitempty"`
	ProductFamily string     `json:"productFamily,omitempty"`
	Sku           string     `json:"sku,omitempty"`
}

//PriceDimension compute instance price related to term condition
type PriceDimension struct {
	AppliesTo    []string           `json:"appliesTo,omitempty"`
	BeginRange   string             `json:"beginRange,omitempty"`
	Description  string             `json:"description,omitempty"`
	EndRange     string             `json:"endRange,omitempty"`
	PricePerUnit map[string]float32 `json:"pricePerUnit,omitempty"`
	RateCode     string             `json:"RateCode,omitempty"`
	Unit         string             `json:"Unit,omitempty"`
}

//PriceDimensions compute instance price dimensions
type PriceDimensions struct {
	PriceDimensionMap map[string]PriceDimension `json:"price_dimension_map,omitempty"`
}

//TermAttributes compute instance terms
type TermAttributes struct {
	LeaseContractLength string `json:"leaseContractLength,omitempty"`
	OfferingClass       string `json:"offeringClass,omitempty"`
	PurchaseOption      string `json:"purchaseOption,omitempty"`
}

//Card compute instance price card
type Card struct {
	EffectiveDate   string          `json:"effectiveDate,omitempty"`
	OfferTermCode   string          `json:"offerTermCode,omitempty"`
	PriceDimensions PriceDimensions `json:"priceDimensions,omitempty"`
	Sku             string          `json:"sku,omitempty"`
	TermAttributes  TermAttributes  `json:"termAttributes,omitempty"`
}

//OnDemand on demand compute instance cards
type OnDemand struct {
	Cards map[string]Card
}

//Reserved reserved compute instance cards
type Reserved struct {
	Cards map[string]Card `json:"cards,omitempty"`
}

//Terms compute instance prices terms
type Terms struct {
	OnDemand OnDemand `json:"onDemand,omitempty"`
	Reserved Reserved `json:"reserved,omitempty"`
}

//Price Compute instance price information
type Price struct {
	Product         Product `json:"product,omitempty"`
	PublicationDate string  `json:"publicationDate,omitempty"`
	ServiceCode     string  `json:"serviceCode,omitempty"`
	Terms           Terms   `json:"terms,omitempty"`
}

//GetImage returns the Image referenced by id
func (c *Client) GetImage(id string) (*api.Image, error) {
	images, err := c.EC2.DescribeImages(&ec2.DescribeImagesInput{
		ImageIds: []*string{aws.String(id)},
	})
	if err != nil {
		return nil, err
	}
	if len(images.Images) == 0 {
		return nil, fmt.Errorf("Image %s does not exists", id)
	}
	img := images.Images[0]
	return &api.Image{
		ID:   *img.ImageId,
		Name: *img.Name,
	}, nil
}

//GetTemplate returns the Template referenced by id
func (c *Client) GetTemplate(id string) (*api.VMTemplate, error) {
	input := pricing.GetProductsInput{
		Filters: []*pricing.Filter{
			{
				Field: aws.String("ServiceCode"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("AmazonEC2"),
			},
			{
				Field: aws.String("location"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("US East (Ohio)"),
			},

			{
				Field: aws.String("preInstalledSw"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("NA"),
			},
			{
				Field: aws.String("operatingSystem"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("Linux"),
			},

			{
				Field: aws.String("instanceType"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String(id),
			},
		},
		FormatVersion: aws.String("aws_v1"),
		MaxResults:    aws.Int64(100),
		ServiceCode:   aws.String("AmazonEC2"),
	}

	p, err := c.Pricing.GetProducts(&input)
	if err != nil {
		return nil, err
	}
	for _, price := range p.PriceList {
		jsonPrice, err := json.Marshal(price)
		if err != nil {
			continue
		}
		price := Price{}
		err = json.Unmarshal(jsonPrice, &price)
		if err != nil {
			continue
		}
		if strings.Contains(price.Product.Attributes.Usagetype, "USE2-BoxUsage:") {
			cores, err := strconv.Atoi(price.Product.Attributes.Vcpu)
			if err != nil {
				continue
			}

			tpl := api.VMTemplate{
				ID:   price.Product.Attributes.InstanceType,
				Name: price.Product.Attributes.InstanceType,
				VMSize: api.VMSize{
					Cores:    cores,
					DiskSize: int(parseStorage(price.Product.Attributes.Storage)),
					RAMSize:  float32(parseMemory(price.Product.Attributes.Memory)),
				},
			}
			return &tpl, nil
		}
	}
	return nil, fmt.Errorf("Unable to find template %s", id)

}

func parseStorage(str string) float64 {
	r, _ := regexp.Compile("([0-9]*) x ([0-9]*(\\.|,)?[0-9]*) ?([a-z A-Z]*)?")
	b := bytes.Buffer{}
	b.WriteString(str)
	tokens := r.FindAllStringSubmatch(str, -1)
	if len(tokens) <= 0 || len(tokens[0]) <= 1 {
		return 0.0
	}
	factor, err := strconv.ParseFloat(tokens[0][1], 64)
	if err != nil {
		return 0.0
	}
	sizeStr := strings.Replace(tokens[0][2], ",", "", -1)
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0.0
	}
	if size < 10 {
		size = size * 1000
	}
	//	fmt.Println((factor * size))
	return factor * size
}
func parseMemory(str string) float64 {
	r, err := regexp.Compile("([0-9]*(\\.|,)?[0-9]*) ?([a-z A-Z]*)?")
	if err != nil {
		return 0.0
	}
	b := bytes.Buffer{}
	b.WriteString(str)
	tokens := r.FindAllStringSubmatch(str, -1)
	sizeStr := strings.Replace(tokens[0][1], ",", "", -1)
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0.0
	}

	//	fmt.Println((factor * size))
	return size
}

//ListTemplates lists available VM templates
//VM templates are sorted using Dominant Resource Fairness Algorithm
func (c *Client) ListTemplates() ([]api.VMTemplate, error) {
	input := pricing.GetProductsInput{
		Filters: []*pricing.Filter{
			{
				Field: aws.String("ServiceCode"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("AmazonEC2"),
			},
			{
				Field: aws.String("location"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("US East (Ohio)"),
			},
			{
				Field: aws.String("preInstalledSw"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("NA"),
			},
			{
				Field: aws.String("operatingSystem"),
				Type:  aws.String("TERM_MATCH"),
				Value: aws.String("Linux"),
			},
		},
		FormatVersion: aws.String("aws_v1"),
		MaxResults:    aws.Int64(100),
		ServiceCode:   aws.String("AmazonEC2"),
	}
	tpls := []api.VMTemplate{}
	//prices := map[string]interface{}{}
	err := c.Pricing.GetProductsPages(&input,
		func(p *pricing.GetProductsOutput, lastPage bool) bool {

			for _, price := range p.PriceList {
				jsonPrice, err := json.Marshal(price)
				if err != nil {
					continue
				}
				price := Price{}
				err = json.Unmarshal(jsonPrice, &price)
				if err != nil {
					continue
				}
				if strings.Contains(price.Product.Attributes.Usagetype, "USE2-BoxUsage:") {
					cores, err := strconv.Atoi(price.Product.Attributes.Vcpu)
					if err != nil {
						continue
					}

					tpl := api.VMTemplate{
						ID:   price.Product.Attributes.InstanceType,
						Name: price.Product.Attributes.InstanceType,
						VMSize: api.VMSize{
							Cores:    cores,
							DiskSize: int(parseStorage(price.Product.Attributes.Storage)),
							RAMSize:  float32(parseMemory(price.Product.Attributes.Memory)),
						},
					}
					tpls = append(tpls, tpl)
				}
			}
			return lastPage
		})
	if err != nil {
		return nil, err
	}
	return tpls, nil
}

//CreateKeyPair creates and import a key pair
func (c *Client) CreateKeyPair(name string) (*api.KeyPair, error) {
	publicKey, privateKey, err := system.CreateKeyPair()
	if err != nil {
		return nil, err
	}
	c.EC2.ImportKeyPair(&ec2.ImportKeyPairInput{
		KeyName:           aws.String(name),
		PublicKeyMaterial: publicKey,
	})
	// out, err := c.EC2.CreateKeyPair(&ec2.CreateKeyPairInput{
	// 	KeyName: aws.String(name),
	// })
	if err != nil {
		return nil, err
	}
	return &api.KeyPair{
		ID:         name,
		Name:       name,
		PrivateKey: string(privateKey),
		PublicKey:  string(publicKey),
	}, nil
}

//GetKeyPair returns the key pair identified by id
func (c *Client) GetKeyPair(id string) (*api.KeyPair, error) {
	out, err := c.EC2.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{
		KeyNames: []*string{aws.String(id)},
	})
	if err != nil {
		return nil, err
	}
	kp := out.KeyPairs[0]
	return &api.KeyPair{
		ID:         pStr(kp.KeyName),
		Name:       pStr(kp.KeyName),
		PrivateKey: "",
		PublicKey:  pStr(kp.KeyFingerprint),
	}, nil
}

//ListKeyPairs lists available key pairs
func (c *Client) ListKeyPairs() ([]api.KeyPair, error) {
	out, err := c.EC2.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
	if err != nil {
		return nil, err
	}
	keys := []api.KeyPair{}
	for _, kp := range out.KeyPairs {
		keys = append(keys, api.KeyPair{
			ID:         pStr(kp.KeyName),
			Name:       pStr(kp.KeyName),
			PrivateKey: "",
			PublicKey:  pStr(kp.KeyFingerprint),
		})

	}
	return keys, nil
}

//DeleteKeyPair deletes the key pair identified by id
func (c *Client) DeleteKeyPair(id string) error {
	_, err := c.EC2.DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: aws.String(id),
	})
	return err
}

func (c *Client) saveNetwork(n api.Network) error {
	b, err := json.Marshal(n)
	if err != nil {
		return err
	}
	buffer := bytes.NewReader(b)
	return c.PutObject("gpac.aws.networks", api.Object{
		Name:    n.ID,
		Content: buffer,
	})
}

func (c *Client) getNetwork(netID string) (*api.Network, error) {
	o, err := c.GetObject("gpac.aws.networks", netID, nil)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	buffer.ReadFrom(o.Content)
	net := api.Network{}
	err = json.Unmarshal(buffer.Bytes(), &net)
	if err != nil {
		return nil, err
	}
	return &net, err
}
func (c *Client) removeNetwork(netID string) error {
	return c.DeleteObject("gpac.aws.networks", netID)
}

//CreateNetwork creates a network named name
func (c *Client) CreateNetwork(req api.NetworkRequest) (*api.Network, error) {
	vpcOut, err := c.EC2.CreateVpc(&ec2.CreateVpcInput{
		CidrBlock: aws.String(req.CIDR),
	})
	if err != nil {
		return nil, err
	}
	sn, err := c.EC2.CreateSubnet(&ec2.CreateSubnetInput{
		CidrBlock: aws.String(req.CIDR),
		VpcId:     vpcOut.Vpc.VpcId,
	})
	if err != nil {
		c.DeleteNetwork(*vpcOut.Vpc.VpcId)
		return nil, err
	}
	gw, err := c.EC2.CreateInternetGateway(&ec2.CreateInternetGatewayInput{})
	if err != nil {
		c.DeleteNetwork(*vpcOut.Vpc.VpcId)
		return nil, err
	}
	_, err = c.EC2.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		VpcId:             vpcOut.Vpc.VpcId,
		InternetGatewayId: gw.InternetGateway.InternetGatewayId,
	})
	if err != nil {
		c.DeleteNetwork(*vpcOut.Vpc.VpcId)
		return nil, err
	}
	table, err := c.EC2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String("vpc-id"),
				Values: []*string{
					vpcOut.Vpc.VpcId,
				},
			},
		},
	})
	if err != nil || len(table.RouteTables) < 1 {
		return nil, err
	}

	if err != nil {
		c.DeleteNetwork(*vpcOut.Vpc.VpcId)
		return nil, err
	}
	_, err = c.EC2.CreateRoute(&ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            gw.InternetGateway.InternetGatewayId,
		RouteTableId:         table.RouteTables[0].RouteTableId,
	})
	if err != nil {
		c.DeleteNetwork(*vpcOut.Vpc.VpcId)
		return nil, err
	}
	_, err = c.EC2.AssociateRouteTable(&ec2.AssociateRouteTableInput{
		RouteTableId: table.RouteTables[0].RouteTableId,
		SubnetId:     sn.Subnet.SubnetId,
	})
	if err != nil {
		c.DeleteNetwork(*vpcOut.Vpc.VpcId)
		return nil, err
	}

	req.GWRequest.PublicIP = true
	req.GWRequest.IsGateway = true
	req.GWRequest.NetworkIDs = append(req.GWRequest.NetworkIDs, *vpcOut.Vpc.VpcId)
	vm, err := c.CreateVM(req.GWRequest)
	if err != nil {
		c.DeleteNetwork(*vpcOut.Vpc.VpcId)
		return nil, wrapError("Error creating network", err)
	}
	net := api.Network{
		CIDR:      pStr(vpcOut.Vpc.CidrBlock),
		ID:        pStr(vpcOut.Vpc.VpcId),
		Name:      req.Name,
		IPVersion: req.IPVersion,
		GatewayID: vm.ID,
	}
	err = c.saveNetwork(net)
	if err != nil {
		c.DeleteNetwork(*vpcOut.Vpc.VpcId)
		return nil, err
	}
	return &net, nil
}

//GetNetwork returns the network identified by id
func (c *Client) GetNetwork(id string) (*api.Network, error) {
	net, err := c.getNetwork(id)
	if err != nil {
		return nil, err
	}
	out, err := c.EC2.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: []*string{aws.String(id)},
	})
	if err != nil {
		return nil, err
	}
	net.CIDR = *out.Vpcs[0].CidrBlock
	net.ID = *out.Vpcs[0].VpcId
	return net, nil
}

//ListNetworks lists available networks
func (c *Client) ListNetworks() ([]api.Network, error) {
	out, err := c.EC2.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}
	nets := []api.Network{}
	for _, vpc := range out.Vpcs {
		net, err := c.getNetwork(*vpc.VpcId)
		if err != nil {
			return nil, err
		}
		net.CIDR = *vpc.CidrBlock
		net.CIDR = *vpc.VpcId
		nets = append(nets, *net)
	}
	return nets, nil

}

//DeleteNetwork deletes the network identified by id
func (c *Client) DeleteNetwork(id string) error {
	net, err := c.getNetwork(id)
	if err == nil {
		c.DeleteVM(net.GatewayID)
		addrs, _ := c.EC2.DescribeAddresses(&ec2.DescribeAddressesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("domain"),
					Values: []*string{
						aws.String("vpc"),
					},
				},
				{
					Name: aws.String("instance-id"),
					Values: []*string{
						aws.String(net.GatewayID),
					},
				},
			},
		})
		for _, addr := range addrs.Addresses {
			c.EC2.DisassociateAddress(&ec2.DisassociateAddressInput{
				AssociationId: addr.AssociationId,
			})
			c.EC2.ReleaseAddress(&ec2.ReleaseAddressInput{
				AllocationId: addr.AllocationId,
			})
		}
	}

	_, err = c.EC2.DeleteVpc(&ec2.DeleteVpcInput{
		VpcId: aws.String(id),
	})
	return err
}

func (c *Client) getSubnets(vpcIDs []string) ([]*ec2.Subnet, error) {
	filters := []*ec2.Filter{}
	for _, id := range vpcIDs {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("vpc-id"),
			Values: []*string{&id},
		})
	}
	out, err := c.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}
	return out.Subnets, nil

}
func getState(state *ec2.InstanceState) (VMState.Enum, error) {
	// The low byte represents the state. The high byte is an opaque internal value
	// and should be ignored.
	//
	//    * 0 : pending
	//
	//    * 16 : running
	//
	//    * 32 : shutting-down
	//
	//    * 48 : terminated
	//
	//    * 64 : stopping
	//
	//    * 80 : stopped
	fmt.Println("State", state.Code)
	if state == nil {
		return VMState.ERROR, fmt.Errorf("Unexpected VM state")
	}
	if *state.Code == 0 {
		return VMState.STARTING, nil
	}
	if *state.Code == 16 {
		return VMState.STARTED, nil
	}
	if *state.Code == 32 {
		return VMState.STOPPING, nil
	}
	if *state.Code == 48 {
		return VMState.STOPPED, nil
	}
	if *state.Code == 64 {
		return VMState.STOPPING, nil
	}
	if *state.Code == 80 {
		return VMState.STOPPED, nil
	}
	return VMState.ERROR, fmt.Errorf("Unexpected VM state")
}

//Data structure to apply to userdata.sh template
type userData struct {
	//Name of the default user (api.DefaultUser)
	User string
	//Private key used to create the VM
	Key string
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

func (c *Client) prepareUserData(request api.VMRequest, kp *api.KeyPair, gw *api.VM) (string, error) {
	dataBuffer := bytes.NewBufferString("")
	var ResolveConf string
	var err error
	// if !request.PublicIP {
	// 	var buffer bytes.Buffer
	// 	for _, dns := range client.Cfg.DNSList {
	// 		buffer.WriteString(fmt.Sprintf("nameserver %s\n", dns))
	// 	}
	// 	ResolveConf = buffer.String()
	// }
	ip := ""
	if gw != nil {
		if len(gw.PrivateIPsV4) > 0 {
			ip = gw.PrivateIPsV4[0]
		} else if len(gw.PrivateIPsV6) > 0 {
			ip = gw.PrivateIPsV6[0]
		}
	}
	data := userData{
		User:        api.DefaultUser,
		Key:         strings.Trim(kp.PublicKey, "\n"),
		IsGateway:   request.IsGateway,
		AddGateway:  !request.PublicIP,
		ResolveConf: ResolveConf,
		GatewayIP:   ip,
	}
	err = c.UserDataTpl.Execute(dataBuffer, data)
	if err != nil {
		return "", err
	}
	encBuffer := bytes.Buffer{}
	enc := base64.NewEncoder(base64.StdEncoding, &encBuffer)
	enc.Write(dataBuffer.Bytes())
	return encBuffer.String(), nil
}

func (c *Client) saveVM(vm api.VM) error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(vm)
	if err != nil {
		return err
	}
	return c.PutObject("gpac.aws.wms", api.Object{
		Name:    vm.ID,
		Content: bytes.NewReader(buffer.Bytes()),
	})
}
func (c *Client) removeVM(vmID string) error {
	return c.DeleteObject("gpac.aws.wms", vmID)
}
func (c *Client) readVM(vmID string) (*api.VM, error) {
	o, err := c.GetObject("gpac.aws.wms", vmID, nil)
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

func (c *Client) createSecurityGroup(vpcID string, name string) (string, error) {
	out, err := c.EC2.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName: aws.String(name),
		VpcId:     aws.String(vpcID),
	})
	if err != nil {
		return "", err
	}
	_, err = c.EC2.AuthorizeSecurityGroupEgress(&ec2.AuthorizeSecurityGroupEgressInput{
		IpPermissions: []*ec2.IpPermission{
			&ec2.IpPermission{
				IpProtocol: aws.String("-1"),
			},
		},
	})
	if err != nil {
		return "", err
	}

	_, err = c.EC2.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		IpPermissions: []*ec2.IpPermission{
			&ec2.IpPermission{
				IpProtocol: aws.String("-1"),
			},
		},
	})
	if err != nil {
		return "", err
	}
	return *out.GroupId, nil
}

//CreateVM creates a VM that fulfils the request
func (c *Client) CreateVM(request api.VMRequest) (*api.VM, error) {

	//If no KeyPair is supplied a temporay one is created
	kp := request.KeyPair
	if kp == nil {
		kpTmp, err := c.CreateKeyPair(request.Name)
		if err != nil {
			return nil, err
		}
		defer c.DeleteKeyPair(kpTmp.ID)
		kp = kpTmp
	}
	//If the VM is not a Gateway, get gateway of the first network
	gwID := ""
	var gw *api.VM
	if !request.IsGateway {
		net, err := c.getNetwork(request.NetworkIDs[0])
		if err != nil {
			return nil, err
		}
		gwID = net.GatewayID
		gw, err = c.GetVM(gwID)
		if err != nil {
			return nil, err
		}

	}
	//get subnet of each network
	sns, err := c.getSubnets(request.NetworkIDs)

	//Prepare user data
	userData, err := c.prepareUserData(request, kp, gw)
	if err != nil {
		return nil, err
	}

	//Create networks interfaces
	networkInterfaces := []*ec2.InstanceNetworkInterfaceSpecification{}

	vpcs := map[string][]*ec2.Subnet{}
	for _, sn := range sns {
		vpcs[*sn.VpcId] = append(vpcs[*sn.VpcId], sn)

	}

	i := 0
	for _, netID := range request.NetworkIDs {
		if len(vpcs[netID]) < 1 {
			continue
		}
		sn := vpcs[netID][0]
		networkInterfaces = append(networkInterfaces, &ec2.InstanceNetworkInterfaceSpecification{
			SubnetId:                 sn.SubnetId,
			AssociatePublicIpAddress: aws.Bool(false),
			DeleteOnTermination:      aws.Bool(true),
			DeviceIndex:              aws.Int64(int64(i)),
		})
		i++
	}

	//Run instance
	out, err := c.EC2.RunInstances(&ec2.RunInstancesInput{
		ImageId:           aws.String(request.ImageID),
		KeyName:           aws.String(kp.Name),
		InstanceType:      aws.String(request.TemplateID),
		NetworkInterfaces: networkInterfaces,
		MaxCount:          aws.Int64(1),
		MinCount:          aws.Int64(1),
		UserData:          aws.String(userData),
		// TagSpecifications: []*ec2.TagSpecification{
		// 	{
		// 		Tags: []*ec2.Tag{
		// 			{
		// 				Key:   aws.String("Name"),
		// 				Value: aws.String(request.Name),
		// 			},
		// 		},
		// 	},
		// },
	})
	if err != nil {
		return nil, err
	}
	instance := out.Instances[0]

	netIFs, err := c.EC2.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("attachment.instance-id"),
				Values: []*string{instance.InstanceId},
			},
			&ec2.Filter{
				Name:   aws.String("attachment.device-index"),
				Values: []*string{aws.String(fmt.Sprintf("%d", 0))},
			},
		},
	})
	if err != nil {
		c.DeleteVM(*instance.InstanceId)
		return nil, err
	}

	addr, err := c.EC2.AllocateAddress(&ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
	})
	if err != nil {
		c.DeleteVM(*instance.InstanceId)
		return nil, err
	}
	//Wait that VM is started
	service := providers.Service{
		ClientAPI: c,
	}
	_, err = service.WaitVMState(*instance.InstanceId, VMState.STARTED, 120*time.Second)
	if err != nil {
		return nil, err
	}
	_, err = c.EC2.AssociateAddress(&ec2.AssociateAddressInput{
		NetworkInterfaceId: netIFs.NetworkInterfaces[0].NetworkInterfaceId,
		AllocationId:       addr.AllocationId,
	})
	if err != nil {
		c.DeleteVM(*instance.InstanceId)
		return nil, err
	}
	//Create api.VM

	tpl, err := c.GetTemplate(*instance.InstanceType)
	if err != nil {
		c.DeleteVM(*instance.InstanceId)
		return nil, err
	}
	v4IPs := []string{}
	for _, nif := range instance.NetworkInterfaces {
		v4IPs = append(v4IPs, *nif.PrivateIpAddress)
	}
	accessAddr := ""
	if instance.PublicIpAddress != nil {
		accessAddr = *instance.PublicIpAddress
	}
	state, err := getState(instance.State)
	if err != nil {
		c.DeleteVM(*instance.InstanceId)
		return nil, err
	}

	vm := api.VM{
		ID:           pStr(instance.InstanceId),
		Name:         request.Name,
		Size:         tpl.VMSize,
		PrivateIPsV4: v4IPs,
		AccessIPv4:   accessAddr,
		PrivateKey:   kp.PrivateKey,
		State:        state,
		GatewayID:    gwID,
	}
	c.saveVM(vm)
	return &vm, nil
}

//GetVM returns the VM identified by id
func (c *Client) GetVM(id string) (*api.VM, error) {

	out, err := c.EC2.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	})
	if err != nil {
		return nil, err
	}
	instance := out.Reservations[0].Instances[0]
	vm, err := c.readVM(id)
	if err != nil {
		vm = &api.VM{
			ID: *instance.InstanceId,
		}
	}

	vm.State, err = getState(instance.State)
	if err != nil {
		return nil, err
	}
	tpl, err := c.GetTemplate(*instance.InstanceType)
	if err != nil {
		return nil, err
	}
	vm.Size = tpl.VMSize
	v4IPs := []string{}
	for _, nif := range instance.NetworkInterfaces {
		v4IPs = append(v4IPs, *nif.PrivateIpAddress)
	}
	accessAddr := ""
	if instance.PublicIpAddress != nil {
		accessAddr = *instance.PublicIpAddress
	}
	vm.PrivateIPsV4 = v4IPs
	vm.AccessIPv4 = accessAddr

	return vm, nil
}

//ListVMs lists available VMs
func (c *Client) ListVMs() ([]api.VM, error) {
	panic("Not Implemented")
}

//DeleteVM deletes the VM identified by id
func (c *Client) DeleteVM(id string) error {
	c.removeVM(id)
	ips, err := c.EC2.DescribeAddresses(&ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("instance-id"),
				Values: []*string{aws.String(id)},
			},
		},
	})
	if err != nil {
		for _, ip := range ips.Addresses {
			c.EC2.ReleaseAddress(&ec2.ReleaseAddressInput{
				AllocationId: ip.AllocationId,
			})
		}
	}
	_, err = c.EC2.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	})
	return err

}

//StopVM stops the VM identified by id
func (c *Client) StopVM(id string) error {
	_, err := c.EC2.StopInstances(&ec2.StopInstancesInput{
		Force:       aws.Bool(true),
		InstanceIds: []*string{aws.String(id)},
	})
	return err
}

//StartVM starts the VM identified by id
func (c *Client) StartVM(id string) error {
	_, err := c.EC2.StartInstances(&ec2.StartInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	})
	return err
}

//GetSSHConfig creates SSHConfig from VM
func (c *Client) GetSSHConfig(vmID string) (*system.SSHConfig, error) {
	vm, err := c.GetVM(vmID)
	if err != nil {
		return nil, err
	}
	ip := vm.GetAccessIP()
	sshConfig := system.SSHConfig{
		PrivateKey: vm.PrivateKey,
		Port:       22,
		Host:       ip,
		User:       api.DefaultUser,
	}
	if vm.GatewayID != "" {
		gw, err := c.GetVM(vm.GatewayID)
		if err != nil {
			return nil, err
		}
		ip := gw.GetAccessIP()
		GatewayConfig := system.SSHConfig{
			PrivateKey: gw.PrivateKey,
			Port:       22,
			User:       api.DefaultUser,
			Host:       ip,
		}
		sshConfig.GatewayConfig = &GatewayConfig
	}

	return &sshConfig, nil
}

func toVolumeType(speed VolumeSpeed.Enum) string {
	switch speed {
	case VolumeSpeed.COLD:
		return "sc1"
	case VolumeSpeed.HDD:
		return "st1"
	case VolumeSpeed.SSD:
		return "gp2"
	}
	return "st1"
}

func toVolumeSpeed(t *string) VolumeSpeed.Enum {
	if t == nil {
		return VolumeSpeed.HDD
	}
	if *t == "sc1" {
		return VolumeSpeed.COLD
	}
	if *t == "st1" {
		return VolumeSpeed.HDD
	}
	if *t == "gp2" {
		return VolumeSpeed.SSD
	}
	return VolumeSpeed.HDD
}

func toVolumeState(s *string) VolumeState.Enum {
	// VolumeStateCreating = "creating"
	// VolumeStateAvailable = "available"
	// VolumeStateInUse = "in-use"
	// VolumeStateDeleting = "deleting"
	// VolumeStateDeleted = "deleted"
	// VolumeStateError = "error"
	if s == nil {
		return VolumeState.ERROR
	}
	if *s == "creating" {
		return VolumeState.CREATING
	}
	if *s == "available" {
		return VolumeState.AVAILABLE
	}
	if *s == "in-use" {
		return VolumeState.USED
	}
	if *s == "deleting" {
		return VolumeState.DELETING
	}
	if *s == "deleted" {
		return VolumeState.DELETING
	}
	if *s == "error" {
		return VolumeState.ERROR
	}
	return VolumeState.OTHER
}

func (c *Client) saveVolumeName(id, name string) error {
	return c.PutObject("gpac.aws.volumes", api.Object{
		Name:    id,
		Content: strings.NewReader(name),
	})
}

func (c *Client) getVolumeName(id string) (string, error) {
	obj, err := c.GetObject("gpac.aws.volumes", id, nil)
	if err != nil {
		return "", err
	}
	buffer := bytes.Buffer{}
	buffer.ReadFrom(obj.Content)
	return buffer.String(), nil
}

func (c *Client) removeVolumeName(id string) error {
	return c.DeleteObject("gpac.aws.volumes", id)
}

//CreateVolume creates a block volume
//- name is the name of the volume
//- size is the size of the volume in GB
//- volumeType is the type of volume to create, if volumeType is empty the driver use a default type
func (c *Client) CreateVolume(request api.VolumeRequest) (*api.Volume, error) {
	v, err := c.EC2.CreateVolume(&ec2.CreateVolumeInput{
		Size:       aws.Int64(int64(request.Size)),
		VolumeType: aws.String(toVolumeType(request.Speed)),
	})
	if err != nil {
		return nil, err
	}
	err = c.saveVolumeName(*v.VolumeId, request.Name)
	if err != nil {
		c.DeleteVolume(*v.VolumeId)
	}
	volume := api.Volume{
		ID:    pStr(v.VolumeId),
		Name:  request.Name,
		Size:  int(pInt64(v.Size)),
		Speed: toVolumeSpeed(v.VolumeType),
		State: toVolumeState(v.State),
	}
	return &volume, nil
}

//GetVolume returns the volume identified by id
func (c *Client) GetVolume(id string) (*api.Volume, error) {
	out, err := c.EC2.DescribeVolumes(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(id)},
	})
	if err != nil {
		return nil, err
	}
	v := out.Volumes[0]
	name, err := c.getVolumeName(id)
	if err != nil {
		return nil, err
	}
	volume := api.Volume{
		ID:    pStr(v.VolumeId),
		Name:  name,
		Size:  int(pInt64(v.Size)),
		Speed: toVolumeSpeed(v.VolumeType),
		State: toVolumeState(v.State),
	}
	return &volume, nil
}

//ListVolumes list available volumes
func (c *Client) ListVolumes() ([]api.Volume, error) {
	out, err := c.EC2.DescribeVolumes(&ec2.DescribeVolumesInput{})
	if err != nil {
		return nil, err
	}
	volumes := []api.Volume{}
	for _, v := range out.Volumes {
		name, err := c.getVolumeName(*v.VolumeId)
		if err != nil {
			return nil, err
		}
		volume := api.Volume{
			ID:    pStr(v.VolumeId),
			Name:  name,
			Size:  int(pInt64(v.Size)),
			Speed: toVolumeSpeed(v.VolumeType),
			State: toVolumeState(v.State),
		}
		volumes = append(volumes, volume)
	}

	return volumes, nil
}

//DeleteVolume deletes the volume identified by id
func (c *Client) DeleteVolume(id string) error {
	_, err := c.EC2.DeleteVolume(&ec2.DeleteVolumeInput{
		VolumeId: aws.String(id),
	})
	return err
}

// func (c *Client) saveVolumeAttachmentName(id, name string) error {
// 	return c.PutObject("__volume_atachements__", api.Object{
// 		Name:    id,
// 		Content: strings.NewReader(name),
// 	})
// }

// func (c *Client) getVolumeAttachmentName(id string) (string, error) {
// 	obj, err := c.GetObject("__volume_atachements__", id, nil)
// 	if err != nil {
// 		return "", err
// 	}
// 	buffer := bytes.Buffer{}
// 	buffer.ReadFrom(obj.Content)
// 	return buffer.String(), nil
// }

// func (c *Client) removeVolumeAttachmentName(id string) error {
// 	return c.DeleteObject("__volume_atachements__", id)
// }

// func vaID(vmID string, volumeID string) string {
// 	return fmt.Sprintf("%s###%s", vmID, volumeID)
// }

//CreateVolumeAttachment attaches a volume to a VM
//- name the name of the volume attachment
//- volume the volume to attach
//- vm the VM on which the volume is attached
func (c *Client) CreateVolumeAttachment(request api.VolumeAttachmentRequest) (*api.VolumeAttachment, error) {
	va, err := c.EC2.AttachVolume(&ec2.AttachVolumeInput{
		InstanceId: aws.String(request.ServerID),
		VolumeId:   aws.String(request.VolumeID),
	})
	if err != nil {
		return nil, err
	}
	return &api.VolumeAttachment{
		Device:   pStr(va.Device),
		ID:       pStr(va.VolumeId),
		Name:     request.Name,
		ServerID: pStr(va.InstanceId),
		VolumeID: pStr(va.VolumeId),
	}, nil
}

//GetVolumeAttachment returns the volume attachment identified by id
func (c *Client) GetVolumeAttachment(serverID, id string) (*api.VolumeAttachment, error) {
	out, err := c.EC2.DescribeVolumes(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(id)},
	})
	if err != nil {
		return nil, err
	}
	v := out.Volumes[0]
	for _, va := range v.Attachments {
		if *va.InstanceId == serverID {
			return &api.VolumeAttachment{
				Device:   pStr(va.Device),
				ServerID: pStr(va.InstanceId),
				VolumeID: pStr(va.VolumeId),
			}, nil
		}
	}
	return nil, fmt.Errorf("Volume attachment of volume %s on server %s does not exists", serverID, id)
}

//ListVolumeAttachments lists available volume attachment
func (c *Client) ListVolumeAttachments(serverID string) ([]api.VolumeAttachment, error) {
	out, err := c.EC2.DescribeVolumes(&ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("attachment.instance-id"),
				Values: []*string{aws.String(serverID)},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	vas := []api.VolumeAttachment{}
	for _, v := range out.Volumes {
		for _, va := range v.Attachments {
			vas = append(vas, api.VolumeAttachment{
				Device:   pStr(va.Device),
				ServerID: pStr(va.InstanceId),
				VolumeID: pStr(va.VolumeId),
			})
		}
	}
	return vas, nil

}

//DeleteVolumeAttachment deletes the volume attachment identifed by id
func (c *Client) DeleteVolumeAttachment(serverID, id string) error {
	_, err := c.EC2.DetachVolume(&ec2.DetachVolumeInput{
		InstanceId: aws.String(serverID),
		VolumeId:   aws.String(id),
	})
	return err
}

//CreateContainer creates an object container
func (c *Client) CreateContainer(name string) error {
	svc := s3.New(c.Session)
	input := &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(c.AuthOpts.Region),
		},
	}

	_, err := svc.CreateBucket(input)
	return err
}

//DeleteContainer deletes an object container
func (c *Client) DeleteContainer(name string) error {
	svc := s3.New(c.Session)
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}
	_, err := svc.DeleteBucket(input)
	return err
}

//ListContainers list object containers
func (c *Client) ListContainers() ([]string, error) {
	svc := s3.New(c.Session)
	input := &s3.ListBucketsInput{}

	result, err := svc.ListBuckets(input)
	if err != nil {
		return nil, err
	}
	buckets := []string{}
	for _, b := range result.Buckets {
		buckets = append(buckets, *b.Name)
	}
	return buckets, nil
}

func createTagging(m map[string]string) string {
	tags := []string{}
	for k, v := range m {
		tags = append(tags, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(tags, "&")
}

//PutObject put an object into an object container
func (c *Client) PutObject(container string, obj api.Object) error {
	svc := s3.New(c.Session)

	//Manage object life cycle
	expires := obj.DeleteAt != time.Time{}
	if expires {
		_, err := svc.PutBucketLifecycle(&s3.PutBucketLifecycleInput{
			Bucket: aws.String(container),
			LifecycleConfiguration: &s3.LifecycleConfiguration{
				Rules: []*s3.Rule{
					&s3.Rule{
						Expiration: &s3.LifecycleExpiration{
							Date: &obj.DeleteAt,
						},
						Prefix: aws.String(obj.Name),
						Status: aws.String("Enabled"),
					},
				},
			},
		})
		if err != nil {
			return err
		}
	}

	if obj.Metadata == nil {
		obj.Metadata = map[string]string{}
	}
	dateBytes, _ := time.Now().MarshalText()
	obj.Metadata["__date__"] = string(dateBytes)
	dateBytes, _ = obj.DeleteAt.MarshalText()
	obj.Metadata["__delete_at__"] = string(dateBytes)
	input := &s3.PutObjectInput{
		Body:        aws.ReadSeekCloser(obj.Content),
		Bucket:      aws.String(container),
		Key:         aws.String(obj.Name),
		ContentType: aws.String(obj.ContentType),
		Tagging:     aws.String(createTagging(obj.Metadata)),
	}

	_, err := svc.PutObject(input)

	return err
}

//UpdateObjectMetadata update an object into  object container
func (c *Client) UpdateObjectMetadata(container string, obj api.Object) error {
	//meta, err := c.GetObjectMetadata(container, obj.Name
	svc := s3.New(c.Session)
	tags := []*s3.Tag{}
	for k, v := range obj.Metadata {
		tags = append(tags, &s3.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	input := &s3.PutObjectTaggingInput{
		Bucket: aws.String(container),
		Key:    aws.String(obj.Name),
		Tagging: &s3.Tagging{
			TagSet: tags,
		},
	}
	_, err := svc.PutObjectTagging(input)
	return err
}

func pStr(s *string) string {
	if s == nil {
		var s string
		return s
	}
	return *s
}
func pTime(p *time.Time) time.Time {
	if p == nil {
		return time.Time{}
	}
	return *p
}
func pInt64(p *int64) int64 {
	if p == nil {
		var v int64
		return v
	}
	return *p
}

//GetObject get  object content from an object container
func (c *Client) GetObject(container string, name string, ranges []api.Range) (*api.Object, error) {
	svc := s3.New(c.Session)
	var rList []string
	for _, r := range ranges {
		rList = append(rList, r.String())
	}
	sRanges := strings.Join(rList, ",")
	out, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(container),
		Key:    aws.String(name),
		Range:  aws.String(sRanges),
	})
	if err != nil {
		return nil, err
	}

	obj, err := c.GetObjectMetadata(container, name)
	if err != nil {
		return nil, err
	}
	return &api.Object{
		Content:       aws.ReadSeekCloser(out.Body),
		ContentLength: pInt64(out.ContentLength),
		ContentType:   pStr(out.ContentType),
		LastModified:  pTime(out.LastModified),
		Name:          name,
		Metadata:      obj.Metadata,
		Date:          obj.Date,
		DeleteAt:      obj.DeleteAt,
	}, nil
}

//GetObjectMetadata get  object metadata from an object container
func (c *Client) GetObjectMetadata(container string, name string) (*api.Object, error) {
	svc := s3.New(c.Session)
	tagging, err := svc.GetObjectTagging(&s3.GetObjectTaggingInput{
		Bucket: aws.String(container),
		Key:    aws.String(name),
	})
	meta := map[string]string{}
	date := time.Time{}
	deleteAt := time.Time{}
	if err != nil {
		for _, t := range tagging.TagSet {
			if *t.Key == "__date__" {
				buffer := bytes.Buffer{}
				buffer.WriteString(*t.Value)
				date.UnmarshalText(buffer.Bytes())
			} else if *t.Key == "__delete_at__" {
				buffer := bytes.Buffer{}
				buffer.WriteString(*t.Value)
				deleteAt.UnmarshalText(buffer.Bytes())
			}
			meta[*t.Key] = *t.Value
		}
	}
	return &api.Object{
		Name:     name,
		Metadata: meta,
		Date:     date,
		DeleteAt: deleteAt,
	}, nil
}

//ListObjects list objects of a container
func (c *Client) ListObjects(container string, filter api.ObjectFilter) ([]string, error) {
	svc := s3.New(c.Session)
	var objs []string

	prefix := strings.Join([]string{filter.Path, filter.Prefix}, "/")
	err := svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(container),
		Prefix: aws.String(prefix),
	},
		func(out *s3.ListObjectsV2Output, last bool) bool {
			for _, o := range out.Contents {
				objs = append(objs, *o.Key)
			}
			return last
		},
	)
	if err != nil {
		return nil, err
	}
	return objs, err

}

//CopyObject copies an object
func (c *Client) CopyObject(containerSrc, objectSrc, objectDst string) error {
	svc := s3.New(c.Session)
	src := strings.Join([]string{containerSrc, objectDst}, "/")
	_, err := svc.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(containerSrc),
		Key:        aws.String(objectSrc),
		CopySource: aws.String(src),
	})
	return err
}

//DeleteObject deleta an object from a container
func (c *Client) DeleteObject(container, object string) error {
	svc := s3.New(c.Session)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(container),
		Key:    aws.String(object),
	})
	return err
}
