#iPXE boot

Astral boot uses [iPXE](http://ipxe.org) as the initial tftp boot. This provides an HTTP interface and a menu system to kick off the install process.

The undionly.kpxe that is included with the astralboot distribution has been custom build with a user identifier to stop DHCP looping on boot. This is currently set to “skinny”

# Build your own undionly.kpxe

References : [Download](http://ipxe.org/download) and [Embed](http://ipxe.org/embed)

##Download 

```sh

git clone https://github.com/ipxe/ipxe 

cd ipxe/src

Create the embed file server.ipxe 

```
#!ipxe 

set user-class skinny
autoboot
```
Build the rom 

make EMBED=server.ipxe

Copy the bin/undionly.kpxe into the tftp folder 

