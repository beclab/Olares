# Develop Olares applications

Developing applications on Olares leverages standard web technologies and containerization. If you are familiar with building web applications or Docker containers, you already have the skills needed to build for Olares.

This guide takes you through the complete lifecycle of an Olares application, from your first line of code in Studio to publishing on the Market.

## Before you begin
Before getting started, it's helpful to review some concepts:
- [Application](../concepts/application.md)
- [Network](../concepts/network.md)

## Step 1: Develop with Studio
Olares Studio is a development platform that accelerates your build cycle. It provides a pre-configured workspace to build, debug, and test your applications directly on the platform.

* **[Deploy an app](./tutorial/deploy.md)**: Learn how to quickly deploy an app from an existing Docker image, configure it, and test it in Studio.
* **[Develop in a dev container](./tutorial/develop.md)**: Spin up a remote development environment (Dev Container) and connect it to VS Code for a seamless coding experience.
* **[Package and upload](./tutorial/package-upload.md)**: Convert your running application into an Olares-compatible package and upload it for testing.
* **[Add app assets](./tutorial/assets.md)**: Configure icons, screenshots, and descriptions to make your application store-ready.

## Step 2: Package your application
To publish your application to the Olares Market, you must structure it according to the Olares Application Chart (OAC) specification. This format extends Helm Charts to support Olares-specific features like permission management and sandboxing.

* **[Understand the Olares Application Chart](./package/chart.md)**: Understand the file structure and requirements of an application package.
* **[Understand `OlaresManifest.yaml`](./package/manifest.md)**: A comprehensive guide to the `OlaresManifest.yaml` file, which defines your app's metadata, permissions, and system integration points.
* **[Understand Helm extensions](./package/extension.md)**: Learn about the custom fields and capabilities Olares adds to standard Helm deployments.

## Step 3: Submit your application
Once your application is built and packaged, the final step is to share it with the Olares community.

* **[Submit to Market](/developer/develop/distribute-index.md)**: Learn how to submit your application to the Olares Market for review and distribution.