---
description: Developing applications on Olares leverages standard web technologies and containerization.
---
# Develop Olares applications

Developing applications on Olares leverages standard web technologies and containerization. If you are familiar with building web applications or Docker containers, you already have the skills needed to build for Olares.

This guide takes you through the complete lifecycle of an Olares application, from planning the package structure to publishing on the Market.

## Before you begin
Before getting started, it's helpful to review some concepts:
- [Application](../concepts/application.md)
- [Network](../concepts/network.md)

## Step 1: Package your application
To publish your application to the Olares Market, you must structure it according to the Olares Application Chart (OAC) specification. This format extends Helm Charts to support Olares-specific features like permission management and sandboxing.

* **[Understand the Olares Application Chart](./package/chart.md)**: Understand the file structure and requirements of an application package.
* **[Understand `OlaresManifest.yaml`](./package/manifest.md)**: A comprehensive guide to the `OlaresManifest.yaml` file, which defines your app's metadata, permissions, and system integration points.
* **[Understand Helm extensions](./package/extension.md)**: Learn about the custom fields and capabilities Olares adds to standard Helm deployments.

## Step 2: Submit your application
Once your application is built and packaged, the final step is to share it with the Olares community.

* **[Submit to Market](/developer/develop/distribute-index.md)**: Learn how to submit your application to the Olares Market for review and distribution.
