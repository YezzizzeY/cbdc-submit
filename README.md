# CBDC Experiment



## Technical report of setting up hyperledger fabric testnet, installing chaincodes and using sdk

### 1 Preparing prerequisites

#### 1.1 Golang

#### 1.2 Git

#### 1.3 cURL

#### 1.4 Docker

#### 1.5 Nodejs

#### 1.6 npm



### 2 Installing Fabric and Fabric Samples



#### 2.1 Download Fabric sample, Docker images and binaries

- Download a simple Fabric test network using Docker compose, and a set of sample applications that demonstrate its core capabilities. 
- Download precompiled Fabric CLI tool binaries and Fabric Docker Images which will be downloaded to environment.
- Downloads the latest Hyperledger Fabric Docker images and tags them as latest
- Downloads the following platform-specific Hyperledger Fabric CLI tool binaries and config files into the fabric-samples /bin and /config directories. (network.configtxgen, configtxlator, cryptogen, discover, idemixgen, orderer, osnadmin, peer, fabric-ca-client, fabric-ca-server)

```bash
curl -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh && chmod +x install-fabric.sh
```



#### 2.2 Choosing components

To specify the components to download add one or more of the following arguments. 

To pull the Docker containers and clone the samples repo, run one of these commands for example.

```bash
./install-fabric.sh docker samples binary
or
./install-fabric.sh d s b
```





### 3 Build up the network

After having downloaded the Hyperledger Fabric Docker images and samples, you can deploy a test network by using scripts that are provided in the fabric-samples repository. The test network is provided for learning about Fabric by running nodes on your local machine. Developers can use the network to test their smart contracts and applications. The network is meant to be used only as a tool for education and testing and not as a model for how to set up a network. 

In general, modifications to the scripts are discouraged and could break the network. It is based on a limited configuration that should not be used as a template for deploying a production network: It includes two peer organizations and an ordering organization. For simplicity, a single node Raft ordering service is configured. To reduce complexity, a TLS Certificate Authority (CA) is not deployed. All certificates are issued by the root CAs. The sample network deploys a Fabric network with Docker Compose. Because the nodes are isolated within a Docker Compose network, the test network is not configured to connect to other running Fabric nodes.



#### 3.1 Bring up the test network

```bash
cd fabric-samples/test-network
```

In this directory, you can find an annotated script, `network.sh`, that stands up a Fabric network using the Docker images on your local machine.

From inside the `test-network` directory, run the following command to remove any containers or artifacts from any previous runs:

```bash
./network.sh down
```

If we have used this before, the output is like:

```bash
root@ubuntu:/home/yezzi/Desktop/CBDC/fabric-samples/test-network# ./network.sh down

......
......
......


Removing ca_orderer             ... done
Removing ca_org2                ... done
Removing ca_org1                ... done
......
```



You can then bring up the network by issuing the following command. You will experience problems if you try to run the script from another directory:

```go
./network.sh up
```

This command creates a Fabric network that consists of two peer nodes, one ordering node. No channel is created when you run `./network.sh up`:

```bash
root@ubuntu:/home/yezzi/Desktop/CBDC/fabric-samples/test-network# ./network.sh up

......
......
......


Creating cli                    ... done
CONTAINER ID   IMAGE                               COMMAND             CREATED         STATUS                  PORTS                                                                                                                             NAMES
3169ca7e4b27   hyperledger/fabric-tools:latest     "/bin/bash"         1 second ago    Up Less than a second                                                                                                                                     cli
43a15c22b444   hyperledger/fabric-orderer:latest   "orderer"           2 seconds ago   Up Less than a second   0.0.0.0:7050->7050/tcp, :::7050->7050/tcp, 0.0.0.0:7053->7053/tcp, :::7053->7053/tcp, 0.0.0.0:9443->9443/tcp, :::9443->9443/tcp   orderer.example.com
4f41e88efd9a   hyperledger/fabric-peer:latest      "peer node start"   2 seconds ago   Up 1 second             0.0.0.0:7051->7051/tcp, :::7051->7051/tcp, 0.0.0.0:9444->9444/tcp, :::9444->9444/tcp                                              peer0.org1.example.com
b006a656b0cd   hyperledger/fabric-peer:latest      "peer node start"   2 seconds ago   Up 1 second             0.0.0.0:9051->9051/tcp, :::9051->9051/tcp, 7051/tcp, 0.0.0.0:9445->9445/tcp, :::9445->9445/tcp      
```



#### 3.2 Check the docker components and network design

After your test network is deployed, you can take some time to examine its components. Run the following command to list all of Docker containers that are running on your machine. You should see the three nodes that were created by the `network.sh` script:

```bash
docker ps -a
```

Output:

```bash
CONTAINER ID   IMAGE                               COMMAND             CREATED              STATUS              PORTS                                                                                                                             NAMES
3169ca7e4b27   hyperledger/fabric-tools:latest     "/bin/bash"         About a minute ago   Up About a minute                                                                                                                                     cli
43a15c22b444   hyperledger/fabric-orderer:latest   "orderer"           About a minute ago   Up About a minute   0.0.0.0:7050->7050/tcp, :::7050->7050/tcp, 0.0.0.0:7053->7053/tcp, :::7053->7053/tcp, 0.0.0.0:9443->9443/tcp, :::9443->9443/tcp   orderer.example.com
4f41e88efd9a   hyperledger/fabric-peer:latest      "peer node start"   About a minute ago   Up About a minute   0.0.0.0:7051->7051/tcp, :::7051->7051/tcp, 0.0.0.0:9444->9444/tcp, :::9444->9444/tcp                                              peer0.org1.example.com
b006a656b0cd   hyperledger/fabric-peer:latest      "peer node start"   About a minute ago   Up About a minute   0.0.0.0:9051->9051/tcp, :::9051->9051/tcp, 7051/tcp, 0.0.0.0:9445->9445/tcp, :::9445->9445/tcp     
```



#### 3.3 Creating a channel

Now that we have peer and orderer nodes running on our machine, we can use the script to create a Fabric channel for transactions between Org1 and Org2. Channels are a private layer of communication between specific network members. Channels can be used only by organizations that are invited to the channel, and are invisible to other members of the network. Each channel has a separate blockchain ledger. Organizations that have been invited “join” their peers to the channel to store the channel ledger and validate the transactions on the channel.

```bash
./network.sh createChannel
```

output:

```bash
+ configtxlator proto_encode --input config_update_in_envelope.json --type common.Envelope --output Org2MSPanchors.tx
2024-01-13 05:11:30.811 UTC 0001 INFO [channelCmd] InitCmdFactory -> Endorser and orderer connections initialized
2024-01-13 05:11:30.822 UTC 0002 INFO [channelCmd] update -> Successfully submitted channel update
Anchor peer set for org 'Org2MSP' on channel 'mychannel'
Channel 'mychannel' joined
root@ubuntu:/home/yezzi/Desktop/
```





#### 3.5 Installing chaincodes on the channel

In Fabric, smart contracts are deployed on the network in packages referred to as chaincode. A Chaincode is installed on the peers of an organization and then deployed to a channel, where it can then be used to endorse transactions and interact with the blockchain ledger. Before a chaincode can be deployed to a channel, the members of the channel need to agree on a chaincode definition that establishes chaincode governance. When the required number of organizations agree, the chaincode definition can be committed to the channel, and the chaincode is ready to be used.

```bash
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go
```





### 4 Test with CBDC Demo



#### 4.1 Get the CBDC github repo

First, we need to get the CBDC sdk and CBDC smart contract using github:

```bash
git clone https://github.com/YezzizzeY/CBDC.git
```



#### 4.2 Add CBDC_transaction doc folder to fabric-samples

download fabric_samples from github

https://github.com/hyperledger/fabric-samples

then put CBDC_transaction folder into this

CBDC_transaction has two folders, which are `application-gateway-go`, the SDK tool and  `CBDC_contract_go` which is the golang version of application of CBDC



#### 4.3 Setting basic network

This command will deploy the Fabric test network with two peers, an ordering service, and three certificate authorities (Orderer, Org1, Org2). Instead of using the cryptogen tool, we bring up the test network using certificate authorities, hence the `-ca` flag. Additionally, the org admin user registration is bootstrapped when the certificate authority is started.

```bash
./network.sh up createChannel -c mychannel -ca
```



(each time after using this, and before you close your VMware/ubuntu, please use ./network.sh down to delete images)



#### 4.4 Compile and deploy the CBDC smart contract

```bash
./network.sh deployCC -ccn testgo -ccp ../CBDC_transaction/CBDC_contract_go/ -ccl go
```



#### 4.5 Integrate SDK into our application

go into application-gateway-go and run 

```bash
go run assetTransfer.go
```

the success output is:

```bash
root@ubuntu:/home/yezzi/Desktop/CBDC/fabric-samples/CBDC_transaction/application-gateway-go# go run assetTransfer.go 
connect to contract succeed:  &{0xc0001eba10 0xc000150ec0 mychannel testgo }

--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger 
*** Transaction committed successfully

......
......
......


--> Submit Transaction: UpdateAsset asset70, asset70 does not exist and should return an error
*** Successfully caught the error:
Endorse error for transaction fbf3ccdd3d70a2f7bf18f96078d335671fb2db42b15c7c4e08b3e8563fdbc076 with gRPC status Aborted: rpc error: code = Aborted desc = failed to endorse transaction, see attached details for more info
Error Details:
- address: peer0.org1.example.com:7051, mspId: Org1MSP, message: chaincode response 500, Function UpdateAsset not found in contract SmartContract

```



### 5 SDK and smart contract

To help read the smart contract and CBDC sdk here, I've put the source code here

#### 5.1 smart contract

this is written by go, here is the core document, if you want to see the whole files, please refer to `CBDC_smart_contract_go/chaincode`

```go
package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

......
......
......



```



#### 5.2 SDK

It's in `application-gateway-go/assetTransfer.go`

```go
/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main


......
......
......



```



### 6 CBDC

can refer to https://github.com/mit-dci/opencbdc-tx

for example, in 2pc architecture https://github.com/mit-dci/opencbdc-tx/blob/trunk/docs/2pc_atomizer_user_guide.md

there is a problem -- after you install the requirements of opencbdc, 

and run `docker compose --file docker-compose-2pc.yml up --build`

the problem is you should add - after docker, for example: `docker-compose --file docker-compose-2pc.yml up --build`





### 7 Caliper



benchmark: https://hyperledger.github.io/caliper/v0.6.0/bench-config/

fabric config: https://hyperledger.github.io/caliper/v0.6.0/fabric-config/new/

One thing to note: when config fabric, you only need to define the MSP peer nodes

for example in organizations in `networkConfig.yaml`

the related dir of certificates is:

```
root@ubuntu:/home/yezzi/Desktop/CBDC/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp#
```

the related dir of peers is:

```
root@ubuntu:/home/yezzi/Desktop/CBDC/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/msp# 
```

Remember, what is configured in this file is the info of peer nodes organization

run caliper:

```
npx caliper launch manager \
    --caliper-workspace . \
    --caliper-benchconfig caliper-bench-config.yaml \
    --caliper-networkconfig networkConfig.yaml
```
