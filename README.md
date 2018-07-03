# XNFV Http Server Agent (XSHA)
Micro Http Server for Xenuim NFV testbed.

## Requirements
1. Any linux distribution with installed OVS 2.5+ (if you want full SFlow support) and ip command support
2. Install ovs-docker

```
    $ cd /usr/bin
    $ wget https://raw.githubusercontent.com/openvswitch/ovs/master/utilities/ovs-docker
    $ chmod a+rwx ovs-docker
```
3. Install docker

## Run and Build
Install Golang then Run below command

```
GOOS=linux GOARCH=amd64 go build -v
```

OR you can download the pre-build version "xsha" from root directory

```
sudo chmod +x xsha
sudo ./xsha
```

## Using XNFV Http Server Agent
You can Create any topology with this agent without using Xenium-NFV testbed. The XSHA using rest API to execute commands.

### Create Switch


    http://Agent-server-IP:8000/createSwitch

    json body (application/json)

    {
        "SwitchName": "br1",
        "SwitchControllerIp": "172.17.8.37",
        "SwitchControllerPort": "6633"
    }


### Delete Switch


    http://Agent-server-IP:8000/deleteSwitch

    json body (application/json)

    {
        "SwitchName": "br1"
    }


### Create switches Veth Pairs


    http://Agent-server-IP:8000/createVethPair

    json body (application/json)

    {
        "switchL": "br1",
        "switchR": "br2"
    }


### Delete switches Veth Pairs


    http://Agent-server-IP:8000/deleteVethPair

    json body (application/json)

    {
        "switchL": "br1",
        "switchR": "br2"
    }


### Create Docker container (As a VNF or Host)
    you can download basic VNF image from here
    https://hub.docker.com/r/nephilimboy/xnfv_vnf_basic/


    http://Agent-server-IP:8000/createVNFDocker

    json body (application/json)

    {
        "name": "u1",
        "img": "nephilimboy/xnfv_vnf_basic"
    }


### Delete Docker container


    http://Agent-server-IP:8000/deleteVNFDocker

    json body (application/json)

    {
        "name": "u1"
    }


### Connect Docker Container to switch


    http://Agent-server-IP:8000/createOVSDockerPort

    json body (application/json)

    {
        "vnfName": "u1",
        "vnfIpAddress": "10.10.10.1",
        "vnfInterfaceName": "eth0",
        "switchName": "br1"
    }


### Disconnect Docker Container from switch


    http://Agent-server-IP:8000/deleteOVSDockerPort

    json body (application/json)

    {
        "vnfName": "u1",
        "vnfInterfaceName": "eth0",
        "switchName": "br1"
    }


### Set SFlow Agent on Switch


    http://Agent-server-IP:8000/setSflowAgent

    json body (application/json)

    {
        "switchName": "br1",
        "agentId": "@sflow_br1",
        "senderInterface": "ens33",
        "collectorIp": "192.168.1.4",
        "collectorPort": "6343",
        "samplingRate": "10",
        "pollingRate": "10"
    }


### Remove SFlow Agent from Switch


    http://Agent-server-IP:8000/deleteSflowAgent

    json body (application/json)

    {
        "switchName": "br1",
        "agentId": "5e3dee16-55e7-449a-bc82-7dc2b476f821"
    }






