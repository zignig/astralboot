#Astral Boot

Simon Kirkby 
tigger@interthingy.com
20150304

This repository has been moved to gb format  https://getgb.io/ , it’s cool.

## Description 

Astralboot is a golang server that provides network services to boot virtual and metal machines from pxe boot.
The following services are provided

1. DHCP , for ip allocation and boot information
2. TFTP , simple file transfer
3. HTTP , for serving images and configs

It can pull its data files out of [ipfs](http://ipfs.io/), which means that they are downloaded on request and then stored locally.

Local file serving also works with local file system folders ( see INSTRUCTIONS for details )

## Required for development

1. golang dev environment
2. running ipfs node
3. a network under your control.

## Warning

As this server has a naive dhcp server it can be dangerous to run in an office environment. Running this server can interfere with normal network services. 

## Installation

assumes a working golang environment.

```sh

git clone github.com/zignig/astralboot

cd astralboot

gb build

```

also the ipfs service, which is currently  in alpha , is available from http://github.com/ipfs/go-ipfs

will need to be installed and running 

## Setup 

Testing so far has been done on a virtual machine with two network interfaces, one on a home network and the other an isolated VM network.

This machine will probably need to have masquerading setup , this is not needed for astral boot , but is for the machines to access the internet.

enable forwarding 
```sh
echo 1 > /proc/sys/net/ipv4/ip_forward
```
make it stick 
```sh
edit /etc/sysctl.conf  and change  net.ipv4.ip_forward = 1
```
change the firewall 
```sh
/sbin/iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
```
The test machines to be bootstrapped have one network interface attached to isolated VM network.

The default hashes for booting are included in the git repository , put them into place by running.
```sh
cp refs.toml.dist refs.toml
```

It is worth noting that this will need to be run as root , dhcp, tftp and http not running on the machine

Now you are good to go, run the astralboot binary

1. If the config.toml file does not exists it will ask some questions to set up
2. It will grab some files from ipfs ( or local file system )and load up the various operating system files.
3. On the first run it will populate the leases.db file with empty ip addresses.
4. All the services will start and it will be ready to serve.

Verbosity can be changed by adding -v , -vv and -vvv to the command line.

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
```sh
ipfs get -o=data “hash from the refs.toml file”
```
If the config has IPFS = false the local file system will be used.

# Development

all comments, patches and pull requests welcome

# TODO 

1. Better templating of preseed 
2. Add more operating systems
4. More stuff
