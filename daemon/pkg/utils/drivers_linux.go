//go:build linux && cgo
// +build linux,cgo

package utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"bytetrade.io/web3os/terminusd/pkg/commands"
	mountutils "k8s.io/mount-utils"

	udev "github.com/jochenvg/go-udev"
	"github.com/rubiojr/go-usbmon"
	"k8s.io/klog/v2"
)

func detectdStorageDevices(ctx context.Context, bus string) (usbDevs []storageDevice, err error) {
	u := udev.Udev{}
	e := u.NewEnumerate()

	// Add some FilterAddMatchSubsystemDevtype
	e.AddMatchSubsystem("scsi")
	e.AddMatchProperty("DEVTYPE", "scsi_device")
	e.AddMatchIsInitialized()
	devices, err := e.Devices()
	if err != nil {
		return
	}

	var usbs []*udev.Device
	addDevice := func(ds []*udev.Device) {
		for _, d := range ds {
			if d.Properties()["ID_BUS"] == bus {
				usbs = append(usbs, d)
			} else if d.Properties()["ID_BUS"] == "ata" &&
				d.Properties()["ID_USB_TYPE"] == "disk" &&
				bus == "usb" {
				usbs = append(usbs, d)
			}
		}
	}
	for _, device := range devices {
		ec := u.NewEnumerate()
		ec.AddMatchParent(device)
		ec.AddMatchSubsystem("block")
		ec.AddMatchProperty("DEVTYPE", "partition")
		ec.AddMatchIsInitialized()

		children, err := ec.Devices()
		if err != nil {
			return nil, err
		}

		if len(children) > 0 {
			addDevice(children)
		} else {
			ec := u.NewEnumerate()
			ec.AddMatchParent(device)
			ec.AddMatchSubsystem("block")
			ec.AddMatchProperty("DEVTYPE", "disk")
			ec.AddMatchIsInitialized()

			children, err = ec.Devices()
			if err != nil {
				return nil, err
			}

			addDevice(children)
		}

	}

	for _, device := range usbs {
		syspath := device.Syspath()
		// fmt.Println("devtype:", device.Devtype(),
		// 	"syspath: ", syspath,
		// 	"subsystem: ", device.Subsystem(),
		// )

		token := strings.Split(syspath, "/")
		devPath := filepath.Join("/dev", token[len(token)-1])
		klog.Info("device path:", device.Properties())
		vender := device.Properties()["ID_VENDOR"]
		if vender == "" {
			vender = device.Properties()["ID_USB_VENDOR"]
		}

		idSerial := device.Properties()["ID_SERIAL"]
		idSerialShort := device.Properties()["ID_SERIAL_SHORT"]
		partUUID := device.Properties()["ID_PART_ENTRY_UUID"]

		usbDevs = append(usbDevs, storageDevice{
			DevPath:       devPath,
			Vender:        vender,
			IDSerial:      idSerial,
			IDSerialShort: idSerialShort,
			PartitionUUID: partUUID,
		})
	}

	return
}

func DetectdUsbDevices(ctx context.Context) (usbDevs []storageDevice, err error) {
	return detectdStorageDevices(ctx, "usb")
}

func DetectdHddDevices(ctx context.Context) (usbDevs []storageDevice, err error) {
	return detectdStorageDevices(ctx, "ata")
}

func MonitorUsbDevice(ctx context.Context, cb func(action string) error) error {
	filter := &usbmon.ActionFilter{Action: usbmon.ActionAll}
	devs, err := usbmon.ListenFiltered(ctx, filter)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case dev := <-devs:
				fmt.Printf("-- Device %s\n", dev.Action())
				fmt.Println("Serial: " + dev.Serial())
				fmt.Println("Path: " + dev.Path())
				fmt.Println("Vendor: " + dev.Vendor())

				if cb != nil {
					err = cb(dev.Action())
					if err != nil {
						klog.Error("usb action callback error, ", err, ", ", dev.Action())
					}
				}
			}
		}
	}()

	klog.Info("start to monitor usb devices")
	return nil
}

func getMountedPath(devs []storageDevice) ([]string, error) {
	mounter := mountutils.New("")
	list, err := mounter.List()
	if err != nil {
		klog.Error("list mount path error, ", err)
		return nil, err
	}

	var paths []string
	for _, m := range list {
		if slices.ContainsFunc(devs, func(u storageDevice) bool { return u.DevPath == m.Device }) {
			klog.Infof("mount: %v, %v, %v", m.Path, m.Device, devs)
			paths = append(paths, m.Path)
		}
	}

	return paths, nil

}

func MountedUsbPath(ctx context.Context) ([]string, error) {
	usbs, err := DetectdUsbDevices(ctx)
	if err != nil {
		return nil, err
	}

	if len(usbs) == 0 {
		return nil, nil
	}

	return getMountedPath(usbs)
}

func MountedHddPath(ctx context.Context) ([]string, error) {
	hdds, err := DetectdHddDevices(ctx)
	if err != nil {
		return nil, err
	}

	if len(hdds) == 0 {
		return nil, nil
	}

	return getMountedPath(hdds)
}

func MountUsbDevice(ctx context.Context, mountBaseDir string, dev []storageDevice) (mountedPath []string, err error) {
	mounter := mountutils.New("")
	mountedList, err := mounter.List()
	if err != nil {
		klog.Error("list mount path error, ", err)
		return nil, err
	}

	isMounted := func(devPath string) (bool, string) {
		for _, m := range mountedList {
			if devPath == m.Device {
				return true, m.Path
			}
		}

		return false, ""
	}

	for i, d := range dev {
		mountDirPrefix := d.Vender
		if mountDirPrefix == "" {
			mountDirPrefix = "disk"
		}
		mountDir := filepath.Join(mountBaseDir, fmt.Sprintf("%s-%d", mountDirPrefix, i))
		if ok, p := isMounted(d.DevPath); ok {
			mountedPath = append(mountedPath, p)
			continue
		}

		// try to make dir
		// try 100 another paths
		mkMountDir := mountDir
		foundDir := false
		for n := 0; n < 100; n++ {
			err = os.Mkdir(mkMountDir, 0755)
			if err != nil {
				if os.IsExist(err) {
					var empty bool
					empty, err = IsEmptyDir(mkMountDir)
					if err != nil {
						klog.Error("check dir is empty error, ", err)
						break
					}

					if !empty {
						mkMountDir = fmt.Sprintf("%s-%d", mountDir, n)
						continue
					}

					// exists a empty dir
					foundDir = true
					break
				}

				klog.Error("mkdir error, ", err, ", ", mkMountDir)
				return
			}

			// success to make empty mount dir
			foundDir = true
			break
		} // end loop retry

		if !foundDir {
			continue
		}

		if err = mounter.Mount(d.DevPath, mkMountDir, "", []string{"uid=1000", "gid=1000"}); err != nil {
			klog.Warning("mount usb error, ", err, ", ", d.DevPath, ", ", mkMountDir)
			// clear the empty mount dir
			if err = os.RemoveAll(mkMountDir); err != nil {
				klog.Error("remove the mount dir error, ", err)
			}

		} else {
			mountedPath = append(mountedPath, mkMountDir)
		}
	} // end loop dev

	return
}

func umountAndRemovePath(ctx context.Context, path string) error {
	mounter := mountutils.New("")
	err := mounter.Unmount(path)
	if err != nil {
		klog.Error("umount path error, ", err, ", ", path)
		return err
	}

	if err = os.RemoveAll(path); err != nil {
		klog.Error("remove mount point error, ", err)
	}

	return err
}

// check the cifs mount point if the network is broken
// since the cifs will reconnect by itself, so if network broken for 2 minutes,
// we think it's really broken
type latestConnected struct {
	lastCheck time.Time
	invalid   bool
}

var hostsLastestConnected map[string]latestConnected = map[string]latestConnected{}

// tryConnect try to connect to a samba service with specified host and port.
func tryConnect(host string, port string) bool {
	timeout := time.Second * 2
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		klog.Errorf("Try to connect: %s:%s err=%v", host, port, err)
		return false
	}
	if conn != nil {
		defer conn.Close()
		return true
	}

	return false
}

func cifsBroken(mountPoint *mountutils.MountPoint) (broken, invalid bool) {
	if strings.HasPrefix(mountPoint.Device, "//") {
		token := strings.Split(strings.TrimPrefix(mountPoint.Device, "//"), "/")
		if len(token) > 0 {
			host := token[0]
			port := "445" // default samba port

			if ok := tryConnect(host, port); ok {
				hostsLastestConnected[host] = latestConnected{time.Now(), false}
				return false, false
			}

			lastestConnection, ok := hostsLastestConnected[host]
			if !ok || lastestConnection.invalid == false {
				lastestConnection = latestConnected{time.Now(), true}
			} else {
				lastestConnection.invalid = true
			}
			hostsLastestConnected[host] = lastestConnection

			// cannot be connected
			if time.Since(lastestConnection.lastCheck) > 2*time.Minute {
				return true, true
			} else {
				return false, true
			}
		}
	}

	return false, false
}

func IsCifsInvalid(mountPoint *mountutils.MountPoint) bool {
	if strings.HasPrefix(mountPoint.Device, "//") {
		token := strings.Split(strings.TrimPrefix(mountPoint.Device, "//"), "/")
		if len(token) > 0 {
			host := token[0]

			if connection, ok := hostsLastestConnected[host]; ok {
				return connection.invalid
			}
		}
	}

	return false
}

/*
umount mount point if it's an usb device and remove the mount point path
*/
func UmountUsbDevice(ctx context.Context, path string) error {
	paths, err := MountedUsbPath(ctx)
	if err != nil {
		return err
	}

	if slices.Contains(paths, path) {
		return umountAndRemovePath(ctx, path)
	}

	return errors.New("not a mounted usb path")
}

func UmountBrokenUsbMount(ctx context.Context, baseDir string) error {
	usbdevs, err := DetectdUsbDevices(ctx)
	if err != nil {
		return err
	}

	hdddevs, err := DetectdHddDevices(ctx)
	if err != nil {
		return err
	}

	mounter := mountutils.New("")
	list, err := mounter.List()
	if err != nil {
		klog.Error("list mount path error, ", err)
		return err
	}

	for _, m := range list {
		if strings.HasPrefix(m.Path, baseDir) && !strings.HasPrefix(m.Path, path.Join(baseDir, "ai")) {
			if !slices.ContainsFunc(usbdevs, func(u storageDevice) bool { return u.DevPath == m.Device }) &&
				!slices.ContainsFunc(hdddevs, func(u storageDevice) bool { return u.DevPath == m.Device }) {

				if m.Type == "cifs" {
					if broken, _ := cifsBroken(&m); !broken {
						continue
					}
				}

				klog.Infof("broken mountpoint: %v, %v, %v", m.Path, m.Device, usbdevs)

				if err = umountAndRemovePath(ctx, m.Path); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// apt install cifs-utils
func MountSambaDriver(ctx context.Context, mountBaseDir string, smbPath string, user, pwd string) error {
	mounter := mountutils.New("")

	if !strings.HasPrefix(smbPath, "//") {
		return fmt.Errorf("invalid samba shared path, %v", smbPath)
	}

	smbPath = strings.TrimRight(smbPath, "/")
	pathToken := strings.Split(strings.TrimLeft(smbPath, "//"), "/")
	if len(pathToken) < 2 {
		return fmt.Errorf("invalid samba shared path, %v", smbPath)
	}

	sharePath := pathToken[len(pathToken)-1]
	mntPath := filepath.Join(mountBaseDir, sharePath)
	err := os.MkdirAll(mntPath, 0755)
	if err != nil {
		klog.Error("create mount path error, ", err)
		return err
	}

	var opts []string
	if user != "" {
		opts = append(opts, "user="+user)
	}
	if pwd != "" {
		opts = append(opts, "password="+pwd)
	}

	// check duplicate mount
	mountedPath, err := MountedPath(ctx)
	if err != nil {
		klog.Warning("list mounted path error, ", err)
	} else {
		for _, m := range mountedPath {
			if m.Path == mntPath {
				return errors.New("duplicate mounted path")
			}
		}
	}

	opts = append(opts, "uid=1000", "gid=1000", "cache=none", "fsc", "noserverino")
	err = mounter.Mount(smbPath, mntPath, "cifs", opts)
	if err != nil {
		klog.Error("mount path as rw error, ", err)

		// retry to mount as read-only
		opts = append(opts, "ro")
		err = mounter.Mount(smbPath, mntPath, "cifs", opts)
		if err != nil {
			if e := os.Remove(mntPath); e != nil {
				klog.Error("remove dir error, ", e, ", ", mntPath)
			}
		}
	}

	return err
}

func UmountSambaDriver(ctx context.Context, mountDir string) error {
	mounter := mountutils.New("")

	err := mounter.Unmount(mountDir)
	if err != nil {
		klog.Error("umount path error, ", err)
		return err
	}

	return os.Remove(mountDir)
}

func ForceMountHdd(ctx context.Context) {
	devs, err := DetectdHddDevices(ctx)
	if err != nil {
		klog.Error("detect hdd devices error, ", err)
		return
	}

	if len(devs) > 0 {
		mounted, err := getMountedPath(devs)
		if err != nil {
			klog.Error("get mounted hdd error, ", err)
			return
		}

		if len(mounted) < len(devs) {
			cmd := exec.CommandContext(ctx, "mount", "-a")
			cmd.Env = os.Environ()
			output, err := cmd.CombinedOutput()
			klog.Info(string(output))

			if err != nil {
				klog.Error("exec cmd error, ", err, ", mount -a")
				return
			}

			// chown
			mounted, err = getMountedPath(devs)
			if err != nil {
				klog.Error("get mounted hdd error, ", err)
				return
			}

			for _, m := range mounted {
				if !strings.HasPrefix(m, commands.OS_ROOT_DIR) {
					// ignore out of control path
					continue
				}
				cmd := exec.CommandContext(ctx, "chown", "-R", "1000:1000", m)
				cmd.Env = os.Environ()
				output, err = cmd.CombinedOutput()
				klog.Info(string(output))
				if err != nil {
					klog.Error("exec cmd error, ", err, ", chown -R 1000:1000 ", m)
				}
			}
		}
	}
}

func isReadOnly(mp *mountutils.MountPoint) bool {
	return slices.Contains(mp.Opts, "ro")
}

func MountedSambaPath(ctx context.Context) ([]mountedPath, error) {
	mounter := mountutils.New("")
	list, err := mounter.List()
	if err != nil {
		klog.Error("list mount path error, ", err)
		return nil, err
	}

	var paths []mountedPath
	for _, m := range list {
		if m.Type == "cifs" {
			paths = append(paths, mountedPath{m.Path, SMB, IsCifsInvalid(&m), "", "", "", m.Device, isReadOnly(&m)})
		}
	}

	return paths, nil

}

func MountedPath(ctx context.Context) ([]mountedPath, error) {
	usbs, err := DetectdUsbDevices(ctx)
	if err != nil {
		return nil, err
	}

	hdds, err := DetectdHddDevices(ctx)
	if err != nil {
		return nil, err
	}

	mounter := mountutils.New("")
	list, err := mounter.List()
	if err != nil {
		klog.Error("list mount path error, ", err)
		return nil, err
	}

	var paths []mountedPath
	for _, m := range list {
		idx := -1
		switch {
		case func() bool {
			idx = slices.IndexFunc(usbs, func(u storageDevice) bool { return u.DevPath == m.Device })
			return idx >= 0
		}():
			paths = append(paths, mountedPath{m.Path, USB, false, usbs[idx].IDSerial, usbs[idx].IDSerialShort, usbs[idx].PartitionUUID, "", false})
		case func() bool {
			idx = slices.IndexFunc(hdds, func(u storageDevice) bool { return u.DevPath == m.Device })
			return idx >= 0
		}():
			paths = append(paths, mountedPath{m.Path, HDD, false, hdds[idx].IDSerial, hdds[idx].IDSerialShort, hdds[idx].PartitionUUID, "", false})
		case m.Type == "cifs":
			paths = append(paths, mountedPath{m.Path, SMB, IsCifsInvalid(&m), "", "", "", m.Device, isReadOnly(&m)})
		}

	}

	return paths, nil
}
