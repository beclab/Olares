---
outline: [2, 3]
description: Find answers to common questions during the installation and activation of Olares.
---
# Olares installation, activation & login FAQs

This page lists the FAQs about installing, configuring, activating, and logging into Olares on your hardware.

## Installation

### What platforms does Olares support?

Install Olares on Linux (Ubuntu or Debian) for best performance.

For product evaluation, install Olares on the following platforms:
* Proxmox VE
* Raspberry Pi
* macOS
* Windows

### What are the minimum hardware requirements for installing Olares?

Requirements vary by platform. Generally, you need:
* **CPU**: Minimum 4 cores with x86-64 architecture, such as Intel or AMD.
* **Memory**: At least 8 GB of available RAM.
* **Storage**: Minimum 150 GB SSD.

For detailed requirements, refer to the [installation docs](../get-started/install-olares.md).

### Is it possible to use a mechanical hard drive to install Olares?

No. You must use an SSD. Installations on mechanical hard drives likely fail due to slower read and write speeds, which cause timeouts during system initialization.

### Does the system support NVIDIA GPUs?

Yes, but only NVIDIA Turing architecture or newer, such as GTX 16xx, RTX 20xx, 30xx, 40xx, 50xx series, and later. GPUs with older architectures are not recognized by Olares, and AI applications that require GPU access will not run.

This requirement stems from the NVIDIA open-source driver requiring the hardware GSP module of Turing, and CUDA 13.x dropping support for older GPU architectures.

For supported GPUs, Olares automatically handles driver installation. It supports multiple GPUs on a single motherboard, allowing you to leverage all available compute power for AI workloads.

### How do I manually install NVIDIA drivers if automatic setup fails?

The Olares installer detects and installs drivers automatically. However, if your system previously had NVIDIA drivers installed, the process might be skip or fail due to conflicts.

To resolve this:
1. Reboot the machine after the Olares installation to clear any old driver components.
2. Manually trigger the driver installation using the command `olares-cli gpu install`.
3. After installation, confirm the system recognizes your GPU by running `nvidia-smi`.

### Why does installation fail with `failed to build Kubernetes objects` or `Ensure CRDs are installed first`?

While these error messages suggest a problem with Custom Resource Definitions (CRDs), they often indicate poor disk performance.

Olares relies on etcd, the backing database for Kubernetes. etcd is very sensitive to storage speed. If you install Olares on a slow disk, such as a traditional HDD, etcd cannot respond fast enough. This causes the API server to time out while attempting to apply CRDs.

Install Olares on SSD storage to fix this issue.

### How do I find the password if the Olares installation times out without showing it?

This typically occurs when the installation times out due to insufficient system resources, especially in a virtual machine (VM). You can retrieve the password from the installation log file with the following command:

```bash
# Replace v1.12.2 with your specific Olares version number.
grep password $HOME/.olares/versions/v1.12.2/logs/install.log
```
An installation timeout often indicates some services failed to start correctly. After you find your password, run `kubectl get pod -A` to check the status of all services.

## Activation

### Is it possible to activate Olares with a non-local network?

Yes. Before activation, users typically access the activation wizard using the local IP address, which requires the device and the client to be on the same network. However, if Olares uses a public IP, such as on a public cloud, this local network limitation no longer applies.

Note that IP access is only used during activation. After activation, access devices via domain names on both internal and external networks.

### My Olares is powered on and connected to LAN, but I can't find it in LarePass. What should I do?

Ensure your phone and Olares device are on the same network. If they are not, LarePass cannot discover Olares automatically.

If you cannot connect via Wi-Fi, use the Bluetooth network setup in the LarePass app to connect Olares to the same network as your phone.

For details, see [Activate Olares using Bluetooth](../../manual/larepass/activate-olares.md#activate-olares-using-bluetooth).

### I received the MFA binding error. How to resolve it?

The request to bind Multi-Factor Authentication (MFA) failed due to a network timeout. Check your network connection and retry the setup process.

### I received a DID binding error. How to resolve it?

The request to the binding-zone interface failed due to a network timeout. This typically happens during account activation. Ensure your network is stable and retry.

### What does this error message mean: "Invalid jws, timestamp is out of range"?

This occurs when the time difference between the device running LarePass and the Olares device exceeds 20 minutes. Check the system time on your computer or phone and ensure it syncs correctly with internet time.

### I see this error message during setup: "Resolve name error". How to fix it?

This error indicates your Olares device cannot connect to the Olares identity service, which is usually caused by the network issue. Check your internet connection and ensure your Olares device has proper network access.

## Login and authentication

### What should I do when I see "Authentication failed, incorrect password"?

This indicates the password entered is not accurate. Check your password for typos or case sensitivity issues, and try again.

### What does this mean: "Authentication failed, user not found"?

The system cannot find an account associated with the entered username. Verify the username is correct and already registered on Olares.

### I received an error saying "Authentication failed, failed to query user from lldap service". How to resolve it?

The system was unable to retrieve user details from the internal directory. This happens if the user data corrupts or if the service is momentarily busy. Try again or contact your administrator.

### I am locked out with the message "too many failed login attempts, retry again later after 5 minutes". What should I do?

For security reasons, the system temporarily locks logins after multiple incorrect attempts. Wait 5 minutes, and then try logging in again.

### Why am I getting "Authentication failed, disk space is full"?

Your Olares device storage is completely full, which prevents the authentication service from running correctly. Free up disk space on the device to resolve this issue.

### What does this error message mean: "Authentication failed, lldap service is unavailable"?

The internal identity service "LLDAP" on your Olares device is not running or is experiencing issues. Restart the Olares device and try again, or contact your administrator.

### What does this error message mean: "Authentication failed, citus service is unavailable"?

The internal database service "Citus" is currently unavailable. This is usually a temporary system issue. Restart your Olares device.
