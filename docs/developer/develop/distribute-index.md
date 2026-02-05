---
outline: [2, 3]
description: Understand how application distribution works in Olares.
---
# Distribute Olares applications

Distributing applications on Olares is based on open standards and automated validation.
If your application is already packaged as an Olares Application Chart (OAC), you can publish it to Olares Market and make it available to users with minimal friction.

This guide walks you through the distribution lifecycle of an Olares application, from understanding how Market indexing works to publishing, maintaining, and promoting your app.

## Before you begin

Before distributing your application, it helps to understand a few core concepts:

- **[Olares Application Chart (OAC)](/developer/develop/package/chart.md)**
    
    The packaging format used to describe an Olares application, including metadata, ownership, versioning, and installation configuration.

- **Application index**

    A service that provides application metadata to Olares Market. Olares includes a default public index, and developers can deploy their own.

- **Terminus-Gitbot**

    The automated validation system that checks application submissions and enforces distribution rules.

- **Owners file (`owners`)**
  
    A file in the OAC root directory used to validate ownership and permissions. The file has no extension.
    ```text
    owners:
    - <your-github-username>
    - <collaborator1-username>
    - <collaborator2-username>
    ```
- **Control files**

    Special empty files in the OAC root directory that control distribution status:
    - `.suspend`: suspend distribution
    - `.remove`: remove an app from the Market
    For details, see [Manage the app lifecycle](/developer/develop/manage-apps.md#control-files).

## Ownership and collaboration
    
To collaborate as a team:
- Add all maintainers to the owners file (recommended). Each listed owner can independently fork the repo and submit changes for that app.
- Add teammates as collaborators to your forked repository so they can push commits to your PR branch.

## App distribution workflow

### 1. Prepare your app package

Before an app can be distributed, it must be packaged as an **Olares Application Chart (OAC)**.

At this stage, developers typically:
- Develop and test the app on an Olares host.
- Verify installation and upgrade behavior.
- Finalize chart metadata and structure.

For details, see [Olares Application Chart (OAC)](/developer/develop/package/chart.md)

### 2. Submit the app to the default index

Olares Market indexes applications from Git repositories.
To publish an app to the default public index, developers submit their OAC by opening a PR to the official repository.

During submission:
- The PR title declares the action type.
- Terminus-Gitbot validates file scope, ownership, and version rules.
- Valid PRs are merged automatically without manual review.

For details, see [Submit applications](/developer/develop/submit-apps.md)

### 3. Automated validation and indexing

After a PR is submitted, Terminus-Gitbot performs automated checks to ensure the submission follows distribution rules.

If all checks pass, the PR is merged automatically.
After a short delay, the application becomes visible in Olares Market.

### 4. Manage the application lifecycle

After an application is published, developers continue to manage its lifecycle through Pull Requests.

Lifecycle actions include:
- Releasing new versions.
- Temporarily suspending distribution.
- Permanently removing an application from the Market.

These actions are controlled using PR types and special control files in the OAC.

For details, see [Manage the app lifecycle](/developer/develop/manage-apps.md).

### 5. Optimize your Market listing

Once published, you can improve how your app is presented in Olares Market by adding icons, screenshots, and featured images.

For details, see [Promote your apps](/developer/develop/promote-apps.md).

### 6. (Optional) Publish paid applications

Olares Market also supports paid application distribution.
Paid apps require additional identity registration, pricing configuration, and license management.

For details, see [Publish paid applications](/developer/develop/paid-apps.md).