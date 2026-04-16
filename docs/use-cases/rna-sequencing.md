---
outline: [2, 3]
description: Run GPU-accelerated single-cell RNA sequencing analysis on Olares using a pre-configured Jupyter Lab environment with RAPIDS-singlecell.
head:
  - - meta
    - name: keywords
      content: Olares, RNA Sequencing, scRNA-seq, single-cell, Jupyter Lab, RAPIDS, GPU, bioinformatics, self-hosted
app_version: "1.0.2"
doc_version: "1.0"
doc_updated: "2026-04-13"
---

# Analyze single-cell RNA sequencing data

RNA Sequencing is a GPU-accelerated single-cell RNA sequencing (scRNA-seq) analysis environment running on Jupyter Lab. It uses the RAPIDS-singlecell library to perform data preprocessing, quality control, visualization, and downstream analysis entirely on the GPU, significantly faster than CPU-based tools.

## Prerequisites

- An Olares device with AMD64 architecture and a NVIDIA GPU with at least 10 GB VRAM

## Install RNA Sequencing

1. Open Market and search for "RNA Sequencing".
   <!-- ![RNA Sequencing](/images/manual/use-cases/rna-sequencing.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Run the analysis notebook

Open RNA Sequencing from Launchpad. The app opens in a Jupyter Lab interface.

<!-- ![Jupyter Lab interface](/images/manual/use-cases/rna-sequencing-jupyter.png#bordered) -->

1. In the file browser on the left, double-click `scRNA_analysis_preprocessing.ipynb` to open the analysis notebook.

   <!-- ![Open notebook](/images/manual/use-cases/rna-sequencing-notebook.png#bordered) -->

   :::tip Back up the original notebook
   If you want to keep the original notebook unchanged, right-click the file and select **Copy**, then paste it in the same directory. Run the copy instead.
   :::

2. Click **Run** > **Run All Cells** to execute the full analysis pipeline.

   <!-- ![Run All Cells](/images/manual/use-cases/rna-sequencing-run-all.png#bordered) -->

3. Wait for all cells to finish. You can view inline visualizations directly in the notebook as cells complete.

4. Find the output data files in the `h5` folder in the file browser.

You can also modify the notebook, create new notebooks, or run individual cells as needed. For general Jupyter Lab usage, refer to the [JupyterLab documentation](https://jupyterlab.readthedocs.io/en/latest/).

## Learn more

- [RAPIDS-singlecell documentation](https://rapids-singlecell.readthedocs.io/en/latest/): API reference and tutorials for GPU-accelerated scRNA-seq analysis.
- [NVIDIA Single-Cell Playbook](https://build.nvidia.com/spark/single-cell): The original playbook this app is based on.
