---
outline: [2, 3]
description: Run NVIDIA Isaac Lab on Olares for GPU-accelerated robot simulation training, with WebRTC streaming to visualize training progress in real time.
head:
  - - meta
    - name: keywords
      content: Olares, Isaac Lab, Isaac Sim, NVIDIA, robot simulation, reinforcement learning, GPU, WebRTC streaming, self-hosted
app_version: "1.0.18"
doc_version: "1.1"
doc_updated: "2026-10-10"
---

# Run robot simulations with Isaac Lab

Isaac Lab is NVIDIA's open-source framework for robot learning. On Olares, it runs as a terminal-based app where you start demos, training tasks, or the bundled Isaac Sim simulator from the command line.

Use this guide to install Isaac Lab, run Isaac Lab workloads, start Isaac Sim, connect to WebRTC streaming, and manage background processes.

## Learning objectives

In this guide, you will learn how to:

- Install Isaac Lab on Olares.
- Prepare the Isaac Lab terminal and get the WebRTC endpoint address.
- Run Isaac Lab demos and reinforcement learning tasks.
- Start the bundled Isaac Sim simulator.
- Connect to simulations with the NVIDIA WebRTC Streaming Client.
- Manage background workloads, GPU processes, and runtime logs.

## Prerequisites

Before you begin, make sure:
- Olares is running on a machine with an NVIDIA GPU.
- The Olares host uses an AMD64 architecture if you want to use WebRTC streaming. ARM64 hosts (such as DGX Spark) support only non-streaming workloads.
- [NVIDIA's WebRTC Streaming Client](https://docs.isaacsim.omniverse.nvidia.com/6.1.0/installation/download.html) is installed on your local computer.

:::info Host vs. local computer architecture
The AMD64 requirement applies to the Olares host where Isaac Lab runs, not to the local computer used for streaming.

For example, if Isaac Lab runs on an Olares One with an Intel CPU and NVIDIA GPU, the Olares host is AMD64. Your local computer can use a different architecture, but it must be able to run NVIDIA's WebRTC Streaming Client.
:::

:::tip Recommended GPU mode
Isaac Lab and Isaac Sim use significant GPU resources. For better stability, set the GPU mode to [**Exclusive**](/manual/olares/settings/gpu-resource.md#change-a-gpu-mode) and bind the GPU to Isaac Lab before running workloads. This helps reduce resource contention with other GPU apps.
:::

## Understand Isaac Lab and Isaac Sim

Isaac Lab and Isaac Sim serve different purposes:

| Component | Description | Use it when |
|:--|:---|:--|
| Isaac Lab | A robot learning framework built<br> on Isaac Sim. | Run predefined demos, reinforcement learning tasks, or training scripts. |
| Isaac Sim | The underlying simulator bundled <br>with Isaac Lab. | Start the simulator directly for scene exploration, simulation testing, or interactive use. |

- To run an Isaac Lab demo or training task, use the commands in [Run Isaac Lab workloads](#run-isaac-lab-workloads).
- To start the Isaac Sim simulator itself, use [Run Isaac Sim](#run-isaac-sim).

## Install Isaac Lab

1. Open Market and search for "IsaacLab".
   
   ![Isaac Lab](/images/manual/use-cases/isaac-lab.png#bordered){width=90%}

2. Click **Get**, then **Install**, and wait for installation to complete.

## Prepare Isaac Lab terminal

Whenever you open Isaac Lab or want to switch to a new task, prepare your terminal environment first.

1. Open Isaac Lab from Launchpad to access its terminal interface for running command-line scripts.

   ![Isaac Lab terminal](/images/manual/use-cases/isaac-lab-terminal.png#bordered)

2. Get the public endpoint address used for WebRTC streaming.

   ```bash
   echo $PUBLIC_IP
   ```
   Example output:
   ```plain
   192.168.50.169
   ```
   Save the output value. You will need it later to connect the WebRTC Streaming Client.

3. Check the GPU status and look for existing workloads.

   ```bash
   nvidia-smi
   ```
   - If no workload is running, the `Processes` section shows `No running processes found`.

   ![No running process](/images/manual/use-cases/isaac-lab-no-running-process.png#bordered){width=90%}

   - If the `Processes` section lists an Isaac Lab or Isaac Sim process, note its `PID` before starting a new workload. In this example, the active process uses PID `2314`.

   ![Active process](/images/manual/use-cases/isaac-lab-active-process.png#bordered){width=90%}

4. If an existing workload is running, stop it before starting a new one:

   ```bash
   kill -9 <PID>
   ```

   Replace `<PID>` with the actual process ID shown in your terminal.

## Run Isaac Lab workloads

Isaac Lab workloads are started from the terminal. The command you use depends on the architecture of the Olares host where Isaac Lab is installed.

Stop any existing workload before starting a new demo or task. Run one command at a time, in the foreground. Official Isaac Lab docs recommend `uv run isaaclab`. `./isaaclab.sh` still accepts the same arguments.

### Run a demo

<Tabs>
<template #AMD64-host>

Use this command when Isaac Lab runs on an AMD64 Olares host and you want to enable WebRTC streaming.

1. Start the quadruped locomotion demo:

   ```bash
   uv run isaaclab -p scripts/demos/quadrupeds.py --livestream 1
   ```

2. Wait until the workload starts and the logs in this terminal become stable.

   Warnings from Isaac Sim, Omniverse, or PyTorch may appear during startup. If no fatal error appears and the process remains running, continue to the next step.

3. Open the WebRTC Streaming Client on your local computer, enter the `PUBLIC_IP` value, then click **Connect**.

   ![Connect to Streaming client](/images/manual/use-cases/isaac-lab-connect-streaming-client.png#bordered){width=90%}

4. View the simulation progress.

   ![View simulation progress](/images/manual/use-cases/isaac-lab-view-simulation.png#bordered){width=90%}
</template>

<template #ARM64-host>

Use this command when Isaac Lab runs on an ARM64 Olares host, such as DGX Spark.

On ARM64 Olares hosts, WebRTC streaming is not supported. Do not pass `--livestream`.

```bash
uv run isaaclab -p scripts/demos/quadrupeds.py
```

Because WebRTC streaming is not available on ARM64 Olares hosts, you cannot visualize the demo through the WebRTC Streaming Client.

</template>
</Tabs>

### Run a reinforcement learning task

<Tabs>
<template #AMD64-host>

Use this command when Isaac Lab runs on an AMD64 Olares host and you want to enable WebRTC streaming.

1. Start the reinforcement learning training task:

   ```bash
   uv run isaaclab train --rl_library rsl_rl --task Isaac-Velocity-Rough-H1 --livestream 1
   ```

2. Wait until the workload starts and the logs in this terminal become stable.

   Warnings from Isaac Sim, Omniverse, or PyTorch may appear during startup. If no fatal error appears and the process remains running, continue to the next step.

3. Open the WebRTC Streaming Client on your local computer, enter the `PUBLIC_IP` value, then click **Connect**.

   ![Connect to Streaming client](/images/manual/use-cases/isaac-lab-connect-streaming-client.png#bordered){width=90%}

4. View the training progress.
   
   ![View RL training progress](/images/manual/use-cases/isaac-lab-view-RL-training.png#bordered){width=90%}

</template>

<template #ARM64-host>

Use this command when Isaac Lab runs on an ARM64 Olares host, such as DGX Spark.

On ARM64 Olares hosts, WebRTC streaming is not supported. Do not pass `--livestream`.

```bash
uv run isaaclab train --rl_library rsl_rl --task Isaac-Velocity-Rough-H1
```

Because WebRTC streaming is not available on ARM64 Olares hosts, you cannot visualize the training process through the WebRTC Streaming Client.

</template>
</Tabs>

For the training command used on this page, refer to:

- [Training with an RL Agent](https://isaac-sim.github.io/IsaacLab/release/3.0.0/source/how-to/run_rl_training.html)

## Run Isaac Sim

Isaac Lab includes the Isaac Sim simulator. Start Isaac Sim directly when you want to use the simulator itself instead of running a specific Isaac Lab demo or training script.

Stop any existing workload before starting Isaac Sim.

<Tabs>
<template #AMD64-host>

Use this command when Isaac Sim runs on an AMD64 Olares host and you want to enable WebRTC streaming.

1. Start Isaac Sim:

   ```bash
   /isaac-sim/runheadless.sh
   ```

2. Wait until this terminal shows that the Isaac Sim streaming app is loaded.

   For example:

   ```text
   Isaac Sim Full Streaming App is loaded.
   ```

3. Open the WebRTC Streaming Client on your local computer, enter the `PUBLIC_IP` value, then click **Connect**.

   ![Connect to Streaming client](/images/manual/use-cases/isaac-lab-connect-streaming-client.png#bordered){width=90%}

If the streaming view turns gray, see [The streaming view turns gray](#the-streaming-view-turns-gray).

</template>

<template #ARM64-host>

Use this command when Isaac Sim runs on an ARM64 Olares host, such as DGX Spark.

On ARM64 Olares hosts, WebRTC streaming is not supported. Do not use `runheadless.sh` for streaming.

Start Isaac Sim in non-graphical mode:

```bash
./_isaac_sim/runapp.sh
```

Because graphical streaming is not available on ARM64 Olares hosts, this mode is mainly for limited non-graphical usage.

</template>
</Tabs>

## Terminal command reference

Commands run in the foreground. Closing the WebRTC Streaming Client does not stop the workload. Press **Ctrl + C** in the terminal to stop it.

Use these commands for quick terminal management and troubleshooting:

| Task | Command |
|:-----|:--------|
| Get the WebRTC endpoint address | `echo $PUBLIC_IP` |
| Check GPU processes | `nvidia-smi` |
| Stop a leftover process | `kill -9 <PID>` |

## Troubleshooting

### WebRTC Client cannot connect

Check the following:

1. Make sure the workload is still running:

   ```bash
   nvidia-smi
   ```

2. Confirm that you entered the correct `PUBLIC_IP` value:

   ```bash
   echo $PUBLIC_IP
   ```

3. Make sure only one WebRTC Streaming Client is connected.

4. Read the terminal where the workload is running, and look for errors such as:

```text
Error
Traceback
CUDA out of memory
Segmentation fault
Killed
```

If none of these errors appear and the process is still running, wait for the logs to stabilize, then reconnect or click **Reload** in the WebRTC Streaming Client.

### The streaming view turns gray

When loading a complex scene, the streaming view might turn gray or stop responding temporarily. Wait for the terminal logs to stabilize. Depending on the scene complexity, this may take one to several minutes.

After the logs become stable, click **View** > **Reload** in the WebRTC Streaming Client to reconnect to the stream.

![Reload WebRTC stream](/images/manual/use-cases/isaac-lab-reload-WebRTC-stream.png#bordered){width=90%}

### Logs show warnings during startup

Warnings from Isaac Sim, Omniverse, or PyTorch may appear while a workload is starting. These warnings do not always mean the workload has failed.

If the process remains running in `nvidia-smi` and no fatal error appears in the logs, continue with the WebRTC connection step.

## Learn more

- [Isaac Lab documentation](https://isaac-sim.github.io/IsaacLab/release/3.0.0/index.html): Official Isaac Lab documentation, including tutorials, environments, reinforcement learning workflows, and API references.
- [Isaac Sim documentation](https://docs.isaacsim.omniverse.nvidia.com/6.1.0/index.html): NVIDIA Isaac Sim user guide.
