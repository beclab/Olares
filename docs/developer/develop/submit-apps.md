---
outline: [2, 3]
description: Learn how to submit applications in Olares Market.
---
# Submit applications

This guide explains how to submit a new Olares application to the default index by creating a Pull Request (PR) against `beclab/apps:main`. 

Terminus-Gitbot validates your PR based on the title, file scope, and ownership rules, and automatically close invalid PRs.

## Prerequisites

Before submitting, make sure your application has been fully tested on Olares.

Recommended workflow:

- Use [Studio](/developer/develop/tutorial/develop.md) development container to test and debug in a real online environment.
- [Install the app via Olares Market](/developer/develop/tutorial/package-upload.md) to test installation and upgrades as a user would.

## Submit a new application

### Step 1: Add your OAC to your fork

1. Fork the [official repository](https://github.com/beclab/apps) `beclab/apps`.
2. In your fork, add your [Olares Application Chart (OAC)](/developer/develop/package/chart.md) under a new folder.
3. Create an [`owners` file](/developer/develop/distribute-index.md#before-you-begin) (no extension) in your OAC root directory. Ensure your GitHub username is included.
    
    :::info Folder naming convention
    The folder name is your OAC directory name (chart folder name). Terminus-Gitbot uses it in the PR title and for file-scope validation. It must:
    - Contain only lowercase letters and digits.
    - Not include hyphens (`-`).
    - Be no longer than 30 characters.
    :::

### Step 2: Create a draft PR

Create a draft PR targeting the `beclab/apps:main` branch. 

Terminus-Gitbot checks both your PR metadata (such as the title and file scope) and your chart content (such as required files in the OAC root). Make sure you have completed Step 1 before proceeding.

To pass the Terminus-Gitbot automated checks, your PR must strictly follow these rules:
1. **Title format**: The title must imply your intent and follow this exact format:
```text
[PR type][Chart folder name][Version] Title content
```
    
| Field | Description|
|--|--|
| PR type | <ul><li>**NEW**: Submit a new application. <br></li><li>**UPDATE**: Update an already successfully merged application.<br></li><li>**REMOVE**: Remove an already successfully merged application.<br></li><li>**SUSPEND**: Suspend an already successfully merged application from<br> distribution through the application store.</li></ul> |
| Chart folder name | The directory name of your OAC. Must match the naming convention. |
| Version | Chart version of your app. It must match:<br><ul><li>`version` in `Chart.yaml`</li><br><li>`version` under `metadata` in `OlaresManifest.yaml`</li></ul> |
| Title content | A brief summary of your PR. |

2. **File scope**: The PR only adds or modifies content under the chart folder name declared in the title.
3. **No duplicate PR**: Ensure no other Open or Draft PR  exists for this chart folder.
4. **Clean structure (For new apps)**: 
    - The folder name does not already exist in `beclab/apps:main`.
    - Your chart folder does not contain [control files](/developer/develop/manage-apps.md#control-files) (`.suspend` or `.remove`).

:::tip Draft PR is editable
During the Draft stage, you can continue pushing commits to adjust your files.
:::

When everything is ready, click **Ready for review**.

### Step 3: Wait for Terminus-Gitbot

After you submit, Terminus-Gitbot automatically validates the PR.

- If all checks pass, Terminus-Gitbot automatically merges the PR into `beclab/apps:main`. 
- After a short delay, your application appears in Olares Market.

## Track PR status

### Type labels

When your PR is labeled `NEW`, `UPDATE`, `REMOVE`, or `SUSPEND`, it indicates the PR type in the title is recognized. 

:::warning No type change
- Do not change the PR type after it is labeled. 
- If the type is wrong, close the PR and create a new PR.
:::

### Status labels

- `waiting to submit`: Issues found. You may continue pushing commits. Terminus-Gitbot will re-check and update the status.
- `waiting to merge`: All checks passed and the PR is queued for auto-merge. Do not push new commits or manually intervene.
- `merged`: The PR has been merged into `beclab/apps:main`.
- `closed`: The PR is invalid or contains unrecoverable issues. Do not reopen it. Fix the issues and submit a new PR.