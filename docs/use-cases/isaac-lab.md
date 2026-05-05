---
outline: [2, 3]
description: Run NVIDIA Isaac Lab on Olares for GPU-accelerated robot simulation training, with WebRTC streaming to visualize training progress in real time.
head:
  - - meta
    - name: keywords
      content: Olares, Isaac Lab, Isaac Sim, NVIDIA, robot simulation, reinforcement learning, GPU, WebRTC streaming, self-hosted
app_version: "1.0.3"
doc_version: "1.0"
doc_updated: "2026-04-13"
---

# Train robot simulations with Isaac Lab

Isaac Lab is NVIDIA's open-source, GPU-accelerated framework for robot learning. Built on Isaac Sim, it provides fast and accurate physics simulation for reinforcement learning, imitation learning, and motion planning workflows. Running Isaac Lab on Olares gives you a dedicated environment for robot training with optional real-time visualization through WebRTC streaming.

## Prerequisites

- Olares running on a machine with an NVIDIA GPU.
- For real-time streaming and visualization, an AMD64 (x86_64) architecture machine is required. ARM64 machines (such as DGX Spark) can only run training in headless mode without streaming.
- NVIDIA WebRTC Streaming Client installed on your local computer (AMD64 only). Download it from [NVIDIA's official site](https://docs.isaacsim.omniverse.nvidia.com/5.1.0/installation/download.html).

:::warning Dedicate your GPU to Isaac Lab
Isaac Lab requires significant GPU resources. To avoid affecting other GPU apps, navigate to **Settings** > **General** and set the GPU mode to **Exclusive** for Isaac Lab. This ensures the GPU is fully dedicated to this app.
:::

## Install Isaac Lab

1. Open Market and search for "IsaacLab".
   <!-- ![Isaac Lab](/images/manual/use-cases/isaac-lab.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Run Isaac Lab training
:::info ARM64 users
On ARM64 machines, streaming is not supported. Remove `LIVESTREAM=1` and `--headless` from the commands to run training in headless mode without visualization.
:::

The following example runs a quadruped locomotion demo:
1. In the Isaac Lab terminal, start the training:

```bash
LIVESTREAM=1 nohup ./isaaclab.sh -p scripts/demos/quadrupeds.py --headless &
```

2. In the same terminal, run `echo $PUBLIC_IP` to get the server address.


3. Once the training reaches a stable state, connect the WebRTC Streaming Client:

   a. Open the WebRTC Streaming Client on your local computer and enter the `PUBLIC_IP` value as the server address.

   b. Click **Connect**. Only one client can connect at a time.


:::tip Find more scripts and tasks
For a full list of available training scripts, tasks, and environment variables, see:
- [Existing RL scripts](https://isaac-sim.github.io/IsaacLab/main/source/overview/reinforcement-learning/rl_existing_scripts.html)
- [Environments](https://isaac-sim.github.io/IsaacLab/main/source/overview/environments.html)
- [Newton physics integration](https://isaac-sim.github.io/IsaacLab/main/source/experimental-features/newton-physics-integration/training-environments.html)
:::

## Run Isaac Sim
:::info ARM64 users
On ARM64 machines, streaming is not available.
:::

Isaac Lab includes a bundled Isaac Sim simulator for interactive scene exploration and testing. 

Start it from the terminal:

1. In the terminal, run the following:

  ```bash
  nohup ./_isaac_sim/runheadless.sh --/app/livestream/publicEndpointAddress=$PUBLIC_IP --/app/livestream/port=49100 &
  ```

   When the terminal log shows `Isaac Sim Full Streaming App is loaded.`, the simulator is ready for WebRTC connections.
2. Connect the WebRTC Streaming Client using the same [steps above](#run-isaac-lab-training).

  :::info Loading complex scenes
  When loading a complex scene, the streaming view might go grey. Wait for the terminal logs to stabilize (this can take one to several minutes depending on the scene complexity), then click **Reload** in the WebRTC Streaming Client to reconnect.
  :::

## Terminal reference

Isaac Lab opens to a terminal interface. Here are some common commands you will use:

| Command | Description |
|:--------|:------------|
| `echo $PUBLIC_IP` | Display the public IP address for WebRTC streaming |
| `nvidia-smi` | Check GPU status and list running processes |
| `kill -9 <pid>` | Stop a background process by its PID |
| `tail -f nohup.out` | Follow logs for the currently running process |

## Learn more

- [Isaac Lab documentation](https://isaac-sim.github.io/IsaacLab/main/index.html): Official framework docs with API reference and tutorials.
- [Isaac Sim documentation](https://docs.isaacsim.omniverse.nvidia.com/5.1.0/index.html): NVIDIA Isaac Sim user guide.
