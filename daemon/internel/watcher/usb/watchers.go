package usb

import (
	"context"
	"os"
	"time"

	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

var _ watcher.Watcher = &usbWatcher{}

type usbWatcher struct{}

func NewUsbWatcher() *usbWatcher {
	w := &usbWatcher{}
	return w
}

var UsbSerialKey = struct{}{}

func WithSerial(ctx context.Context, serial string) context.Context {
	return context.WithValue(ctx, UsbSerialKey, serial)
}

func (w *usbWatcher) Watch(ctx context.Context) {
	retry := 1
	devs, err := utils.DetectdUsbDevices(ctx)
	for {
		if err != nil {
			klog.Error("list usb devices error, ", err)
			return
		}

		klog.Info("get usb device, ", devs)

		if len(devs) == 0 {
			if retry > 0 {
				delay := time.NewTimer(5 * time.Second)
				<-delay.C

				retry--
				devs, err = utils.DetectdUsbDevices(ctx)
				continue
			}
		}

		break
	}

	if _, err := os.Stat(commands.MOUNT_BASE_DIR); err != nil {
		if os.IsNotExist(err) {
			// mount dir not exists, terminus is not ready
			return
		}

		klog.Error("get stat error, ", err)
		return
	}

	serial := ctx.Value(UsbSerialKey).(string)
	if serial != "" {
		klog.Info("mount usb device with serial, ", serial)
		devs = utils.FilterArray(devs, utils.FilterBySerial(serial))
	}

	mountedPath, err := utils.MountUsbDevice(ctx, commands.MOUNT_BASE_DIR, devs)
	if err != nil {
		klog.Error("mount usb error, ", err)
		return
	}

	klog.Info("mount usb devices on paths, ", mountedPath)
}

var _ watcher.Watcher = &umountWatcher{}

type umountWatcher struct{}

func NewUmountWatcher() *umountWatcher {
	w := &umountWatcher{}
	return w
}

func (w *umountWatcher) Watch(ctx context.Context) {
	if err := utils.UmountBrokenMount(ctx, commands.MOUNT_BASE_DIR); err != nil {
		klog.Error("umount broken mount point error, ", err)
	}
}

func NewUsbMonitor(ctx context.Context) error {
	return utils.MonitorUsbDevice(ctx, func(action, serial string) error {
		switch action {
		case "add":
			delay := time.NewTimer(2 * time.Second)
			go func() {
				<-delay.C
				NewUsbWatcher().Watch(WithSerial(ctx, serial))
			}()
		case "remove":
			NewUmountWatcher().Watch(ctx)
		}

		return nil
	})
}
