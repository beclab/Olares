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

- **Olares Market**

    The application marketplace that indexes, displays, distributes, and installs Olares applications.

- **Application index**

    A service that provides application metadata to Olares Market. Olares includes a default public index, and developers may deploy their own.

- **GitBot**

    The automated validation system that checks application submissions and enforces distribution rules.

## How application distribution works

To submit an application to the Olares default index, follow these steps:
1. Develop and test your application on Olares, and create an OAC according to the guidelines.
2. Fork the [`beclab/apps` repository](https://github.com/beclab/apps) and add your application OAC to your fork.
3. Create a Pull Request targeting `beclab/apps:main` and wait for GitBot to validate whether the OAC meets the requirements.
4. Update the Pull Request if needed until all checks pass and the PR is automatically merged.
5. After a short delay, you can find your submitted application in Olares Market.