---
outline: [2, 3]
description: Use an AI agent powered by olares-cli agent skills to publish your website project with a custom domain.
head:
  - - meta
    - name: keywords
      content: Olares, Lares, publish website, custom domain, olares-cli, olares-cli agent skills
---

# Publish your website to a custom domain

With Olares, you can ask an AI agent to deploy a website you have already built. Powered by the Olares CLI, the agent packages the site, pushes the image, installs it on your device, and binds a custom domain so the site is available over HTTPS at a stable, public URL.

This tutorial uses Lares and a GitHub-hosted website project as an example to walk you through the whole deployment process. The same flow applies when you use another AI agent with Olares CLI agent skills to deploy your own website project.

## Prerequisites

Before you begin, ensure you have:

**Olares environment**

- An Olares device running v1.12.7 or later.
- Lares and Router installed, with a connected local model.
- Olares CLI (v1.12.7 or later) and Agent Skills. The CLI is included with Olares, so every AI agent running on your Olares device already has it. If you use an AI agent on your computer, install the [Olares CLI and Agent Skills](/developer/cli-overview.md#drive-olares-from-an-ai-agent), log in with your Olares ID, and all your local agents can use them.

**Your project**

- A working website project with source code available.

**Accounts and access**

- One of the following image registry options:
  - A Docker Hub account.
  - A GitHub account with permission to create a personal access token with the `write:packages` scope.
- A custom domain you own, with access to its DNS management console.

## Understand the deployment flow

The agent handles most of the packaging work, but the overall flow follows these stages:

1. **Source code**: The agent reads your project from Github, Olares Files, or your computer.
2. **Preview**: You run the site in the agent and open a preview URL to confirm it works.
3. **Build and deploy**: The agent creates a production build, Dockerfile, container image, and Olares chart, then installs the app.
4. **Bind and share**: The agent issues an RSA certificate and binds your custom domain. You add two DNS records, and the site is reachable over HTTPS.

## Step 1: Prepare your project source code

Tell the agent where your project source code is located. The steps differ depending on the source location.

### Source code on GitHub

Tell the agent the repository URL.
- For a public repository, provide the link and ask it to clone.
- For a private repository, give the agent a GitHub personal access token or an SSH key for authentication as prompted.

**Example:**

```text
Clone the repository from https://github.com/arnobt78/Portfolio-Landing-Page-7-React-Frontend
```

**Result:**

Lares clones the project into its default workspace at `Data/lares/data/workspace/` in Olares Files and starts working in that folder. It then gives a brief overview of the project and asks how to proceed. For example:

```text
Cloned successfully to Portfolio-Landing-Page-7-React-Frontend in the workspace. Quick overview: ...

Want me to install dependencies and run it locally, or explore the source code?
```

### Source code in Olares Files

In Lares, choose the project folder in Olares Files as the workspace. The agent runs build, package, and deploy commands from inside that workspace.

### Source code on your computer

If the source code is on your computer, you need to make it available to Lares on Olares. You can do this yourself or ask the agent to do it:

- **Push it to GitHub**: Push the project to a repository, then give the repository URL to the agent.
- **Upload it to Olares Files**: Upload the project to Files, then tell the agent the path.

## Step 2: Preview the website

Once the project is in place, Lares asks how to proceed. Tell it to run the site. Lares installs the project's dependencies if needed, starts a dev server, and returns a temporary preview URL. Open the URL in your browser and make sure the site works as expected.

**Example:**

```text
The app is up and running 🎉

Dev server — Vite v7.3.1:
Local: http://localhost:5173/
Verified: HTTP 200, page serves correctly
```

## Step 3: Build, deploy, and publish

After you confirm the preview, tell the agent to publish the site to your custom domain.

The agent takes it from there. It checks your environment (Olares version, node architecture, Docker setup), builds the production site, and packages it as a container image right on your Olares device, matched to the node's architecture.

**Example:**

```text
The preview looks good. Publish it to `website.bellame.online`.
```

### Provide an image registry

The agent needs a registry to store the image and will ask which one to use.

- **Docker Hub (recommended)**: Give the agent your Docker Hub username. When it asks for credentials, follow its instructions to create an access token and paste it. The agent logs in, pushes the image, and verifies the image can be pulled anonymously, which is how the Olares node downloads it. Delete the token in Docker Hub afterwards. It is only needed for this one push.

- **GitHub Container Registry**: The agent can push the image to `ghcr.io` under your GitHub account. When it asks for credentials, create a personal access token with the `write:packages` scope and paste it. The package must be set to public so the Olares node can pull it anonymously.

Once the image is pushed, the agent creates an Olares chart with the entrance auth level set to `public` (a custom domain requires this), uploads the chart, and installs the app. When it finishes, the app appears on the Launchpad.

If the installation gets stuck, see [Common issues](#common-issues).

## Step 4: Bind the custom domain

Once the app is running, the agent starts binding your custom domain. It issues an RSA certificate for the domain and attaches it to the app's entrance. You only need to add two DNS records in your provider's console, and the agent gives you the exact values. The app restarts once during the binding, which is normal.

Olares only accepts RSA certificates. The ECDSA certificates that some tools default to can break the BFL service, so the agent always requests RSA.

### 1. Add a TXT record

To prove you own the domain, the agent asks you to add one TXT record in your DNS console.

**Example:**

- Type: `TXT`
- Name: `_acme-challenge.website` (must include the full subdomain)
- Value: The exact value provided by the agent

The agent keeps checking public DNS and issues the RSA certificate automatically once the TXT record shows up there. This can take a few seconds to a few minutes. The temporary TXT record is no longer needed after issuance and you can delete it then.

### 2. Add a CNAME record

Add a CNAME record that points your domain to Olares, so that visits to your domain are routed to your Olares device.

**Example:**

- Type: `CNAME`
- Name: `website`
- Value: `laresprime.olares.com`

### 3. Verify

The agent runs the final checks: the certificate, the domain binding, the CNAME record, and an end-to-end HTTPS test. When everything passes, it reports that the site is live.

**Example:**

```text
✓ RSA certificate issued and valid
✓ Domain website.bellame.online bound to the app entrance
✓ CNAME record detected and active
✓ HTTPS check passed (HTTP 200)
The site is live at https://website.bellame.online 🎉
```

Open the live site in your browser to confirm, and send the URL to the people you want to share it with.

:::tip Certificate renewal
Let's Encrypt certificates are valid for 90 days. When you renew, ask the agent to re-bind the domain with the new certificate. The app restarts briefly during the re-bind.
:::

## Step 5: Update the website (optional)

After the site is live, you can keep improving it.

Tell the agent what you want to change. The agent rebuilds the site with a new image tag, uses a new chart version, and re-uploads and reinstalls the app. Your domain, certificate, and DNS records stay as they are.

## Common issues

### Installation gets stuck in `Initializing`

Common causes:

- The image architecture does not match the Olares node.
- The image tag was reused and the old layer is cached on the node.
- The chart version was not changed before re-uploading.

Try uninstalling, deleting the old version, using new image tag and chart versions, and reinstalling:

```bash
olares-cli market uninstall <app-name>
olares-cli market delete --version <old-version>
olares-cli market upload <new-chart>
olares-cli market install <app-name> -s upload --watch
```

### Image architecture does not match the node

The agent detects your node's architecture and builds the image on the Olares device itself, so a mismatch rarely happens. If an image built for the wrong architecture ever gets installed (for example `linux/arm64` on an AMD64 node), the container crashes with an `exec format error`. Tell the agent to rebuild the image for the node's architecture and reinstall the app.

### nginx permission errors

If the container runs as a non-root user, nginx may fail to write its PID file. Make the Dockerfile update the PID path and give the nginx user ownership of required directories:

```dockerfile
RUN sed -i 's|/run/nginx.pid|/tmp/nginx.pid|' /etc/nginx/nginx.conf \
    && chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx /var/log/nginx /etc/nginx/conf.d
USER nginx
```

### BFL service becomes unreachable after uploading a certificate

This usually means the certificate is ECDSA instead of RSA. Re-request the certificate with `--key-type rsa` and upload it again.

### DNS TXT record validation fails

When adding the TXT record in your DNS provider, the record name must include the full subdomain. For example, if your domain is `n1.monster` and the certificate is for `portfolio.n1.monster`, the TXT name should be `_acme-challenge.portfolio`, not just `_acme-challenge`.

If the record is correct but validation still fails, a public resolver cache may be serving an old TXT value from a previous attempt. Wait for the cache to expire, or check the value against your domain's authoritative name servers.
