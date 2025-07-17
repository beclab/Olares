---
outline: [2, 3]
description: Learn how to run Olares as a containerized application on Mac with Docker, covering image setup and container configuration.
---
# Install Olares on Mac with Docker image
You can use Docker to install and run Olares in a containerized environment. This guide walks you through setting up Olares with Docker, preparing the installation environment, completing the activation process, and managing the container lifecycle.

:::warning Not for production use
Currently, Olares on Mac has certain limitations including:
- Lack of distributed storage support.
- Inability to add local nodes.

We recommend using it only for development or testing purposes.
:::

<!--@include: ./reusables.md{41,47}-->

## System requirements
Make sure your device meets the following requirements.

- Architecture: AMD64 or ARM64
- CPU: At least 4 cores
- RAM: At least 8GB of available memory
- Storage: At least 64GB of available space (SSD recommended)

## Before you begin
Before you begin, ensure the following:
- [Docker](https://www.docker.com/) is installed and running on your system.
- You know the IP address of the current device.
  ::: tip View IP Address
  To view the IP address on a Mac, there are two methods:
   - Using the graphical interface: Open **System Settings** (or **System Preferences**) > **Network**, and check the details under the currently active network connection.
   - Using the command line: Open a terminal window and enter `ipconfig getifaddr en0` for Wi-Fi, or `ipconfig getifaddr en1` for wired network.
     :::
- You have [created an Olares ID via LarePass](/manual/get-started/create-olares-id.md).

## Run `olaresd-proxy`
::: tip Check Mac chip  
If you are unsure which chip your Mac is using, go to the Apple menu and select **About This Mac** to verify.
:::
<tabs>
<template #Apple-Silicon>

1. Download `olaresd-proxy` via the link: https://dc3p1870nn3cj.cloudfront.net/olaresd-proxy-v0.1.0-darwin-arm64.tar.gz .
2. Unzip the file, then start `olaresd-proxy`.
   :::info Keep `olaresd-proxy` running in the background
   During Olares installation and activation, keep `olaresd-proxy` running in the background.
   :::
</template>

<template #Intel>

1. Download `olaresd-proxy` via the link: https://dc3p1870nn3cj.cloudfront.net/olaresd-proxy-v0.1.0-darwin-amd64.tar.gz .
2. Unzip the file, then start `olaresd-proxy`.
   :::info Keep `olaresd-proxy` running in the background
   During Olares installation and activation, keep `olaresd-proxy` running in the background.
   :::
</template>
</tabs>

## Pull the Olares image

To pull the image of Olares, execute the following command.

Replace `<host ip>` with your device's IP address and `<olares version>` with the desired version of Olares:
```bash{2,9}
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
:::
:::warning Do not add the `--rm` flag
The `--rm` flag automatically deletes the container after it stops. If this happens, you will not be able to restart the container and will need to reinstall Olares to run it again. Omitting this flag preserves the container after stoppage, enabling you to resume it with the`docker start` command.
:::

<!--@include: ./install-and-activate-olares.md-->

<!--@include: ./manage-olares-container.md-->

<!--@include: ./reusables.md{35,39}-->
