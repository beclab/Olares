# Olares Apps

## Overview

This directory contains the code for system applications, primarily for LarePass. The following are the pre-installed system applications that offer tools for managing files, knowledge, passwords, and the system itself.

## System Applications Overview

| Application | Description |
| --- | --- |
| Files | A file management app that manages and synchronizes files across devices and sources, enabling seamless sharing and access. |
| Wise | A local-first and AI-native modern reader that helps to collect, read, and manage information from various platforms. Users can run self-hosted recommendation algorithms to filter and sort online content. |
| Vault | A secure password manager for storing and managing sensitive information across devices. |
| Market | A decentralized and permissionless app store for installing, uninstalling, and updating applications and recommendation algorithms. |
| Desktop | A hub for managing and interacting with installed applications. File and application searching are also supported. |
| Profile | An app to customize the user's profile page. |
| Settings | A system configuration application. |
| Dashboard | An app for monitoring system resource usage. |
| Control Hub | The console for Olares, providing precise and autonomous control over the system and its environment. |
| Studio | A development tool for building and deploying Olares applications. |


# Local Development Guide

This document describes how to start and develop various sub-projects locally.

## Available Projects

| Project | Command | Port |
|---------|---------|------|
| Desktop | `npm run dev:desktop` | 1090 |
| Files | `npm run dev:files` | 5090 |
| Settings | `npm run dev:settings` | 9000 |
| Market | `npm run dev:market` | 8080 |
| Vault | `npm run dev:vault` | 8090 |
| Wise | `npm run dev:wise` | 8100 |
| Dashboard | `npm run dev:dashboard` | 9003 |
| Control Hub | `npm run dev:hub` | 9002 |
| Share | `npm run dev:share` | 5070 |
| Editor | `npm run dev:editor` | 9100 |
| Preview | `npm run dev:preview` | 9001 |
| Studio | `npm run dev:studio` | 9001 |

## Step 1: Modify Local Hosts File

Projects require access through a specific domain name. You need to configure the local hosts file first.

### macOS / Linux

1. Open terminal and edit the hosts file with administrator privileges:

```bash
sudo vim /etc/hosts
```

Or use the nano editor:

```bash
sudo nano /etc/hosts
```

2. Add the following content at the end of the file:

```
127.0.0.1       test.xxx.olares.com
```

3. Save the file and exit
   - vim: Press `ESC`, type `:wq` and press Enter
   - nano: Press `Ctrl + O` to save, `Ctrl + X` to exit

### Windows

1. Run Notepad as administrator:
   - Search for "Notepad" in the Start menu
   - Right-click on "Notepad" and select "Run as administrator"

2. Open the hosts file in Notepad:
   - Click `File` -> `Open`
   - Paste the path in the filename field: `C:\Windows\System32\drivers\etc\hosts`
   - Change file type to "All Files (*.*)"
   - Click "Open"

3. Add the following content at the end of the file:

```
127.0.0.1       test.xxx.olares.com
```

4. Save the file (`Ctrl + S`)

5. Flush DNS cache (optional):
   - Open Command Prompt (CMD) as administrator
   - Run the following command:

```cmd
ipconfig /flushdns
```

## Step 2: Install Dependencies

Run in the project root directory (`olares-app`):

```bash
npm install
```

## Step 3: Configure Environment Variables

Create or edit the `.env` file in the `packages/app` directory and add the following content:

```env
ACCOUNT_DOMAIN=xxx.olares.com
DEV_DOMAIN=test.xxx.olares.com
```

> **Note**:
> - `ACCOUNT_DOMAIN`: Your Olares account domain, used for API proxy
> - `DEV_DOMAIN`: Local development server domain, must match the domain configured in the hosts file

## Step 4: Start the Project

After configuring the `.env` file, run the corresponding command in the `packages/app` directory:

```bash
# Start Desktop
npm run dev:desktop

# Start Files
npm run dev:files

# Start Settings
npm run dev:settings

# Start Market
npm run dev:market

# Start other projects...
npm run dev:<project>
```

## Step 5: Access the Application

After successful startup, visit in your browser (replace port according to the project):

| Project | URL |
|---------|-----|
| Desktop | `https://test.xxx.olares.com:1090` |
| Files | `https://test.xxx.olares.com:5090` |
| Settings | `https://test.xxx.olares.com:9000` |
| Market | `https://test.xxx.olares.com:8080` |
| Vault | `https://test.xxx.olares.com:8090` |
| Wise | `https://test.xxx.olares.com:8100` |
| Dashboard | `https://test.xxx.olares.com:9003` |
| Control Hub | `https://test.xxx.olares.com:9002` |
| Share | `https://test.xxx.olares.com:5070` |
| Editor | `https://test.xxx.olares.com:9100` |
| Preview | `https://test.xxx.olares.com:9001` |
| Studio | `https://test.xxx.olares.com:9001` |

> **Note**: Since a self-signed certificate is used, the browser may display an insecure connection warning. Click "Advanced" and select "Proceed" to continue.

## Environment Variables (.env file)

| Variable | Description | Example |
|----------|-------------|---------|
| `ACCOUNT_DOMAIN` | Account domain (for API proxy) | `xxx.olares.com` |
| `DEV_DOMAIN` | Development server domain | `test.xxx.olares.com` |

## FAQ

### 1. Cannot Access the Application

- Check if the hosts file is configured correctly
- Ensure the development server has started successfully
- Check if the firewall is blocking the corresponding port

### 2. Certificate Error

The development server uses HTTPS. The browser will show a certificate warning on first visit - this is expected behavior.

### 3. API Request Failed

Ensure the `ACCOUNT_DOMAIN` in the `.env` file is set correctly. The proxy configuration relies on this variable to forward requests to the correct backend service.

## Build for Production

```bash
# Build Desktop
npm run build:desktop

# Build Files
npm run build:files

# Build Settings
npm run build:settings

# Build other projects...
npm run build:<project>
```

### Build Output Directory

| Project | Output Directory |
|---------|------------------|
| Desktop | `dist/apps/desktop` |
| Files | `dist/apps/files` |
| Settings | `dist/apps/settings` |
| Market | `dist/apps/market` |
| Vault | `dist/apps/vault` |
| Dashboard | `dist/apps/dashboard` |
| Control Hub | `dist/apps/control-hub` |
| Share | `dist/apps/share` |
| Editor | `dist/apps/editor` |
| Preview | `dist/apps/preview` |
| **Wise** | `dist/spa` |
| **Studio** | `dist/spa` |

> **Note**: Build outputs for Wise and Studio are located in `dist/spa` directory, not under `dist/apps/`.