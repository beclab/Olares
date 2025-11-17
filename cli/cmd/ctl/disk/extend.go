package disk

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/beclab/Olares/cli/pkg/utils/lvm"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const defaultOlaresVGName = "olares-vg"

func NewExtendDiskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extend",
		Short: "extend disk operations",
		Run: func(cmd *cobra.Command, args []string) {
			// early return if no unmounted disks found
			unmountedDevices, err := lvm.FindUnmountedDevices()
			if err != nil {
				log.Fatalf("Error finding unmounted devices: %v\n", err)
			}

			if len(unmountedDevices) == 0 {
				log.Println("No unmounted disks found to extend.")
				return
			}

			// select volume group to extend
			currentVgs, err := lvm.FindCurrentLVM()
			if err != nil {
				log.Fatalf("Error finding current LVM: %v\n", err)
			}

			if len(currentVgs) == 0 {
				log.Println("No valid volume groups found to extend.")
				return
			}

			selectedVg, err := selectExtendingVG(currentVgs)
			if err != nil {
				log.Fatalf("Error selecting volume group: %v\n", err)
			}
			log.Printf("Selected volume group to extend: %s\n", selectedVg)

			// select logical volume to extend
			lvInVg, err := lvm.FindLvByVgName(selectedVg)
			if err != nil {
				log.Fatalf("Error finding logical volumes in volume group %s: %v\n", selectedVg, err)
			}

			if len(lvInVg) == 0 {
				log.Printf("No logical volumes found in volume group %s to extend.\n", selectedVg)
				return
			}

			selectedLv, err := selectExtendingLV(selectedVg, lvInVg)
			if err != nil {
				log.Fatalf("Error selecting logical volume: %v\n", err)
			}
			log.Printf("Selected logical volume to extend: %s\n", selectedLv)

			// select unmounted devices to create physical volume
			selectedDevice, err := selectExtendingDevices(unmountedDevices)
			if err != nil {
				log.Fatalf("Error selecting unmounted device: %v\n", err)
			}
			log.Printf("Selected unmounted device to use: %s\n", selectedDevice)

			options := &LvmExtendOptions{
				VgName:     selectedVg,
				DevicePath: selectedDevice,
				LvName:     selectedLv,
				DeviceBlk:  unmountedDevices[selectedDevice],
			}

			log.Printf("Extending logical volume %s in volume group %s using device %s\n", options.LvName, options.VgName, options.DevicePath)
			cleanupNeeded, err := options.cleanupDiskParts()
			if err != nil {
				log.Fatalf("Error during disk partition cleanup check: %v\n", err)
			}

			if cleanupNeeded {
				do, err := options.destroyWarning()
				if err != nil {
					log.Fatalf("Error during partition cleanup confirmation: %v\n", err)
				}
				if !do {
					log.Println("Operation aborted by user.")
					return
				}

				err = options.deleteDevicePartitions()
				if err != nil {
					log.Fatalf("Error deleting device partitions: %v\n", err)
				}

			} else {
				do, err := options.makeDecision()
				if err != nil {
					log.Fatalf("Error during extension confirmation: %v\n", err)
				}
				if !do {
					log.Println("Operation aborted by user.")
					return
				}
			}

			err = options.extendLVM()
			if err != nil {
				log.Fatalf("Error extending LVM: %v\n", err)
			}

			log.Println("Disk extension completed successfully.")

			// end of command run, and show result
			// show the result of the extension
			lvInVg, err = lvm.FindLvByVgName(selectedVg)
			if err != nil {
				log.Fatalf("Error finding logical volumes in volume group %s: %v\n", selectedVg, err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

			fmt.Fprint(w, "id\tLV\tVG\tLSize\tMountpoints\n")
			for idx, lv := range lvInVg {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", idx+1, lv.LvName, lv.VgName, lv.LvSize, strings.Join(lv.Mountpoints, ","))
			}
			w.Flush()

		},
	}

	return cmd
}

type LvmExtendOptions struct {
	VgName     string
	DevicePath string
	LvName     string
	DeviceBlk  *lvm.BlkPart
}

func selectExtendingVG(vgs []*lvm.VgItem) (string, error) {
	// if only one vg, return it directly
	if len(vgs) == 1 {
		return vgs[0].VgName, nil
	}

	reader, err := utils.GetBufIOReaderOfTerminalInput()
	if err != nil {
		return "", errors.Wrap(err, "failed to get terminal input reader")
	}

	fmt.Println("Multiple volume groups found. Please select one to extend:")
	fmt.Println("")
	// print header
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	fmt.Fprint(w, "id\tVG\tVSize\tVFree\n")
	for idx, vg := range vgs {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", idx+1, vg.VgName, vg.VgSize, vg.VgFree)
	}
	w.Flush()

LOOP:
	fmt.Printf("\nEnter the volume group id to extend: ")
	var input string
	input, err = reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return "", errors.Wrap(errors.WithStack(err), "read volume group id failed")
	}
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Printf("\ninvalid volume group id, please try again")
		goto LOOP
	}

	selectedIdx, err := strconv.Atoi(input)
	if err != nil || selectedIdx < 1 || selectedIdx > len(vgs) {
		fmt.Printf("\ninvalid volume group id, please try again")
		goto LOOP
	}

	return vgs[selectedIdx-1].VgName, nil
}

func selectExtendingLV(vgName string, lvs []*lvm.LvItem) (string, error) {
	if len(lvs) == 1 {
		return lvs[0].LvName, nil
	}

	if vgName == defaultOlaresVGName {
		selectedLv := ""
		for _, lv := range lvs {
			if lv.LvName == "root" {
				selectedLv = lv.LvName
				continue
			}

			if lv.LvName == "data" {
				selectedLv = lv.LvName
				break
			}
		}

		if selectedLv != "" {
			return selectedLv, nil
		}
	}

	reader, err := utils.GetBufIOReaderOfTerminalInput()
	if err != nil {
		return "", errors.Wrap(err, "failed to get terminal input reader")
	}

	fmt.Println("Multiple logical volumes found. Please select one to extend:")
	fmt.Println("")
	// print header
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	fmt.Fprint(w, "id\tLV\tVG\tLSize\tMountpoints\n")
	for idx, lv := range lvs {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", idx+1, lv.LvName, lv.VgName, lv.LvSize, strings.Join(lv.Mountpoints, ","))
	}
	w.Flush()

LOOP:
	fmt.Printf("\nEnter the logical volume id to extend: ")
	var input string
	input, err = reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return "", errors.Wrap(errors.WithStack(err), "read logical volume id failed")
	}
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Printf("\ninvalid logical volume id, please try again")
		goto LOOP
	}

	selectedIdx, err := strconv.Atoi(input)
	if err != nil || selectedIdx < 1 || selectedIdx > len(lvs) {
		fmt.Printf("\ninvalid logical volume id, please try again")
		goto LOOP
	}

	return lvs[selectedIdx-1].LvName, nil
}

func selectExtendingDevices(unmountedDevices map[string]*lvm.BlkPart) (string, error) {
	if len(unmountedDevices) == 0 {
		return "", errors.New("no unmounted devices available for selection")
	}

	if len(unmountedDevices) == 1 {
		for path := range unmountedDevices {
			return path, nil
		}
	}

	reader, err := utils.GetBufIOReaderOfTerminalInput()
	if err != nil {
		return "", errors.Wrap(err, "failed to get terminal input reader")
	}

	fmt.Println("Multiple unmounted devices found. Please select one to use:")
	fmt.Println("")
	// print header
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	fmt.Fprint(w, "id\tDevice\tSize\n")
	idx := 1
	devicePaths := make([]string, 0, len(unmountedDevices))
	for path, device := range unmountedDevices {
		fmt.Fprintf(w, "%d\t%s\t%s\n", idx, path, device.Size)
		devicePaths = append(devicePaths, path)
		idx++
	}
	w.Flush()

LOOP:
	fmt.Printf("\nEnter the device id to use: ")
	var input string
	input, err = reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return "", errors.Wrap(errors.WithStack(err), "read device id failed")
	}
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Printf("\ninvalid device id, please try again")
		goto LOOP
	}
	selectedIdx, err := strconv.Atoi(input)
	if err != nil || selectedIdx < 1 || selectedIdx > len(devicePaths) {
		fmt.Printf("\ninvalid device id, please try again")
		goto LOOP
	}

	return devicePaths[selectedIdx-1], nil
}

func (o LvmExtendOptions) destroyWarning() (bool, error) {
	reader, err := utils.GetBufIOReaderOfTerminalInput()
	if err != nil {
		return false, errors.Wrap(err, "failed to get terminal input reader")
	}

	fmt.Printf("WARNING: This will DESTROY all data on %s\n", o.DevicePath)
LOOP:
	fmt.Printf("Type 'YES' to continue, CTRL+C to abort: ")
	var input string
	input, err = reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return false, errors.Wrap(errors.WithStack(err), "read confirmation input failed")
	}
	input = strings.ToUpper(strings.TrimSpace(input))
	if input != "YES" {
		goto LOOP
	}
	return true, nil
}

func (o LvmExtendOptions) makeDecision() (bool, error) {
	reader, err := utils.GetBufIOReaderOfTerminalInput()
	if err != nil {
		return false, errors.Wrap(err, "failed to get terminal input reader")
	}

	fmt.Printf("NOTICE: Extending LVM will begin on device %s\n", o.DevicePath)
LOOP:
	fmt.Printf("Type 'YES' to continue, CTRL+C to abort: ")
	var input string
	input, err = reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return false, errors.Wrap(errors.WithStack(err), "read confirmation input failed")
	}
	input = strings.ToUpper(strings.TrimSpace(input))
	if input != "YES" {
		goto LOOP
	}
	return true, nil
}

func (o LvmExtendOptions) cleanupDiskParts() (bool, error) {
	if o.DeviceBlk == nil {
		return false, errors.New("device block is nil")
	}

	if len(o.DeviceBlk.Children) == 0 {
		return false, nil
	}

	return true, nil
}

func (o LvmExtendOptions) deleteDevicePartitions() error {
	log.Printf("Selected device %s has existing partitions. Cleaning up...\n", o.DevicePath)
	if o.DeviceBlk == nil {
		return errors.New("device block is nil")
	}

	if len(o.DeviceBlk.Children) == 0 {
		return nil
	}

	log.Printf("Deleting existing partitions on device %s...\n", o.DevicePath)
	var partitions []string
	for _, part := range o.DeviceBlk.Children {
		partitions = append(partitions, "/dev/"+part.Name)
	}

	vgs, err := lvm.FindVgsOnDevice(partitions)
	if err != nil {
		return errors.Wrap(err, "failed to find volume groups on device partitions")
	}

	if len(vgs) > 0 {
		log.Println("existing volume group on device, delete it first")
		for _, vg := range vgs {
			lvs, err := lvm.FindLvByVgName(vg.VgName)
			if err != nil {
				return errors.Wrapf(err, "failed to find logical volumes in volume group %s", vg.VgName)
			}

			err = lvm.DeactivateLv(vg.VgName)
			if err != nil {
				return errors.Wrapf(err, "failed to deactivate volume group %s", vg.VgName)
			}

			for _, lv := range lvs {
				err = lvm.RemoveLv(lv.LvPath)
				if err != nil {
					return errors.Wrapf(err, "failed to remove logical volume %s", lv.LvPath)
				}
			}

			err = lvm.RemoveVg(vg.VgName)
			if err != nil {
				return errors.Wrapf(err, "failed to remove volume group %s", vg)
			}

			err = lvm.RemovePv(vg.PvName)
			if err != nil {
				return errors.Wrapf(err, "failed to remove physical volume %s", vg.PvName)
			}
		}

	}

	log.Printf("Deleting partitions on device %s...\n", o.DevicePath)
	err = lvm.DeleteDevicePartitions(o.DevicePath)
	if err != nil {
		return errors.Wrapf(err, "failed to delete partitions on device %s", o.DevicePath)
	}

	return nil
}

func (o LvmExtendOptions) extendLVM() error {
	log.Printf("Creating partition on device %s...\n", o.DevicePath)
	err := lvm.MakePartOnDevice(o.DevicePath)
	if err != nil {
		return errors.Wrapf(err, "failed to create partition on device %s", o.DevicePath)
	}

	log.Printf("Creating physical volume on device %s...\n", o.DevicePath)
	err = lvm.AddNewPV(o.DevicePath, o.VgName)
	if err != nil {
		return errors.Wrapf(err, "failed to create physical volume on device %s", o.DevicePath)
	}

	log.Printf("Extending volume group %s with logic volume %s on  device %s...\n", o.VgName, o.LvName, o.DevicePath)
	err = lvm.ExtendLv(o.VgName, o.LvName)
	if err != nil {
		return errors.Wrapf(err, "failed to extend logical volume %s in volume group %s", o.LvName, o.VgName)
	}

	return nil
}
