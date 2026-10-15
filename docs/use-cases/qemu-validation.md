---
outline: [2, 3]
description: Validate QEMU Server installations, SSH connections, resource changes, persistence, and independent Linux instances created with Clone.
head:
  - - meta
    - name: keywords
      content: "Olares, QEMU Server, Linux, acceptance tests, SSH, Clone, persistence"
app_version: "0.0.27"
doc_version: "1.0"
doc_updated: "2026-10-16"
---

# Verify your QEMU Server installation

Use these acceptance checks after installing QEMU Server or changing its configuration. They cover the user workflows in [Run Linux servers with QEMU Server](qemu.md). A test procedure is not a claim that every platform or network path has already been verified.

## Prepare the checks

Use an instance with no critical workload and retain a separate backup. Resource tests restart the VM and interrupt connections. Allow enough host capacity to run a second instance for Clone checks.

Record the application version, Olares version, host architecture, distribution, instance title, and date. Record connection addresses privately; do not put passwords in test reports.

For the resource tests, start with 1 vCPU, 2 GiB RAM, and a 20 GiB disk. These are test values, not installation defaults. Run installation and resource checks sequentially for Ubuntu 24.04 LTS, Debian 13, Rocky Linux 10, and Fedora 44.

## Installation and console

| ID | Procedure | Expected result |
| --- | --- | --- |
| QEMU-01 | Install each distribution from Market with an explicitly entered password. Open the application and wait for initialization. Run `cat /etc/os-release` after login. | The page reaches the ready state. Linux identifies the selected distribution and release. |
| QEMU-02 | Open the browser console and sign in with the distribution's account and installation password. | Boot output and a usable shell are available. The account can use `sudo`. |
| QEMU-03 | During an active image download, restart the application once, then reopen its page. | Download resumes using retained data when supported by the server. Verification and initialization eventually complete. |

## SSH access

| ID | Procedure | Expected result |
| --- | --- | --- |
| QEMU-04 | From the same LAN, enable Overlay gateway for the instance. Copy its assigned IP and connect with SSH on port 22. Run `cat /etc/os-release`. | Password authentication succeeds and the session reaches the intended Linux instance. |
| QEMU-05 | Move the computer to another network. Enable LarePass VPN. Obtain the Olares device address and this instance's external SSH port from Settings. Connect with `ssh -p PORT USER@DEVICE_ADDRESS`, replacing the capitalized values. | Authentication succeeds without relying on same-LAN access. The selected distribution is shown. |
| QEMU-06 | Repeat the SSH login using macOS or Linux OpenSSH and Windows OpenSSH or Xshell. | All clients reach the same instance using the appropriate host, port, username, and password. |

For QEMU-05, testing while still on the same LAN is insufficient evidence of remote connectivity. Record the external network used and verify that the connection depends on the intended VPN route.

## Resource changes and persistence

Before changing resources, record a baseline inside Linux:

```bash
cat /etc/os-release
nproc
free -h
lsblk -b -o NAME,SIZE,TYPE,MOUNTPOINTS
df -h /
findmnt -n -o FSTYPE /
cat /etc/machine-id
cat /proc/sys/kernel/random/boot_id
printf 'qemu-persistence-check\n' > ~/qemu-persistence-check.txt
```

| ID | Procedure | Expected result |
| --- | --- | --- |
| QEMU-07 | Set CPU to 2 and RAM to 3 GiB in the instance's environment variables. Apply and reconnect after initialization. | `nproc` reports 2. Memory increases to approximately the configured capacity. The boot ID changes. |
| QEMU-08 | Increase disk capacity from 20 to 24 GiB. Apply, reconnect, and inspect `lsblk` and `df -h /`. | The virtual disk is 24 GiB. The root filesystem grows; its capacity remains below the whole disk capacity. Files remain readable. |
| QEMU-09 | Set CPU back to 1, RAM back to 2 GiB, and request a 16 GiB disk. Apply and reconnect. | CPU and RAM decrease. The disk remains 24 GiB and the page shows the unsupported-shrink warning. |
| QEMU-10 | Restart the application. Read the marker file, machine ID, boot ID, and disk capacity again. | The marker contains `qemu-persistence-check`; the machine ID is unchanged; the boot ID changes; the 24 GiB disk remains. |

Check the marker with `cat ~/qemu-persistence-check.txt` after every restart. Record the root filesystem type: Ubuntu and Debian use ext4 in the verified images, Rocky Linux uses XFS, and Fedora uses Btrfs. Validate growth inside the guest rather than relying only on the application page's requested capacity.

## Clone and instance isolation

1. Keep the Ubuntu instance and its marker file from the previous checks. Record its machine ID and connection details.
2. In Market, open **My Olares**, choose **Clone** from the instance's menu, and create `Debian Services` with Debian 13, separate resource settings, and its own password.
3. Open the clone and wait until ready. Connect as `debian` and inspect `/etc/os-release` and `/etc/machine-id`.
4. Enable Overlay gateway for both instances, one after the other. Reopen the gateway page after each change.
5. Connect to both instances through their respective Overlay IPs on port 22. Then connect over VPN using each instance's external SSH port.
6. Create a uniquely named file in the clone. Restart only the clone and check both systems again.

| ID | Expected result |
| --- | --- |
| QEMU-11 | The original remains Ubuntu and the clone runs Debian. Their machine IDs differ. The clone is a fresh installation and does not contain files copied from the original disk. |
| QEMU-12 | Both instances remain listed in Overlay gateway with different IP addresses. Each accepts SSH on port 22. Enabling the second does not remove the first. |
| QEMU-13 | VPN connections use the same Olares device address and different external SSH ports. Each session reaches the correct distribution. |
| QEMU-14 | Files persist in their own instance. Restarting or changing resources for the clone does not change the original's distribution, configuration, or files. |

## Recorded validation coverage

The recorded regression for QEMU Server 0.0.27 used Olares 1.12.7 on amd64. The following results describe that run; use the cases above to validate your own environment.

| Distribution | Filesystem | Root capacity before → after expansion | Recorded result |
| --- | --- | --- | --- |
| Ubuntu 24.04 LTS | ext4 | 18.33 → 22.21 GiB | Installation, authenticated internal SSH, console transport, resource changes, and persistence passed. |
| Debian 13 | ext4 | 19.49 → 23.42 GiB | The same checks passed; interrupted download recovery also passed. |
| Rocky Linux 10 | XFS | 18.38 → 22.38 GiB | Installation, authenticated internal SSH, console transport, resource changes, and persistence passed. |
| Fedora 44 | Btrfs | 19.90 → 23.90 GiB | The same checks passed. First boot reported a recoverable hostname warning; subsequent checks confirmed the hostname and no fatal initialization error. |

The automated console check verified its transport and handshake, rather than a complete interactive browser login. SSH authentication was checked through the internal application network. External LAN and VPN SSH, cross-platform SSH clients, and two-instance Clone/Overlay behavior were not completed in this regression. Treat QEMU-04–06 and QEMU-11–14 as pending for this version until you execute them. Native ARM64 and the other listed distribution releases require separate verification.

For each case, record pass, fail, or not run, together with the observed result. Retain screenshots or terminal output for failures, with credentials and personal addresses removed. Do not infer coverage for another architecture or distribution release from one successful installation.
