---
services: load-balancer
platforms: go
author: mcardosos
---

# Getting Started with Azure Resource Manager for load balancers in Go

This sample shows how to manage a load balancer using the Azure Resource Manager APIs for Go.

You can use a load balancer to provide high availability for your workloads in Azure. An Azure load balancer is a Layer-4 (TCP, UDP) type load balancer that distributes incoming traffic among healthy service instances in cloud services or virtual machines defined in a load balancer set.

For a detailed overview of Azure load balancers, see [Azure Load Balancer overview](https://azure.microsoft.com/documentation/articles/load-balancer-overview/).

![alt tag](./lb.JPG)

This sample deploys an internet-facing load balancer. It then creates and deploys two Azure virtual machines behind the load balancer. For a detailed overview of internet-facing load balancers, see [Internet-facing load balancer overview](https://azure.microsoft.com/documentation/articles/load-balancer-internet-overview/).

To deploy an internet-facing load balancer, you'll need to create and configure the following objects.

- Front end IP configuration - contains public IP addresses for incoming network traffic. 
- Back end address pool - contains network interfaces (NICs) for the virtual machines to receive network traffic from the load balancer. 
- Load balancing rules - contains rules mapping a public port on the load balancer to port in the back end address pool.
- Inbound NAT rules - contains rules mapping a public port on the load balancer to a port for a specific virtual machine in the back end address pool.
- Probes - contains health probes used to check availability of virtual machines instances in the back end address pool.

You can get more information about load balancer components with Azure resource manager at [Azure Resource Manager support for Load Balancer](https://azure.microsoft.com/documentation/articles/load-balancer-arm/).

## Tasks performed in this sample

The sample performs the following tasks to create the load balancer and the load-balanced virtual machines: 

1. Create a resource group
1. Create a storage account
1. Create a public IP
1. Create the load balancer with a payload that includes:
    - Front-end IP pool
    - Back-end address pool
    - Health probe
    - Load balancer rule
    - 2 NAT rules
1. Create a virtual network (vnet)
1. Create a subnet
1. Create an availability set
1. Create NIC 1 and the first VM: `Web1`
1. Create NIC 2 and the second VM: `Web2`
1. Delete the resource group and the resources created in the previous steps

## Run this sample

1. If you don't already have a Microsoft Azure subscription, you can register for a [free trial account](http://go.microsoft.com/fwlink/?LinkId=330212).

1. If you don't already have it, [install Go](https://golang.org/dl/).

1. Clone the repository.

    ```
    git clone https://github.com:Azure-Samples/network-go-manage-loadbalancer.git
    ```

1. Install the dependencies using [glide](https://github.com/Masterminds/glide).

    ```
    cd network-go-manage-loadbalancer
    glide install
    ```

1. Create an Azure service principal either through
    [Azure CLI](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal-cli/),
    [PowerShell](https://azure.microsoft.com/documentation/articles/resource-group-authenticate-service-principal/)
    or [the portal](https://azure.microsoft.com/documentation/articles/resource-group-create-service-principal-portal/).

1. Set the following environment variables using the information from the service principle that you created.

    ```
    export AZURE_TENANT_ID={your tenant id}
    export AZURE_CLIENT_ID={your client id}
    export AZURE_CLIENT_SECRET={your client secret}
    export AZURE_SUBSCRIPTION_ID={your subscription id}
    ```

    > [AZURE.NOTE] On Windows, use `set` instead of `export`.

1. Run the sample.

    ```
    go run example.go
    ```
    
## More information

- [Azure SDK for Go](http://github.com/Azure/azure-sdk-for-go) 