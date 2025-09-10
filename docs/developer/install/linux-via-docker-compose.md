---
outline: [2, 3]
description: Learn how to deploy Olares on a Linux server using Docker Compose. This step-by-step guide covers system requirements, configuration, installation, activation, and container management.
---
# Install Olares on Linux using Docker Compose
You can use Docker to install and run Olares in a containerized environment. This guide walks you through setting up Olares with Docker, preparing the installation environment, completing the activation process, and managing the container lifecycle.

:::tip Recommendation for production use
For best performance and stability, we recommend [installing Olares on Linux via script](/manual/get-started/install-olares.md).
:::

## System requirements

Make sure your device meets the following requirements.

- CPU: At least 4 cores
- RAM: At least 8GB of available memory
- Storage: At least 150GB of available SSD storage. 
- Supported systems:
    - Ubuntu 22.04 LTS or later
    - Debian 12 or later

::: warning SSD required
The installation will likely fail if an HDD (mechanical hard drive) is used instead of an SSD.
:::

:::info Version compatibility
While these specific versions are confirmed to work, the process may still work on other versions. Adjustments may be necessary depending on your environment. If you meet any issues with these platforms, feel free to raise an issue on [GitHub](https://github.com/beclab/Olares/issues/new).
:::

## Before you begin
Before you begin, ensure the following:
- [Docker](https://docs.docker.com/engine/install/) and [Docker Compose](https://docs.docker.com/compose/install/) are installed and running on your system.
- You know the IP address of the current device.
  :::tip Verify host IP
  To verify your host IP, run the following command in the terminal:
  ```bash
  ip r
  ```
  Look for the line starting with `default via`. It will show the default gateway and the network interface being used.
  :::
- You have [created an Olares ID via LarePass](/manual/get-started/create-olares-id.md).

## Create a new directory
Create a directory to store the Olares configuration files. For example, you could make a new directory called `olares-config` with the following command:

```bash
mkdir ~/olares-config
cd ~/olares-config
```
## Prepare `docker-compose.yaml`
1. Create a `docker-compose.yaml` file in the `olares-config` directory.
2. Add the appropriate content to the file based on whether GPU support is required:
   :::code-group
   <<< @/code-snippets/docker-compose.yaml
   <<< @/code-snippets/docker-compose-GPU.yaml
   :::
3. Save the `docker-compose.yaml` file.

Here is the updated and properly formatted Markdown content for installing GPU drivers and the NVIDIA Container Toolkit, ready to be included in your installation guide:


## Install GPU dependencies (for GPU-enabled machines)

1. Install GPU drivers for your system:

    ```bash
    curl -o /tmp/keyring.deb -L https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2204/x86_64/cuda-keyring_1.1-1_all.deb && \
    sudo dpkg -i --force-all /tmp/keyring.deb
    
    sudo apt update
    sudo apt install nvidia-kernel-open-570
    sudo apt install nvidia-driver-570
    ````

2. Install the NVIDIA Container Toolkit to enable Docker to access your GPU. 
     
     a. Configure the repository:

    ```bash
    curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey | \
      sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg
    
    curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list | \
      sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' | \
      sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list
    
    sudo sed -i -e '/experimental/ s/^#//g' /etc/apt/sources.list.d/nvidia-container-toolkit.list
    
    sudo apt-get update
    ```

      b. Install the toolkit and restart docker:

   ```bash
   sudo apt-get install -y nvidia-container-toolkit
   sudo nvidia-ctk runtime configure --runtime=docker
   sudo systemctl restart docker
   ```

      c. Verify the installation:

   ```bash
   sudo docker run --rm --runtime=nvidia --gpus all ubuntu nvidia-smi
   ```

    If successful, you should see output similar to the following:

    ```
    +-----------------------------------------------------------------------------------------+
    | NVIDIA-SMI 570.169                Driver Version: 570.169        CUDA Version: 12.8     |
    |-----------------------------------------+------------------------+----------------------+
    | GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |
    | Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
    |                                         |                        |               MIG M. |
    |=========================================+========================+======================|
    |   0  NVIDIA GeForce RTX 4070 ...    Off |   00000000:01:00.0 Off |                  N/A |
    | N/A   41C    P8              1W /   80W |      32MiB /   8188MiB |      0%      Default |
    |                                         |                        |                  N/A |
    +-----------------------------------------+------------------------+----------------------+
    ```


## Set up environment variables and start container

1. In the `olares-config` directory, use the following command to set the environment variables and start the Olares services:

   ```bash [With Docker Compose Plugin]
   VERSION=<olares version> HOST_IP=<host ip> docker compose up -d
   ```
   - `VERSION=<olares version>`: Specifies the Olares version. Replace `<olares version>` with the actual one. For example: `1.11.5`.
   - `HOST_IP=<host ip>`: Specifies the Linux machine's IP address. Replace `<host ip>` with the actual one.
   
   After executing the command, you should see output similar to the following, showing the status and port mappings of all containers:
    ```bash
   [+] Running 20/20
   ✔ olaresd-proxy Pulled                                                                           67.8s
   ✔ 688513194d7a Pull complete                                                                    6.8s
   ✔ bfb59b82a9b6 Pull complete                                                                    6.9s
   ✔ efa9d1d5d3a2 Pull complete                                                                    9.5s
   ✔ a62778643d56 Pull complete                                                                    9.6s
   ✔ 7c12895b777b Pull complete                                                                    9.6s
   ✔ 3214acf345c0 Pull complete                                                                   13.6s
   ✔ 5664b15f108b Pull complete                                                                   14.1s
   ✔ 0bab15eea81d Pull complete                                                                   14.2s
   ✔ 4aa0ea1413d3 Pull complete                                                                   15.0s
   ✔ da7816fa955e Pull complete                                                                   15.1s
   ✔ 9aee425378d2 Pull complete                                                                   15.1s
   ✔ 701c983262e9 Pull complete                                                                   36.2s
   ✔ 221438ca359c Pull complete                                                                   36.3s
   ✔ f3d0ed3b32e0 Pull complete                                                                   36.4s
   ✔ 70d5c1f325f6 Pull complete                                                                   43.2s
   ✔ olares Pulled                                                                                5863.6s
   ✔ 2d5815038f40 Pull complete                                                                 5759.0s
   ✔ 13788179ee16 Pull complete                                                                 5831.6s
   ✔ 5a9b10c3302f Pull complete                                                                 5831.7s
    ```

2. Verify if the container is running successfully:
   ```bash
   docker ps
   ```
   You should see an output like this:
   ```bash
   CONTAINER ID   IMAGE                         COMMAND                  CREATED              STATUS              PORTS                   NAMES
   28e86c473750   beclab/olaresd:proxy-v0.1.0   "/mdns-agent"            About a minute ago   Up About a minute                           olares-olaresd-proxy-1
   5fd68a8709ad   beclab/olares:1.11.5       "/usr/local/bin/entr…"   2 minutes ago        Up About a minute   0.0.0.0:80->80/tcp...   olares-olares-1
   ```

<!--@include: ./install-and-activate-olares.md-->

## Manage the Olares container
Ensure that you are in the directory containing the `docker-compose.yaml` file before proceeding with any commands.
### Stop the container
To stop the running container:
```bash
docker compose stop
```

### Restart the container
To restart the container after it has been stopped:
```bash
docker compose start
```
It may take 6 to 7 minutes for all services to fully initialize after restarting.

### Uninstall the container
To uninstall the container:
```bash
docker compose down
```

<!--@include: ./reusables.md{39,43}-->
   
   

