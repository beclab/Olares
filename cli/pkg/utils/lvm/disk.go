package lvm

/*
wipefs -a /dev/nvme0n1
sgdisk --zap-all /dev/nvme0n1

findmnt -n -J --target /olares
{
   "filesystems": [
      {
         "target": "/olares",
         "source": "/dev/mapper/olares--vg-data[/olares]",
         "fstype": "ext4",
         "options": "rw,relatime"
      }
   ]
}

lvs --reportformat json /dev/mapper/olares--vg-data


sudo parted -a optimal /dev/sdX mkpart primary 1MiB 100%
sudo pvcreate /dev/sdX1
sudo vgextend target_vg /dev/sdX1

*/
