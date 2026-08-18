---
outline: [2, 3]
description: Deploy a website you have already built to Olares from an AI agent such as OpenCode, using the Portfolio Landing Page project as an example.
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, Claude Code, deploy website, custom domain, image registry, preview, portfolio
---

# Deploy a website to Olares from an AI agent

Olares lets you deploy a website you have already built directly from an AI coding agent. You point the agent at your project, and it packages the site as an Olares app, pushes the image, installs it on your Olares device, and binds a custom domain so other people can reach the site over HTTPS.

This tutorial uses **OpenCode** and the **Portfolio Landing Page** project as running examples. You can follow the same steps with Claude Code or your own website project.

## Learning objectives

By the end of this tutorial, you will learn how to:

- Prepare project source code for deployment.
- Preview the website through a temporary preview link before deployment.
- Build and deploy the website to Olares as an app.
- Configure a custom domain with an RSA certificate.

## Prerequisites

Before you begin, ensure you have:
- OpenCode installed on Olares with a local model connected, such as Qwen3.6-27B (llama.cpp).
- Olares CLI (v1.12.6 or later) and Agent Skills installed and logged in on your computer.
- A working website project with source code available.
- One of the following image registry options:
  - A Docker Hub account.
  - Access to the local Olares image registry.
- A custom domain and an RSA SSL certificate if you want to share the site publicly.

## Understand the deployment flow

The agent handles most of the packaging work, but the overall flow follows these stages:

1. **Source code**: The agent reads your project from your computer, Olares Files, or GitHub.
2. **Preview**: You run the site in OpenCode and open a preview URL to confirm it works.
3. **Build and deploy**: The agent creates a production build, Dockerfile, container image, and Olares chart, then installs the app.
4. **Share**: You bind a custom domain, upload an RSA certificate, and set the entrance to public.

## Step 1: Prepare your project source code

Tell the agent where your project lives. The steps differ depending on the source location.

:::info Example project
This tutorial uses the Portfolio Landing Page public repository at `https://github.com/arnobt78/Portfolio-Landing-Page-7-React-Frontend` as an example. Replace the URL and folder names with your own project.
:::

### Source code is on GitHub

Tell the agent the repository URL.
- For a public repository, paste the link and ask it to clone.
- For a private repository, give the agent a GitHub personal access token or an SSH key for authentication first as prompted.

Example:

```text
Clone the Portfolio Landing Page repository from `https://github.com/arnobt78/Portfolio-Landing-Page-7-React-Frontend` into my OpenCode workspace.
```

The agent clones the project and starts working in that folder.

### Source code is in Olares Files

OpenCode can open the project folder as a workspace directly. The agent runs build, package, and deploy commands from inside that workspace.

### Source code is on your computer

If the source code is on your computer, you need to make it available to OpenCode on Olares. You can do this yourself or ask the agent to do it:

- **Push it to GitHub** yourself, or ask the agent to create a repo and push it. Then give the repository URL to the agent.
- **Upload it to Olares Files** under `Home/Code` yourself, or ask the agent to copy it there. Then open that folder as a workspace in OpenCode.

## Step 2: Preview the website

After the project is ready, ask the agent for a preview. If the agent offers to install dependencies and start a dev server first, just confirm. Otherwise, tell it directly.

The agent will run the appropriate commands and return a temporary preview URL. Open the URL in your browser and ensure it works as expected.

Example:

```text
https://1f47cd9b0.laresprime.olares.com/__preview/5173/
```

## Step 3: Build and deploy to Olares

After you confirm the preview, ask the agent to deploy:

```text
The preview looks good, deploy to Olares
```

The agent handles the rest: production build, Dockerfile, image build and push, Olares chart, and installation. When it finishes, the app appears on the Launchpad.

If the installation gets stuck, see [Common issues](#common-issues).

### Choose an image registry

The agent needs a registry to store the image. If you do not tell it which registry to use, it may pick a default and not ask you. State your preference before the build starts.

- **Docker Hub**: Tell the agent your Docker Hub username. For example:

  ```text
  Push the image to Docker Hub under my username `{my-username}`.
  ```

  If Docker credentials already exist in `~/.docker/config.json`, the agent may default to Docker Hub and use that username without asking.

- **Local Olares registry**: If you are working in Olares, ask the agent to use the local registry:

  ```text
  Use the local Olares image registry
  ```

  The agent can often push to `mirrors.olares.com` without extra credentials.

### Avoid architecture mismatch

:::warning Check CPU architecture
If you build the image on an Apple Silicon Mac but your Olares device uses an AMD64 CPU, the container will crash with an `exec format error`.
:::

Make sure the image matches your Olares node architecture:

```text
Build the image for `linux/amd64` because my Olares node is AMD64
```

If you want the image to run on both ARM and AMD64 devices, ask for a multi-architecture build:

```text
Build a multi-architecture image for both ARM64 and AMD64
```

:::tip Change the image tag every time you rebuild
If you rebuild the image after a fix, use a new tag such as `0.1.0`, `0.1.1`, `0.1.2`. Olares may cache the old layer and keep running the broken image if you reuse a tag.
:::

:::tip Change the chart version every time you upload
If you change the chart and upload again, increase the version in `Chart.yaml` and `OlaresManifest.yaml`. Olares may reject or behave unexpectedly with a duplicate chart version.
:::

## Step 4: Configure a custom domain

If you want to use a domain you own, tell the agent the real subdomain.

Example:

```text
Set the entrance URL for this app to `website.bellame.online`
```

The agent first sets the entrance auth level to public, then asks for the TLS certificate and private key. If you do not have them yet, it gives you a certbot command that uses `--key-type rsa`:

Example:

```text
To register website.bellame.online, I need the TLS certificate and private key for this domain. Do you have these files available? I need:

1. Certificate file (full chain PEM, e.g., cert.pem)
2. Private key file (RSA PEM, e.g., key.pem)

If you don't have them yet, you can generate one with certbot:

certbot certonly -d website.bellame.online --key-type rsa

Note: RSA key type is required (certbot defaults to ECDSA, which won't work).
```

Follow the agent prompts to complete the remaining steps: add the CNAME record, generate the certificate, copy the certificate files to a location the agent or `olares-cli` can read, and upload them.

### Add a CNAME record

In your DNS provider, add a CNAME record that points your domain to Olares. For example:
- Name: Your subdomain, such as `website`
- Type: `CNAME`
- Value: `laresprime.olares.com`

### Get an RSA certificate

Olares requires an RSA certificate. The default certbot command requests an ECDSA certificate, which can break the BFL service.

1. Run certbot yourself on your computer:

   ```bash
   sudo certbot certonly --manual --preferred-challenges dns --key-type rsa -d <your-domain>
   ```

   Example:

   ```bash
   sudo certbot certonly \
   --manual \
   --preferred-challenges dns \
   --key-type rsa \
   -d website.bellame.online
   ```

2. Enter your computer password as prompted.

   Example output:

   ```txt
   Saving debug log to /var/log/letsencrypt/letsencrypt.log
   Requesting a certificate for website.bellame.online
   - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
   Please deploy a DNS TXT record under the name:

   _acme-challenge.website.bellame.online.

   with the following value:

   ZqY6WOHPEBpXKwFmIf0LuhYo2jpVhuCmFv9VYSUotfw

   Before continuing, verify the TXT record has been deployed. Depending on the DNS
   provider, this may take some time, from a few seconds to multiple minutes. You can
   check if it has finished deploying with aid of online tools, such as the Google
   Admin Toolbox: https://toolbox.googleapps.com/apps/dig/#TXT/_acme-challenge.website.bellame.online.
   Look for one or more bolded line(s) below the line ';ANSWER'. It should show the
   value(s) you've just added.
   - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
   Press Enter to Continue
   ```

3. Add the provided DNS TXT record in your DNS managemetn console as prompted.
   - Name: `_acme-challenge.website`
   - Type: `TXT`
   - Value: `ZqY6WOHPEBpXKwFmIf0LuhYo2jpVhuCmFv9VYSUotfw`

4. Return to the terminal and press **Enter**.

   Example output:

   ```txt
   Successfully received certificate.
   Certificate is saved at: /etc/letsencrypt/live/website.bellame.online/fullchain.pem
   Key is saved at:         /etc/letsencrypt/live/website.bellame.online/privkey.pem
   This certificate expires on 2026-11-16.
   These files will be updated when the certificate renews.
   ```

### Upload the certificate

On a local computer, certbot saves certificates under `/etc/letsencrypt/live/<your-domain>/`. `olares-cli` cannot read them without root permissions.

1. Copy them to a location your user account can read:

   ```bash
   sudo cp /etc/letsencrypt/live/<your-domain>/fullchain.pem ~/cert.pem
   sudo cp /etc/letsencrypt/live/<your-domain>/privkey.pem ~/key.pem
   sudo chown $(whoami) ~/cert.pem ~/key.pem
   ```

   Example:
   
   ```bash
   olares-cli settings apps domain set portfolio7 portfolio7 \
   --third-party website.bellame.online \
   --cert-file ~/cert.pem \
   --key-file ~/key.pem
   ```

2. Run the domain binding command on your local machine:

   ```bash
   olares-cli settings apps domain set <app-name> <entrance-name> \
     --third-party <your-domain> \
     --cert-file ~/cert.pem \
     --key-file ~/key.pem
   ```

   Example:

   ```bash
   olares-cli settings apps domain set portfolio7 portfolio7 \
     --third-party website.bellame.online \
     --cert-file ~/cert.pem \
     --key-file ~/key.pem
   ```

   Example output:

   ```txt
   updated domain setup for portfolio7/portfolio7
   third-level: -
   third-party: website.bellame.online
   cert:        (set)
   key:         (set)
   laresprime@laresprimedeMacBook-Pro ~ % olares-cli settings apps domain get portfolio7 portfolio7
   App:                   portfolio7
   Entrance:              portfolio7
   Third-level domain:    -
   Third-party domain:    website.bellame.online
   CNAME status:          unset
   CNAME target:          laresprime.olares.com
   CNAME target status:   unset
   Cert configured:       yes
   Key configured:        yes
   ```

### Verify the entrance

Ask the agent to verify the domain is reachable:

```txt
The custom domain has been bound. Please verify that `website.bellame.online`
is reachable over HTTPS.
```

The agent should report output similar to:

```text
The domain is now bound and reachable. You can visit https://website.bellame.online in your browser to view your portfolio site.
```

## Step 5: Share the website

Open your custom domain in a browser and confirm the site works as expected. Then send the URL to the people you want to share it with.

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

### Auth level reverts to private after resuming the app

If you resume the app after changing the authentication level, the chart's `authLevel` setting may overwrite your change. Edit `OlaresManifest.yaml` in the chart source to set `authLevel` to `public`, bump the version, and redeploy.

### Agent cannot read certbot certificates

Certbot stores certificates under `/etc/letsencrypt` with root-only access. Copy the files to your project directory and change ownership before the agent reads them.
