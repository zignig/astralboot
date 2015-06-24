#/bin/sh

echo "file system image"
wget -O coreos_production_pxe_image.cpio.gz http://alpha.release.core-os.net/amd64-usr/current/coreos_production_pxe_image.cpio.gz
eche "boot kernel"
wget -O coreos_production_pxe.vmlinuz http://alpha.release.core-os.net/amd64-usr/current/coreos_production_pxe.vmlinuz


