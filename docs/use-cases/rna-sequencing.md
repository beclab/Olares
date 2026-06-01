---
outline: [2, 3]
description: Run GPU-accelerated single-cell RNA sequencing analysis on Olares using a pre-configured JupyterLab environment with RAPIDS-singlecell.
head:
  - - meta
    - name: keywords
      content: Olares, RNA Sequencing, scRNA-seq, single-cell, JupyterLab, RAPIDS, GPU, bioinformatics, self-hosted
app_version: "1.0.2"
doc_version: "1.0"
doc_updated: "2026-05-09"
---

# Analyze single-cell RNA sequencing data

RNA Sequencing provides a pre-configured JupyterLab environment for GPU-accelerated single-cell RNA sequencing (scRNA-seq) analysis. It includes a ready-to-run notebook, `scRNA_analysis_preprocessing.ipynb`, which demonstrates a preprocessing and analysis workflow based on RAPIDS-singlecell.

Use this guide to open the notebook, run the full analysis pipeline, and find the generated output files.

## Prerequisites

- An Olares device with AMD64 architecture and an NVIDIA GPU with at least 10 GB VRAM available for this app.

## Install RNA Sequencing

1. Open Market and search for "RNA Sequencing".
   
   ![RNA Sequencing](/images/manual/use-cases/rna-sequencing.png#bordered){width=90%}

2. Click **Get**, then **Install**, and wait for installation to complete.

## Run the analysis notebook

1. Open RNA Sequencing from Launchpad. The app opens in a JupyterLab interface.

   ![JupyterLab interface](/images/manual/use-cases/rna-sequencing-jupyter.png#bordered){width=90%}

2. In the file browser on the left, double-click `scRNA_analysis_preprocessing.ipynb` to open the analysis notebook. 

   The notebook is pre-configured and can be run directly.

   ![Open notebook](/images/manual/use-cases/rna-sequencing-notebook.png#bordered){width=95%}

   :::tip Back up the original notebook
   To keep the original notebook unchanged, right-click `scRNA_analysis_preprocessing.ipynb` and select **Duplicate**. This creates a copy in the same folder. Open and run the copied notebook instead.
   :::

3. Click **Run** > **Run All Cells** to execute the full analysis pipeline. 

   ![Run All Cells](/images/manual/use-cases/rna-sequencing-run-all.png#bordered){width=95%}

4. Wait for all cells to finish. 

   While a cell is running, its execution indicator shows `[*]`. After cells finish running, executed cells show numbers such as `[1]` and `[2]`, and generated charts or tables appear below the corresponding cells.

5. After the run completes, open the `h5` folder in the file browser to find the generated output files.

The steps above run the pre-configured pipeline. To customize the workflow, you can edit the notebook, create new notebooks, or run selected cells using standard JupyterLab features. For general JupyterLab usage, refer to the [JupyterLab documentation](https://jupyterlab.readthedocs.io/en/latest/).

## Learn more

- [RAPIDS-singlecell documentation](https://rapids-singlecell.readthedocs.io/en/latest/): API reference and tutorials for GPU-accelerated scRNA-seq analysis.
- [NVIDIA Single-Cell Playbook](https://build.nvidia.com/spark/single-cell): The original playbook this app is based on.
