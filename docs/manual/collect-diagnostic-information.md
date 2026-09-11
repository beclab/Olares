---
outline: [2, 3]
description: Collect Olares system logs with the CLI, download the archive from the host, and share it safely.
head:
  - - meta
    - name: keywords
      content: Olares, technical support, collect logs, export logs, system logs, pod logs
---

# Collect diagnostic information

When a troubleshooting guide does not resolve an issue, collect the relevant context and logs for further investigation. If Settings is available and you only need a system log archive, you can [export system logs in Settings](help/request-technical-support.md) instead.

## Record the problem context

Before collecting logs, write down:

- The exact time the issue occurred and your time zone.
- Your Olares version and the affected app version.
- What you were doing, the expected result, and what happened instead.
- The shortest steps that reproduce the issue.
- The complete error message and a screenshot, if available.
- Any recent change, such as an Olares update, app update, network change, or restart.

This context helps match a log entry to the visible problem.

## Connect to the Olares host

Connect to the device with SSH. On Olares One, the SSH user is `olares`:

```bash
ssh olares@{local_ip_address}
```

For a self-installed Olares device, use the Linux account that has `sudo` access.

## Collect and save the logs

Run the following commands on the Olares host:

```bash
mkdir -p "$HOME/olares-logs"
sudo olares-cli logs --output-dir "$HOME/olares-logs"
```

The command collects recent Olares system, Kubernetes, container, kernel, and network diagnostics. When it finishes, the last line prints the full archive path, for example:

```plain
logs have been collected and archived in: /home/olares/olares-logs/olares-logs-20260827-050839.tar.gz
```

If the issue occurred recently, you can reduce the archive size by specifying a time range, for example `--since 3h`.

## Download the archive

On your own computer, copy the exact file printed in the previous step. Replace `{local_ip_address}` and `{timestamp}` with your values:

```bash
scp olares@{local_ip_address}:~/olares-logs/olares-logs-{timestamp}.tar.gz .
```

The final dot saves the archive to the current directory on your computer. If you use a different SSH account, replace `olares` with that account name.

## Share logs safely

:::warning Logs can contain private system information
System logs can include Olares identifiers, hostnames, IP addresses, network configuration, file paths, and application metadata. Do not attach the complete archive to a public GitHub Issue or Discussion.
:::

1. For a general question that does not require logs, open a [GitHub Discussion](https://github.com/beclab/Olares/discussions/new?category=q-a).
2. For a reproducible software bug, open a [GitHub Issue](https://github.com/beclab/Olares/issues/new) with the problem context, but leave out the full log archive and secrets.
3. When the Olares team requests logs, send them through the private channel provided to you or email [hi@olares.com](mailto:hi@olares.com). Include the related GitHub issue number, if there is one.

Before sending, remove passwords, recovery phrases, API keys, access tokens, and other secrets from your description and screenshots. The Olares team will never need your password or recovery phrase to diagnose a problem.
