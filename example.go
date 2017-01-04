package main

import (
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	groupName              = "your-azure-sample-group"
	westus                 = "westus"
	vNetName               = "vNet"
	subnetName             = "subnet"
	ipName                 = "pip"
	frontEndIPConfigName   = "fip"
	backEndAddressPoolName = "backEndPool"
	probeName              = "probe"
	loadBalancerName       = "lb"
	storageAccountName     = "golangrocksonazure"
	vmName1                = "Web1"
	vmName2                = "Web2"

	groupClient     resources.GroupsClient
	lbClient        network.LoadBalancersClient
	vNetClient      network.VirtualNetworksClient
	subnetClient    network.SubnetsClient
	pipClient       network.PublicIPAddressesClient
	interfaceClient network.InterfacesClient
	availSetClient  compute.AvailabilitySetsClient
	accountClient   storage.AccountsClient
	vmClient        compute.VirtualMachinesClient
)

var (
	subscriptionID string
	spToken        *azure.ServicePrincipalToken
)

func init() {
	subscriptionID = getEnvVarOrExit("AZURE_SUBSCRIPTION_ID")
	tenantID := getEnvVarOrExit("AZURE_TENANT_ID")

	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(tenantID)
	onErrorFail(err, "OAuthConfigForTenant failed")

	clientID := getEnvVarOrExit("AZURE_CLIENT_ID")
	clientSecret := getEnvVarOrExit("AZURE_CLIENT_SECRET")
	spToken, err = azure.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	onErrorFail(err, "NewServicePrincipalToken failed")

	createClients()
}

func main() {
	fmt.Println("Creating resource group")
	resourceGroupParameters := resources.ResourceGroup{
		Location: &westus}
	_, err := groupClient.CreateOrUpdate(groupName, resourceGroupParameters)
	onErrorFail(err, "CreateOrUpdate failed")

	fmt.Println("Creating storage account")
	// apcp := storage.AccountPropertiesCreateParameters{}
	accountParameters := storage.AccountCreateParameters{
		Sku: &storage.Sku{
			Name: storage.StandardLRS,
		},
		Kind:     storage.Storage,
		Location: &westus,
		AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
		// AccountPropertiesCreateParameters: &apcp,
	}
	_, err = accountClient.Create(groupName, storageAccountName, accountParameters, nil)
	onErrorFail(err, "Create failed")

	fmt.Println("Creating public IP address")
	pip := network.PublicIPAddress{
		Location: &westus,
		PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: network.Static,
			DNSSettings: &network.PublicIPAddressDNSSettings{
				DomainNameLabel: to.StringPtr("domain-name"),
			},
		},
	}
	_, err = pipClient.CreateOrUpdate(groupName, ipName, pip, nil)
	onErrorFail(err, "CreateOrUpdate failed")

	fmt.Println("Getting public IP address info")
	pip, err = pipClient.Get(groupName, ipName, "")
	onErrorFail(err, "Get failed")

	fmt.Println("Creating load balancer")
	lb := network.LoadBalancer{
		Location: &westus,
		LoadBalancerPropertiesFormat: &network.LoadBalancerPropertiesFormat{
			FrontendIPConfigurations: &[]network.FrontendIPConfiguration{
				{
					Name: &frontEndIPConfigName,
					FrontendIPConfigurationPropertiesFormat: &network.FrontendIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: network.Dynamic,
						PublicIPAddress:           &pip,
					},
				},
			},
			BackendAddressPools: &[]network.BackendAddressPool{
				{
					Name: &backEndAddressPoolName},
			},
			Probes: &[]network.Probe{
				{
					Name: &probeName,
					ProbePropertiesFormat: &network.ProbePropertiesFormat{
						Protocol:          network.ProbeProtocolHTTP,
						Port:              to.Int32Ptr(80),
						IntervalInSeconds: to.Int32Ptr(15),
						NumberOfProbes:    to.Int32Ptr(4),
						RequestPath:       to.StringPtr("healthprobe.aspx"),
					},
				},
			},
			LoadBalancingRules: &[]network.LoadBalancingRule{
				{
					Name: to.StringPtr("lbRule"),
					LoadBalancingRulePropertiesFormat: &network.LoadBalancingRulePropertiesFormat{
						Protocol:             network.TransportProtocolTCP,
						FrontendPort:         to.Int32Ptr(80),
						BackendPort:          to.Int32Ptr(80),
						IdleTimeoutInMinutes: to.Int32Ptr(4),
						EnableFloatingIP:     to.BoolPtr(false),
						LoadDistribution:     network.Default,
						FrontendIPConfiguration: &network.SubResource{
							ID: to.StringPtr(buildID(subscriptionID, "frontendIPConfigurations", frontEndIPConfigName)),
						},
						BackendAddressPool: &network.SubResource{
							ID: to.StringPtr(buildID(subscriptionID, "backendAddressPools", backEndAddressPoolName)),
						},
						Probe: &network.SubResource{
							ID: to.StringPtr(buildID(subscriptionID, "probes", probeName)),
						},
					},
				},
			},
			InboundNatRules: &[]network.InboundNatRule{
				buildNATrule("natRule1", subscriptionID, 21),
				buildNATrule("natRule2", subscriptionID, 23),
			},
		},
	}
	_, err = lbClient.CreateOrUpdate(groupName, loadBalancerName, lb, nil)
	onErrorFail(err, "CreateOrUpdate failed")

	fmt.Println("Getting load balancer info")
	lb, err = lbClient.Get(groupName, loadBalancerName, "")
	onErrorFail(err, "CreateOrUpdate failed")

	fmt.Println("Creating virtual network")
	vNetParameters := network.VirtualNetwork{
		Location: &westus,
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{
				AddressPrefixes: &[]string{"10.0.0.0/16"},
			},
		},
	}
	_, err = vNetClient.CreateOrUpdate(groupName, vNetName, vNetParameters, nil)
	onErrorFail(err, "CreateOrUpdate failed")

	fmt.Println("Creating subnet")
	subnet := network.Subnet{
		SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
			AddressPrefix: to.StringPtr("10.0.0.0/24"),
		},
	}
	_, err = subnetClient.CreateOrUpdate(groupName, vNetName, subnetName, subnet, nil)
	onErrorFail(err, "CreateOrUpdate failed")

	fmt.Println("Getting subnet info")
	subnet, err = subnetClient.Get(groupName, vNetName, subnetName, "")
	onErrorFail(err, "Get failed")

	fmt.Println("Creating availability set")
	availSet := compute.AvailabilitySet{
		Location: &westus}
	availSet, err = availSetClient.CreateOrUpdate(groupName, "availSet", availSet)
	onErrorFail(err, "CreateOrUpdate failed")

	fmt.Printf("Creating virtual machine '%s'\n", vmName1)
	err = createVM(vmName1, subnet.ID, availSet.ID, pip.IPAddress, lb, 0)
	onErrorFail(err, "createVM failed")

	fmt.Printf("Creating virtual machine '%s'\n", vmName2)
	err = createVM(vmName2, subnet.ID, availSet.ID, pip.IPAddress, lb, 1)
	onErrorFail(err, "createVM failed")

	fmt.Println("Listing resources in resource group")
	list, err := groupClient.ListResources(groupName, "", "", nil)
	onErrorFail(err, "ListResources failed")
	fmt.Printf("Resources in '%s' resource group\n", groupName)
	for _, r := range *list.Value {
		fmt.Printf("----------------\nName: %s\nType: %s\n",
			*r.Name,
			*r.Type)
	}

	fmt.Println("Your load balancer and virtual machines have been created.")
	fmt.Print("Press enter to delete the resources created in this sample...")

	var input string
	fmt.Scanln(&input)

	fmt.Println("Deleting the resource group")
	_, err = groupClient.Delete(groupName, nil)
	onErrorFail(err, "Delete failed")

	fmt.Println("Done!")
}

// createClients initializes and adds token to all needed clients in the sample.
func createClients() {
	groupClient = resources.NewGroupsClient(subscriptionID)
	groupClient.Authorizer = spToken

	lbClient = network.NewLoadBalancersClient(subscriptionID)
	lbClient.Authorizer = spToken

	vNetClient = network.NewVirtualNetworksClient(subscriptionID)
	vNetClient.Authorizer = spToken

	subnetClient = network.NewSubnetsClient(subscriptionID)
	subnetClient.Authorizer = spToken

	pipClient = network.NewPublicIPAddressesClient(subscriptionID)
	pipClient.Authorizer = spToken

	interfaceClient = network.NewInterfacesClient(subscriptionID)
	interfaceClient.Authorizer = spToken

	availSetClient = compute.NewAvailabilitySetsClient(subscriptionID)
	availSetClient.Authorizer = spToken

	accountClient = storage.NewAccountsClient(subscriptionID)
	accountClient.Authorizer = spToken

	vmClient = compute.NewVirtualMachinesClient(subscriptionID)
	vmClient.Authorizer = spToken
}

// buildNATrule returns a network.InboundNatRule struct with all needed fields included.
func buildNATrule(natRuleName, subscriptionID string, frontEndPort int32) network.InboundNatRule {
	return network.InboundNatRule{
		Name: &natRuleName,
		InboundNatRulePropertiesFormat: &network.InboundNatRulePropertiesFormat{
			Protocol:             network.TransportProtocolTCP,
			FrontendPort:         to.Int32Ptr(frontEndPort),
			BackendPort:          to.Int32Ptr(22),
			EnableFloatingIP:     to.BoolPtr(false),
			IdleTimeoutInMinutes: to.Int32Ptr(4),
			FrontendIPConfiguration: &network.SubResource{
				ID: to.StringPtr(buildID(subscriptionID, "frontendIPConfigurations", frontEndIPConfigName)),
			},
		},
	}
}

// buildID returns a certain resource ID.
func buildID(subscriptionID, subType, subTypeName string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s/%s/%s",
		subscriptionID,
		groupName,
		loadBalancerName,
		subType,
		subTypeName)
}

// buildNICparams returns a network.Interface struct with all needed fields included.
func buildNICparams(subnetID *string, lb network.LoadBalancer, natRule int) network.Interface {
	return network.Interface{
		Location: &westus,
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				{
					Name: to.StringPtr("pipConfig"),
					InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
						Subnet: &network.Subnet{
							ID: subnetID,
						},
						LoadBalancerBackendAddressPools: &[]network.BackendAddressPool{
							{
								ID: (*lb.BackendAddressPools)[0].ID,
							},
						},
						LoadBalancerInboundNatRules: &[]network.InboundNatRule{
							{
								ID: (*lb.InboundNatRules)[natRule].ID,
							},
						},
					},
				},
			},
		},
	}
}

// createVM creates a VM, including its NIC.
func createVM(vmName string, subnetID, availSetID, ipAddress *string, lb network.LoadBalancer, natRule int) error {
	nicName := fmt.Sprintf("nic-%s", vmName)

	fmt.Printf("Creating NIC for '%s' machine\n", vmName)
	nic := buildNICparams(subnetID, lb, natRule)
	_, err := interfaceClient.CreateOrUpdate(groupName, nicName, nic, nil)
	if err != nil {
		return err
	}

	fmt.Printf("Getting NIC info for '%s' machine\n", vmName)
	nic, err = interfaceClient.Get(groupName, nicName, "")
	if err != nil {
		return err
	}

	fmt.Printf("Creating machine '%s'\n", vmName)
	vm := buildVMparams(vmName, nic.ID, availSetID)
	_, err = vmClient.CreateOrUpdate(groupName, vmName, vm, nil)
	if err != nil {
		return err
	}

	fmt.Printf("Now you can connect to '%s' via 'ssh %s@%s -p %v' with password '%s'\n",
		vmName,
		*vm.OsProfile.AdminUsername,
		*ipAddress,
		*(*lb.InboundNatRules)[natRule].FrontendPort,
		*vm.OsProfile.AdminPassword)

	return nil
}

// buildVMparams returns a network.VirtualMachine struct with all needed fields included.
func buildVMparams(vmName string, nicID, availSetID *string) compute.VirtualMachine {
	return compute.VirtualMachine{
		Location: &westus,
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			OsProfile: &compute.OSProfile{
				ComputerName:  &vmName,
				AdminUsername: to.StringPtr("notAdmin"),
				AdminPassword: to.StringPtr("Pa$$w0rd1975"),
			},
			HardwareProfile: &compute.HardwareProfile{
				VMSize: compute.StandardDS1,
			},
			StorageProfile: &compute.StorageProfile{
				ImageReference: &compute.ImageReference{
					Publisher: to.StringPtr("Canonical"),
					Offer:     to.StringPtr("UbuntuServer"),
					Sku:       to.StringPtr("16.04.0-LTS"),
					Version:   to.StringPtr("latest"),
				},
				OsDisk: &compute.OSDisk{
					Name:         to.StringPtr("osDisk"),
					Caching:      compute.None,
					CreateOption: compute.FromImage,
					Vhd: &compute.VirtualHardDisk{
						URI: to.StringPtr(buildVhdURI(storageAccountName, vmName)),
					},
				},
			},
			NetworkProfile: &compute.NetworkProfile{
				NetworkInterfaces: &[]compute.NetworkInterfaceReference{
					{
						ID: nicID,
						NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
							Primary: to.BoolPtr(true),
						},
					},
				},
			},
			AvailabilitySet: &compute.SubResource{
				ID: availSetID,
			},
		},
	}
}

// buildVhdURI returns the Vhd URI for a VM's OS disk.
func buildVhdURI(storageAccountName, vmName string) string {
	return fmt.Sprintf("https://%s.blob.core.windows.net/golangcontainer/%s.vhd",
		storageAccountName,
		vmName)
}

// getEnvVarOrExit returns the value of specified environment variable or terminates if it's not defined.
func getEnvVarOrExit(varName string) string {
	value := os.Getenv(varName)
	if value == "" {
		fmt.Printf("Missing environment variable %s\n", varName)
		os.Exit(1)
	}

	return value
}

// onErrorFail prints a failure message and exits the program if err is not nil.
func onErrorFail(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %s", message, err)
		os.Exit(1)
	}
}
