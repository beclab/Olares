# Olares documentation

Welcome to the source repository for the official Olares documentation!

This directory includes the source files for the Olares documentation website, instructions for setting up the project locally, Markdown references, and style guides to ensure consistency and quality across all documentation.

## Quick links

* **Published documentation site**: https://docs.olares.com
* **Olares official website**: https://www.olares.com
* **Olares project on GitHub**: https://github.com/beclab/Olares

## Getting started with Olares documentation

We welcome community contributions! Follow these steps to preview, develop, and build the documentation locally.

### Install dependencies

First, ensure you have [Node.js](https://nodejs.org/) installed (LTS version recommended). Then, run the following command in the project's root directory to install the required dependencies:

```bash
npm install
````

### Start the development server

Start the local development server with hot-reloading. This allows you to see your changes live in the browser as you edit the source files.

```bash
npm run dev
```

Once running, open your browser and navigate to `http://localhost:5173/` to see the local preview.

### Build the site locally

To generate a production-ready build for final review or debugging, run:

```bash
npm run build
```

This command will build the static site into the `dist` directory.

## Versioning and branching strategy

To manage documentation for different product versions effectively, we use the following branching strategy:

* **`main` branch**:
  This branch contains the latest documentation for the **next, in-development version** of Olares. All documentation for new features should be submitted here.

* **`release-{version}` branch**:
  These branches hold the documentation for **recent, stable versions**. For example, the `release-1.11` branch corresponds to the `Olares 1.11` documentation. Fixes or clarifications for a specific stable version should be submitted to its corresponding release branch.

Creating a PR to a corresponding branch will automatically trigger the documentation build for that version.

## Style guide

To ensure clarity, accuracy, and a consistent tone, please read our **[Content and style guide](https://github.com/beclab/Olares/wiki/General-style-reference)** before contributing.