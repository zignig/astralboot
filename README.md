# astralboot

a low level boot server that deploys directly out of ipfs

## Usage

TODO

## Example

1. set up a vm with 2 interfaces:
  - normal network
  - a shared network on a private vm network.
2. set the ip of the private to the bottom of the range eg `192.168.5.1`
3. set the config to use that interface by name `eth1`
4. create other VMs attached to the private network.
5. run program :)

if you want to dev your own boot sequence:

```
ipfs get -o=data <hash in config>
```
