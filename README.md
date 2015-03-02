#Astral Boot

Simon Kirkby 
tigger@interthingy.com
20150304

## Description 

Astralboot is a golang server that provides network services to boot virtual and metal machines from pxe boot.
The following services are provided

1. DHCP , for ip allocation and boot information
2. TFTP , simple file transfer
3  HTTP , for serving images and configs

It pulls its data files out of [ipfs](http://ipfs.io/), which means that they are downloaded on request and then stored locally.

## Required 

1. golang dev environment
2. running ipfs node
3. a network under your control.

## Warning

As this server has a naive dhcp server it can be dangerous to run in an office environment. Running this server can interfere with normal network services. 

## Installation

assumes a working golang environment.

```go get github.com/zignig/astralboot

cd $GOPATH/src/github.com/zignig/astralboot

go build
```

so the ipfs service, which is currently  in alpha , is available from http://github.com/jbenet/go-ipfs

will need to be installed and running 

## Setup 

Testing so far has been done on a virtual machine with two network interfaces, one on a home network and the other an isolated VM network.

This machine will probably need to have masquerading setup , this is not needed for astral boot , but is for the machines to access the internet.

enable forwarding 

`echo 1 > /proc/sys/net/ipv4/ip_forward`

make it stick 

edit /etc/sysctl.conf  and change  net.ipv4.ip_forward = 1

change the firewall 

`/sbin/iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE`

The test machines to be bootstrapped have one network interface attached to isolated VM network.

Once you have the astral boot binary built, edit the config.toml file for interfaces on your virtual machines.

a minimal config is

```ref = “QmCoolIPFSHash”
interface = “eth1”
```

It is worth noting that this will need to be run as root , dhcp, tftp and http not running on the machine

Now you are good to go, run the astralboot binary

1. It will grab some files from ipfs and load up the various operating system files.
2. On the first run it will populate the leases.db file with empty ip addresses.
3. All the services will start and it will be ready to serve.

Now comes the fun bit ....

# Running the server

Create a virtual machine that is connected to the isolated network that astralboot is serving on and configure it to PXE boot.

The boot sequence should happen in this order

1. The new virtual server should ask for an ip address.
2. Astral boot serves an address with extra information pointing back to the astral bootserver.
3. A undionly.kpxe image is served to the machine.
4. it asks again for an ip address ( it will get the same address ).
5. A menu to select the operating system is presented on the boot line.
6. Select the OS of you choice ( coreos , or debian at this point ).
7. It will boot the server.

Debian will be fairly quick , coreos will take some time as the .gz file is 165 Mb , so it will take some time to download 

To precache the files into ipfs, run  ipfs refs -r HashFromConfigFile and it will download everything

# Changing the Services

As the server boots it will show an implied config , this shows possible entries to the config file to change.

Developing boot services, To develop modified boot services it is possible to serve the files from disk rather than ifps 
Downloading the files can be done with the following ipfs commands

In the astralboot folder : 

`ipfs get -o=data “hash from the config file”`

then run astralboot with a -l ( l for larry ) flag an it will use the local file system.

# Development

all comments, patches and pull requests welcome

# TODO 

1. Better templating of preseed and cloudconfig
2. Add more operating systems
3. Have subclasses on each operating system.
4. Add DNS server 
5. More stuff


