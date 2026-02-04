---
outline: [2, 3]
description: Learn how to submit applications in Olares Market.
---
# Submit applications

This guide explains how to submit a new Olares application to the default index by creating a Pull Request (PR) against `beclab/apps:main`. 

GitBot validates your PR based on the title, file scope, and ownership rules, and may automatically close invalid PRs.

## Prerequisites

Before submitting, make sure your application has been fully tested on Olares.

Recommended workflow:

- Use [Studio](/developer/develop/tutorial/develop.md) development container to test and debug in a real online environment.
- [Install the app via Olares Market](/developer/develop/tutorial/package-upload.md) to test installation and upgrades as a user would.

## Submit a new application

### Step 1: Add your OAC to your fork

1. Fork the [official repository](https://github.com/beclab/apps) `beclab/apps`.
2. In your fork, add your [Olares Application Chart (OAC)](/developer/develop/package/chart.md) under a new folder.
3. Ensure your OAC root directory contains an `owners` file with your GitHub username inside.

    :::info Folder naming convention
    The folder name is your OAC directory name (chart folder name). GitBot uses it in the PR title and for file-scope validation. It must:
    - Contain only lowercase letters and digits.
    - Not include hyphens (`-`).
    - Be no longer than 30 characters.
    :::

### Step 2: Create a draft PR

Create a draft PR targeting the `beclab/apps:main` branch, and set the PR title in this format:

```text
[PR type][Chart folder name][Version] Title content
```
    
| Field | Description|
|--|--|
| PR type | **NEW**: Submit a new application. <br>**UPDATE**: Update an already successfully merged application.<br>**REMOVE**: Remove an already successfully merged application.<br>**SUSPEND**: Suspend an already successfully merged application from<br> distribution through the application store. |
| Chart folder name | The directory name of your OAC. Must match the naming convention above. |
| Version | Chart version of your app. It must match:<br>- `version` in `Chart.yaml`<br>- `version` under `metadata` in `OlaresManifest.yaml` |
| Title content | A brief summary of your PR. |

### Step 3: Validate your PR

Before submitting the PR for review, confirm the following checklist:
1. **Title is valid**: The title contains only one PR type, one chart folder name, and one version.
2. **Changes are in scope**: The PR only adds or modifies content under the chart folder name declared in the title.
3. **No duplicate PR**: There is no other Open or Draft PR for the same chart folder name.
4. **You are an owner**: Your GitHub username is included in the `owners` file.
5. **For new applications**: 
    - The folder name does not already exist in `beclab/apps:main`.
    - Your chart folder does not contain `.suspend` or `.remove` files.

During the Draft stage, you can continue pushing commits to adjust your files.

When everything is ready, click **Ready for review**.

### Step 4: Wait for GitBot

After you submit, GitBot automatically validates the PR.

- If all checks pass, GitBot automatically merges the PR into `beclab/apps:main`. 
- After a short delay, your application appears in Olares Market.

## Track PR status

### Type labels

When your PR is labeled `NEW`, `UPDATE`, `REMOVE`, or `SUSPEND`, it indicates the PR type in the title is recognized. 
- Do not change the PR type after it is labeled. 
- If the type is wrong, close the PR and create a new PR.

### Status labels

- `waiting to submit`: Issues found. You may continue pushing commits. GitBot will re-check and update the status.
- `waiting to merge`: All checks passed and the PR is queued for auto-merge. Do not push new commits or manually intervene.
- `merged`: The PR has been merged into `beclab/apps:main`.
- `closed`: The PR is invalid or contains unrecoverable issues. Do not reopen it. Fix the issues and submit a new PR.

## Invite collaborators

You can collaborate on an application in two ways:

- **In the `owners` file (recommended)**: Add other developers' GitHub usernames to the owners file in your OAC. Each listed owner can independently fork the repo and submit changes for that app.
- **As repo collaborators**: Add others as collaborators to your personal forked repository. In this case, you must create the PR, but collaborators can push commits to your PR branch.