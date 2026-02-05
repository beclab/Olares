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

## App distribution workflow

### 1. Prepare your app package

Before an app can be distributed, it must be packaged as an **Olares Application Chart (OAC)**.

At this stage, developers typically:
- Develop and test the app on an Olares host.
- Verify installation and upgrade behavior.
- Finalize chart metadata and structure.

For details on packaging and chart structure, see:
- [Olares Application Chart (OAC)](/developer/develop/package/chart.md)

### 2. Submit the app to the default index

Olares Market indexes applications from Git repositories.
To publish an app to the default public index, developers submit their OAC by opening a Pull Request (PR) to the official repository.

During submission:
- The PR title declares the action type.
- GitBot validates file scope, ownership, and version rules.
- Valid PRs are merged automatically without manual review.

### 3. Automated validation and indexing

After a PR is submitted, **Terminus-Gitbot** performs automated checks to ensure the submission follows distribution rules.

If all checks pass, the PR is merged automatically.
After a short delay, the application becomes visible in Olares Market.

### 4. Manage the application lifecycle

After an application is published, developers continue to manage its lifecycle through Pull Requests.

Lifecycle actions include:
- Releasing new versions.
- Temporarily suspending distribution.
- Permanently removing an application from the Market.

These actions are controlled using PR types and special control files in the OAC.

### 5. Optimize your Market listing

Once published, you can improve how your app is presented in Olares Market by adding icons, screenshots, and featured images.

### 6. (Optional) Publish paid applications

Olares Market also supports paid application distribution.
Paid apps require additional identity registration, pricing configuration, and license management.
