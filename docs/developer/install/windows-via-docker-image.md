---
outline: [2, 3]
description: Learn how to run Olares on Windows using WSL 2 and Docker, including system preparation, configuration, and container management.
---
# Install Olares on Windows (WSL 2) with Docker image
You can use Docker to install and run Olares in a containerized environment. This guide walks you through setting up Olares with Docker and WSL 2, preparing the installation environment, completing the activation process, and managing the container lifecycle.

:::warning Not recommended for production use
Currently, Olares on Windows has certain limitations including:
- Lack of distributed storage support
- Inability to add local nodes.

We recommend using it only for development or testing purposes.
:::
<!--@include: ./reusables.md{41,47}-->

## System requirements
Make sure your Windows meets the following requirements.
- CPU: At least 4 cores
- RAM: At least 16GB of available memory
- Storage: At least 64GB of available space (SSD recommended)
- Supported systems:
    - Windows 10 or 11
    - Linux (on WSL 2): Ubuntu 20.04 LTS or later; Debian 11 or later

## Before you begin
Before you begin, ensure the following:
- [Docker Desktop](https://docs.docker.com/desktop/setup/install/windows-install/)  is installed and running on your system.
  :::info WSL 2 and Hyper-V  
  If Docker Desktop is configured to use **Hyper-V**, GPU support for Olares cannot be enabled. Ensure Docker Desktop is set to run in the **WSL 2** mode.
  :::
- You know the IP address of the current device.
  ::: tip View IP Address
  In PowerShell or Command Prompt, use the following command to confirm your IP address:
  ```bash
  ipconfig | findstr /i "IPv4.*192"
  ```
    :::
- You have [created an Olares ID via LarePass](/manual/get-started/create-olares-id.md).

## Configure WSL 2
1. Open PowerShell and run the following command to confirm the kernel version of WSL installed on your system:
   ```powershell
   wsl --version
   ```
   Example output:
   ```PowerShell{2}
   WSL version: 2.4.8.0
   Kernel version: 5.15.167.4-1
   WSLg version: 1.0.65
   MSRDC version: 1.611.1-81528511
   DXCore version: 10.0.26100.1-240331-1435.ge-release
   Windows version: 10.0.26100.3475
   ```
2. Use the link provided to download the appropriate kernel file matching your version: `https://dc3p1870nn3cj.cloudfront.net/bzImage-<kernel-version>`.
   For example, For kernel version `5.15.167.4`, download the file from `https://dc3p1870nn3cj.cloudfront.net/bzImage-5.15.167.4`.

   Supported kernel versions (above `5.15.146.1`) are:
   -  `linux-msft-wsl-5.15.146.1`
   -  `linux-msft-wsl-5.15.150.1`
   -  `linux-msft-wsl-5.15.153.1`
   -  `linux-msft-wsl-5.15.167.4`
   -  `linux-msft-wsl-6.6.75.1`
   -  `linux-msft-wsl-6.6.36.6`
   -  `linux-msft-wsl-6.6.36.3`
3. Set the default version of WSL to version 2:
  ```bash
  wsl --set-default-version 2
  ```
4. In the directory `C:\Users\<YourUsername>\`, create a file named `.wslconfig` with the following content:
   ```txt
   [wsl2]
   kernel=c:\\path\\to\\your\\kernel\\bzImage-<version> # Note: Use double backslashes (\\) as path separators
   memory=8GB # Recommended: 16GB
   swap=0GB
   ```
   :::info
   If you installed Docker Desktop before modifying the `.wslconfig` file, it is recommended to remove the `docker-desktop` distribution installed under WSL:
   ```bash
   wsl --unregister docker-desktop
   wsl --unregister docker-desktop-data # If this version exists
   ```
   :::
5. Restart your computer to apply the changes.

## Prepare Docker
If you have installed Docker Desktop before modifying `.wslconfig`, remove docker desktop then restart Windows.
   
## Run `olaresd-proxy`
1. Download `olaresd-proxy`from the following link: `https://dc3p1870nn3cj.cloudfront.net/olaresd-proxy-v0.1.0-windows-amd64.tar.gz`.
2. Extract the file and start the `olaresd-proxy` executable.
   :::info Keep `olaresd-proxy` running
   Ensure that `olaresd-proxy` runs in the background during the installation and activation of Olares.
   :::

## Run Olares using the Docker CLI
:::warning CUDA version requirements
CUDA version 12.4 or above is required for GPU support. Older versions are incompatible.
:::
Run the following command to pull the Olares image.
Replace `<host ip>` with your device's IP address and `<olares version>` with the desired version of Olares:
::: code-group
```bash{2,9} [Without GPU support]
docker run -d --privileged -v oic-data:/var \
  -e HOST_IP=<host ip> \
  -p 80:80 \
  -p 443:443 \
  -p 30180:30180 \
  -p 18088:18088 \
  -p 41641:41641/udp \
  --name oic \
 beclab/olares:<olares version>
```
```bash{1,2,9} [With GPU support]
docker run --gpus all -d --privileged -v oic-data:/var \
 -e HOST_IP=<host ip> \
 -p 80:80 \
 -p 443:443 \
 -p 30180:30180 \
 -p 18088:18088 \
 -p 41641:41641/udp \
 --name oic \
 beclab/olares:<olares version>
```
:::
where:
  - `-d`: Starts the container in detached mode to allow it to run in the background.
  - `--privileged`: Grants the container elevated privileges.
  - `-v oic-data:/var`: Binds a Docker volume (`oic-data`) to the `/var` directory inside the container to persist data.
  - `-e HOST_IP=<host ip>`: Specifies the host device's IP address as an environment variable.
  - `-p 80:80`: Maps port `80` on the host to port `80` in the container.
  - `-p 443:443`: Maps port `443` on the host to port `443` in the container.
  - `-p 30180:30180`: Maps port `30180` on the host to port `30180` in the container.
  - `-p 18088:18088`: Maps port `18088` on the host to port `18088` in the container.
  - `-p 41641:41641/udp`: Maps UDP port `41641` on the host to UDP port `41641` in the container.
  - `--name oic`: Names the container `oic` (Olares in container) for easier reference.
  - `beclab/olares:<olares version>`: Specifies the Olares Docker image and version. For example: `beclab/olares:1.11.5`.

When the container is running, you will see a container ID output.

:::warning Do not add the `--rm` flag
The `--rm` flag automatically deletes the container after it stops. If this happens, you will not be able to restart the container and will need to reinstall Olares to run it again. Omitting this flag preserves the container after stoppage, enabling you to resume it with the`docker start` command.
:::

<!--@include: ./install-and-activate-olares.md-->

<!--@include: ./manage-olares-container.md-->

<!--@include: ./reusables.md{35,39}-->
