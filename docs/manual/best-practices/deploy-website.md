---
outline: [2, 3]
description: Publish a website you have already built with a custom domain through an AI agent such as Lares, using the Portfolio Landing Page project as an example.
head:
  - - meta
    - name: keywords
      content: Olares, Lares, Router, deploy website, custom domain, image registry, preview, portfolio
---

# Publish a website with a custom domain

Olares lets you deploy a website you have already built through an AI agent. Powered by the Olares CLI, the agent packages the site, pushes the image, installs it on your device, and binds a custom domain so other people can reach the site over HTTPS.

This tutorial walks through the whole flow with Lares, using a website project on GitHub. You can follow the same steps with other AI agent, or adapt them to your own website project.

## Prerequisites

Before you begin, ensure you have:

**Olares environment**

- An Olares device running v1.12.7 or later.
- Lares and Router installed, with a connected local model.
- Olares CLI (v1.12.7 or later) and Agent Skills. These come preinstalled with Lares. You only need to install and log in them yourself if you use an AI agent on your computer.

**Your project**

- A working website project with source code available.

**Accounts and access**

- One of the following image registry options:
  - A Docker Hub account.
  - A GitHub account with permission to create a personal access token with the `write:packages` scope.
- A custom domain you own, with access to its DNS management console.

## Understand the deployment flow

The agent handles most of the packaging work, but the overall flow follows these stages:

1. **Source code**: The agent reads your project from your computer, Olares Files, or GitHub.
2. **Preview**: You run the site in Lares and open a preview URL to confirm it works.
3. **Build and deploy**: The agent creates a production build, Dockerfile, container image, and Olares chart, then installs the app.
4. **Bind and share**: The agent issues an RSA certificate and binds your custom domain. You add two DNS records, and the site is reachable over HTTPS.

## Step 1: Prepare your project source code

Tell the agent where your project lives. The steps differ depending on the source location.

:::info Example project
This tutorial uses the Portfolio Landing Page public repository at `https://github.com/arnobt78/Portfolio-Landing-Page-7-React-Frontend` as an example. Replace the URL and folder names with your own project.
:::

### Source code is on GitHub

Tell the agent the repository URL.
- For a public repository, paste the link and ask it to clone.
- For a private repository, give the agent a GitHub personal access token or an SSH key for authentication as prompted.

Example:

```text
Clone the repository from https://github.com/arnobt78/Portfolio-Landing-Page-7-React-Frontend
```

Lares clones the project into its default workspace at `Data/lares/data/workspace/` in Olares Files and starts working in that folder. It then gives a brief overview of the project and asks how to proceed. For example:

```text
Cloned successfully to Portfolio-Landing-Page-7-React-Frontend in the workspace. Quick overview: ...

Want me to install dependencies and run it locally, or explore the source code?
```

Reply that you want to run it, and the agent moves on to preview the website.

### Source code is in Olares Files

Lares can open the project folder as a workspace directly. The agent runs build, package, and deploy commands from inside that workspace.

### Source code is on your computer

If the source code is on your computer, you need to make it available to Lares on Olares. You can do this yourself or ask the agent to do it:

- **Push it to GitHub** yourself, or ask the agent to create a repo and push it. Then give the repository URL to the agent.
- **Upload it to Olares Files** under `Home/Code` yourself, or ask the agent to copy it there. Then open that folder as a workspace in Lares.

## Step 2: Preview the website

Once the project is in place, tell Lares to run the site. Lares installs the project's dependencies if needed, starts a dev server, and returns a temporary preview URL. Open the URL in your browser and make sure the site works as expected.

For example, Lares reports the dev server is up and gives the preview link:

```text
The app is up and running 🎉

Dev server — Vite v7.3.1:
Local: http://localhost:5173/
Verified: HTTP 200, page serves correctly
```

## Step 3: Build, deploy, and publish

After you confirm the preview, tell the agent to publish the site to your custom domain:

```text
The preview looks good. Publish it to `website.bellame.online`.
```

The agent takes it from there. It checks your environment (Olares version, node architecture, Docker setup), builds the production site, and packages it as a container image right on your Olares device, matched to the node's architecture.

### Provide an image registry

The agent needs a registry to store the image and will ask which one to use.

- **Docker Hub (recommended)**: Give the agent your Docker Hub username. When it asks for credentials:

  1. In Docker Hub, go to **Account Settings → Personal Access Tokens → Generate new token** and create a token.
  2. Paste the token to the agent. It logs in, pushes the image, and verifies the image can be pulled anonymously, which is how the Olares node downloads it.
  3. Delete the token in Docker Hub afterwards. It is only needed for this one push.

- **GitHub Container Registry**: The agent can push the image to `ghcr.io` under your GitHub account. When it asks for credentials, create a personal access token with the `write:packages` scope and paste it. The package must be set to public so the Olares node can pull it anonymously.

Once the image is pushed, the agent creates an Olares chart with the entrance auth level set to `public` (a custom domain requires this), uploads the chart, and installs the app. When it finishes, the app appears on the Launchpad.

If the installation gets stuck, see [Common issues](#common-issues).

:::tip Update the site later
When you want to change the site, just send the agent your changes. It rebuilds the site with a new image tag, bumps the chart version, and re-uploads and reinstalls the app.
:::

## Step 4: Bind the custom domain

Once the app is running, the agent starts binding your custom domain. It issues an RSA certificate for the domain and attaches it to the app's entrance. You only need to add two DNS records in your provider's console, and the agent gives you the exact values. The app restarts once during the binding, which is normal.

Olares only accepts RSA certificates. The ECDSA certificates that some tools default to can break the BFL service, so the agent always requests RSA.

### 1. Add a TXT record for certificate issuance

To prove you own the domain, the agent asks you to add one TXT record:

- Type: `TXT`
- Name: `_acme-challenge.website` (must include the full subdomain)
- Value: the exact value the agent gives you

The agent polls public DNS and completes the certificate automatically once the record propagates. The temporary TXT record is no longer needed after issuance and can be deleted.

### 2. Add a CNAME record

Add a CNAME record that points your domain to Olares:

- Type: `CNAME`
- Name: `website`
- Value: `laresprime.olares.com`

If the same name already has other records (for example an old A record), delete them first. A CNAME cannot coexist with other records on the same name.

### 3. Verify

The agent verifies the certificate chain, binds the domain, and polls until the platform marks the CNAME as active, then checks the site end to end over HTTPS. When it reports the domain is live, open `https://website.bellame.online` in your browser to confirm, and send the URL to the people you want to share it with.

:::tip Certificate renewal
Let's Encrypt certificates are valid for 90 days. When you renew, ask the agent to re-bind the domain with the new certificate. The app restarts briefly during the re-bind.
:::

## Common issues

### Installation gets stuck in `Initializing`

Common causes:

- The image architecture does not match the Olares node.
- The image tag was reused and the old layer is cached on the node.
- The chart version was not bumped before re-uploading.

Try uninstalling, deleting the old version, bumping both versions, and reinstalling:

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
