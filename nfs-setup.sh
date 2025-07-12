#!/bin/bash
apt-get update
apt-get install -y nfs-kernel-server
mkfs.ext4 -F /dev/sdb
mkdir -p /mnt/data
mount /dev/sdb /mnt/data
echo '/mnt/data 10.0.1.0/28(rw,sync,no_subtree_check)' >> /etc/exports
exportfs -a
systemctl restart nfs-kernel-server
