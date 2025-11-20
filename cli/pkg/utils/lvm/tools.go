package lvm

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"slices"
)

func FindCurrentLVM() ([]*VgItem, error) {
	VG := CommandVGS()
	result, errmsg, err := VG.Run()
	if err != nil {
		log.Printf("failed to run vgs command: %s \n%s\n", err, errmsg)
		return nil, err
	}

	if len(result.Report) == 0 || len(result.Report[0].Vg) == 0 {
		err = errors.New("no volume groups found")
		return nil, err
	}

	var vgs []*VgItem
	for _, vg := range result.Report[0].Vg {
		if vg.PvCount == "0" || vg.LvCount == "0" {
			continue
		}
		vgs = append(vgs, &vg)
	}

	if len(vgs) == 0 {
		err = errors.New("no valid volume groups found")
		return nil, err
	}

	return vgs, nil
}

func FindUnmountedDevices() (map[string]*BlkPart, error) {
	lblkCmd := CommandLBLK()
	result, errmsg, err := lblkCmd.Run()
	if err != nil {
		log.Printf("failed to run lsblk command: %s \n%s\n", err, errmsg)
		return nil, err
	}

	var unmountedDevices map[string]*BlkPart = make(map[string]*BlkPart)
	var unmountedPart func(part BlkPart) bool
	unmountedPart = func(part BlkPart) bool {
		if len(part.Mountpoints) > 0 {
			return false
		}

		if len(part.Mountpoints) == 0 && len(part.Children) == 0 {
			return true
		}

		for _, child := range part.Children {
			if !unmountedPart(child) {
				return false
			}
		}
		return true
	}

	for _, dev := range result.Blockdevices {
		if dev.Type != "disk" {
			continue
		}

		if unmountedPart(dev) {
			unmountedDevices["/dev/"+dev.Name] = &dev
		}
	}

	return unmountedDevices, nil
}

func FindLvByVgName(vgName string) ([]*LvItem, error) {
	LV := CommandLVS()
	result, errmsg, err := LV.Run("-o", "+lv_dm_path,lv_path")
	if err != nil {
		log.Printf("failed to run lvs command: %s \n%s\n", err, errmsg)
		return nil, err
	}

	if len(result.Report) == 0 || len(result.Report[0].Lv) == 0 {
		return nil, nil
	}

	var lvs []*LvItem
	for _, lv := range result.Report[0].Lv {
		if lv.VgName == vgName {
			mountpoints, err := FindMountpointsByLvDmPath(lv.LvDmPath)
			if err == nil {
				lv.Mountpoints = mountpoints
			}
			lvs = append(lvs, &lv)
		}
	}

	return lvs, nil
}

func FindMountpointsByLvDmPath(lvDmPath string) ([]string, error) {
	FINDMNT := CommandFindMnt()
	result, errmsg, err := FINDMNT.Run(lvDmPath)
	if err != nil && errmsg != "" {
		log.Printf("failed to run findmnt command: %s \n%s\n", err, errmsg)
		return nil, err
	}

	if len(result.Filesystems) == 0 {
		return nil, nil
	}

	var mountpoints []string
	for _, fs := range result.Filesystems {
		mountpoints = append(mountpoints, fs.Target)
	}

	return mountpoints, nil
}

/*
wipefs -a /dev/nvme0n1
sgdisk --zap-all /dev/nvme0n1
*/
func DeleteDevicePartitions(devicePath string) error {
	c, err := exec.Command("wipefs", "-a", devicePath).CombinedOutput()
	if err != nil {
		log.Printf("failed to wipe device %s: %s\n", devicePath, c)
		return err
	}

	// c, err = exec.Command("sgdisk", "--zap-all", devicePath).CombinedOutput()
	// if err != nil {
	// 	log.Printf("failed to zap device %s: %s\n", devicePath, c)
	// 	return err
	// }

	return nil
}

/*
sudo parted /dev/sdX mklabel gpt
sudo parted -a optimal /dev/sdX mkpart primary 1MiB 100%
*/
func MakePartOnDevice(devicePath string) error {
	c, err := exec.Command("parted", devicePath, "mklabel", "gpt").CombinedOutput()
	if err != nil {
		log.Printf("failed to make partition table on device %s: %s\n", devicePath, c)
		return err
	}

	c, err = exec.Command("parted", "-a", "optimal", devicePath, "mkpart", "primary", "1MiB", "100%").CombinedOutput()
	if err != nil {
		log.Printf("failed to make partition on device %s: %s\n", devicePath, c)
		return err
	}

	return nil
}

/*
sudo pvcreate /dev/sdX1
sudo vgextend target_vg /dev/sdX1
*/
func AddNewPV(devicePath string, vg string) error {
	partition := devicePath + "p1"
	c, err := exec.Command("pvcreate", partition).CombinedOutput()
	if err != nil {
		log.Printf("failed to create physical volume on device %s: %s\n", partition, c)
		return err
	}

	c, err = exec.Command("vgextend", vg, partition).CombinedOutput()
	if err != nil {
		log.Printf("failed to extend volume group %s with device %s: %s\n", vg, partition, c)
		return err
	}

	return nil
}

/*
lvextend -l +100%FREE "/dev/$VG_NAME/$LV_ROOT_NAME"
resize2fs "/dev/$VG_NAME/$LV_ROOT_NAME"
*/
func ExtendLv(vg, lv string) error {
	c, err := exec.Command("lvextend", "-l", "+100%FREE", "/dev/"+vg+"/"+lv).CombinedOutput()
	if err != nil {
		log.Printf("failed to extend logical volume %s in volume group %s: %s\n", lv, vg, c)
		return err
	}

	c, err = exec.Command("resize2fs", "/dev/"+vg+"/"+lv).CombinedOutput()
	if err != nil {
		log.Printf("failed to resize filesystem on logical volume %s in volume group %s: %s\n", lv, vg, c)
		return err
	}

	return nil
}

func DeactivateLv(vg string) error {
	c, err := exec.Command("lvchange", "-an", vg).CombinedOutput()
	if err != nil {
		log.Printf("failed to deactivate logical volume in volume group %s: %s\n", vg, c)
		return err
	}

	return nil
}

func RemoveLv(lvpath string) error {
	_, err := os.Stat(lvpath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		log.Printf("failed to stat logical volume %s: %s\n", lvpath, err)
		return err
	}

	c, err := exec.Command("lvremove", "-f", lvpath).CombinedOutput()
	if err != nil {
		log.Printf("failed to remove logical volume %s: %s\n", lvpath, c)
		return err
	}

	return nil
}

func RemoveVg(vg string) error {
	c, err := exec.Command("vgremove", "-f", vg).CombinedOutput()
	if err != nil {
		log.Printf("failed to remove volume group %s: %s\n", vg, c)
		return err
	}

	return nil
}

func RemovePv(pv string) error {
	c, err := exec.Command("pvremove", "-f", pv).CombinedOutput()
	if err != nil {
		log.Printf("failed to remove physical volume %s: %s\n", pv, c)
		return err
	}

	return nil
}

func FindVgsOnDevice(devicePaths []string) ([]*VgItem, error) {
	VG := CommandVGS()
	result, errmsg, err := VG.Run("-o", "+pv_name")
	if err != nil {
		log.Printf("failed to run vgs command: %s \n%s\n", err, errmsg)
		return nil, err
	}

	var vgs []*VgItem
	for _, vg := range result.Report[0].Vg {
		if slices.Contains(devicePaths, vg.PvName) {
			vgs = append(vgs, &vg)
		}
	}

	return vgs, nil
}
