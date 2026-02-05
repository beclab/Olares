---
outline: [2, 3]
description: Learn how to update, suspend, or remove your application from Olares Market.
---
# Manage the app lifecycle

This guide explains how to manage an application after it has been published, including updating, suspending, or removing it from Olares Market.

All actions are performed through Pull Requests (PRs) to the `main` branch. Terminus-Gitbot supports three lifecycle actions after your app is published:

- **UPDATE**: Keep your app up to date. Release new versions, fix bugs, or adjust configurations.
- **SUSPEND**: Pause distribution. Stop new discovery, downloads, and installs in Olares Market without affecting existing users.
- **REMOVE**: Retire the app. Permanently stop distribution and prevent the chart folder name from being reused.

:::tip Reduce conflicts
Before opening the PR, sync your fork and rebase your branch onto the latest `main` to reduce potential conflicts.
:::

## Control files

Control files are special empty files in the OAC root directory that manage an application's distribution status in Olares Market.

| File name | Used for | Version rule | Content needed |
|--|--|--|--|
| `.suspend` | Suspend distribution | Upgrade (>) | Empty file |
| `.remove` | Remove application | Same (=) | Empty file |

An `UPDATE` PR or a `NEW` PR must not include these files. They are only used for `SUSPEND` and `REMOVE`.

## Update an app (UPDATE)

To update an existing application, such as releasing a new version, changing configurations, or updating the owners, submit a PR with type `UPDATE`.

The PR must meet the following requirements:

- **Version bump**: The new Chart version must be greater than the current version in the repository. Any change to a chart must bump the Chart version.
- **Clean directory**: The OAC root must not contain `.suspend` or `.remove` files.
- **No conflict**: The PR branch must not conflict with `beclab/apps:main`.

:::warning No rollbacks
Olares Market does not support version rollbacks. If an issue occurs, you must submit a new version to fix it.
:::

## Suspend an app (SUSPEND)

To temporarily stop your application from being listed, downloaded, or installed, submit a PR with type `SUSPEND`. 

The PR must meet the following requirements:
- **Version bump**: The Chart version must be greater than the current version in the repository.
- **Control file**: The OAC root directory contains the `.suspend` file and does not contain the `.remove` file.
- **No conflict**: The PR branch must not conflict with `beclab/apps:main`.

After the PR is merged, the application is no longer listed in Olares Market. Users who have already installed the application can continue to use it.

## Remove an app (REMOVE)

To permanently remove an application from Olares Market, submit a PR with type `REMOVE`.

The PR must meet the following requirements:

- **Same version**: The Chart version in the PR title must be the same as the current version in the repository.
- **Control file**: After the change, the `.remove` file is the only file in the OAC root directory.
- **No conflict**: The PR branch must not conflict with `beclab/apps:main`.

:::warning
Removal is irreversible.
:::

After the PR is merged:

- The chart folder name cannot be reused by the application owners.
- Users who have already installed the application can continue to use it.