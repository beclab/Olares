package utils

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/disk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type MountedPathDetail struct {
	disk.UsageStat `json:",inline"`
	Type           string `json:"type"`
	Invalid        bool   `json:"invalid"`
	IDSerial       string `json:"id_serial"`
	IDSerialShort  string `json:"id_serial_short"`
	PartitionUUID  string `json:"partition_uuid"`
	Device         string `json:"device"`
	ReadOnly       bool   `json:"read_only"`
}

func NotifyBrokenMounts(ctx context.Context) {
	// find the master files pod
	client, err := GetKubeClient()
	if err != nil {
		klog.Error("failed to get kube client: ", err)
		return
	}

	pods, err := client.CoreV1().Pods("os-framework").List(ctx, metav1.ListOptions{LabelSelector: "app=files"})
	if err != nil {
		klog.Error("failed to list pods: ", err)
		return
	}

	nodeName, _, nodeRole, err := GetThisNodeName(ctx, client)
	if err != nil {
		klog.Error("failed to get this node name: ", err)
		return
	}

	for _, pod := range pods.Items {
		if pod.Spec.NodeName == nodeName && nodeRole == "master" {
			status := pod.Status.Phase
			if status != "Running" {
				klog.Warningf("pod %s is not running, status: %s, skip", pod.Name, status)
				continue
			}

			ip := pod.Status.PodIP
			if ip == "" {
				klog.Warningf("pod %s has no IP, skip", pod.Name)
				return
			}

			mountedPath, err := GetMountedPathDetail(ctx, func(us *disk.UsageStat) *disk.UsageStat {
				path := us.Path
				if strings.HasPrefix(path, commands.MOUNT_BASE_DIR) {
					path = strings.TrimPrefix(path, commands.MOUNT_BASE_DIR+"/")
				} else {
					// not in cluster path, ignore
					return nil
				}

				us.Path = path

				return us
			})
			if err != nil {
				klog.Error("failed to get mounted path detail: ", err)
				return
			}

			// send notification to the files pod, and let it handle the broken mount point
			url := "http://" + ip + ":8080/api/mounted_states/"
			httpClient := resty.New().SetTimeout(5 * time.Second)
			res, err := httpClient.R().
				SetContext(ctx).
				SetHeader("Content-Type", "application/json").
				SetBody(mountedPath).Post(url)

			if err != nil {
				klog.Error("failed to send notification: ", err)
				return
			}

			if res.StatusCode() != 200 {
				klog.Error("failed to send notification, status code: ", res.StatusCode())
				return
			}

			klog.Infof("successfully sent broken mount notification to pod %s", pod.Name)
			return
		}
	}

	klog.Warning("no files pod found on this node to notify")
}

func GetMountedPathDetail(ctx context.Context, mutate func(*disk.UsageStat) *disk.UsageStat) ([]*MountedPathDetail, error) {
	paths, err := MountedPath(ctx)
	if err != nil {
		return nil, err
	}

	klog.Info("mounted path, ", paths)

	var res []*MountedPathDetail
	var mountedMountPoints []string
	for _, p := range paths {
		mountedMountPoints = append(mountedMountPoints, p.Path)
		u, err := disk.UsageWithContext(ctx, p.Path)
		if err != nil {
			klog.Error("get path usage error, ", err, ", ", p)

			u = &disk.UsageStat{Path: p.Path}
			p.Invalid = true
		}

		if mutate != nil {
			u = mutate(u)
		}

		if u != nil {
			res = append(res, &MountedPathDetail{
				*u,
				string(p.Type),
				p.Invalid,
				p.IDSerial,
				p.IDSerialShort,
				p.PartitionUUID,
				p.Device,
				p.ReadOnly,
			})
		}
	}

	records, err := LoadMountRecords(commands.MOUNT_RECORDS_FILE)
	if err != nil {
		klog.Warning("load mount records error, ", err)
	}
	for _, r := range records {
		if r.Type != SMB && r.Type != NFS {
			continue
		}
		if slices.Contains(mountedMountPoints, r.MountPoint) {
			continue
		}
		device := r.SmbPath
		if r.Type == NFS {
			device = r.Server + ":" + r.NfsPath
		}
		u := &disk.UsageStat{Path: r.MountPoint}
		if mutate != nil {
			u = mutate(u)
		}
		if u != nil {
			res = append(res, &MountedPathDetail{
				*u,
				string(r.Type),
				true,
				"",
				"",
				"",
				device,
				false,
			})
		}
	}

	return res, nil
}
