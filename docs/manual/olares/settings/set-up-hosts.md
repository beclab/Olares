---
description: Configure and manage your Olares hosts file to customize domain name resolution, map IP addresses, and control access to services with step by step guidance.
---
# Set up hosts file

The hosts file is a system configuration file that maps domain names to specific IP addresses. By editing this file, you can customize how Olares resolves domain names, bypass the default DNS resolution, and control access to certain websites or services.
## Add a host
To add a new domain name and IP address mapping in Olares:
1. Open Settings, and go to **System** > **Hosts**.
2. Select **Add hosts** in the top-right corner, and specify the domain name/IP address pair.
   - **Host name**: Enter the domain name you want to map (e.g., `example.com`).
   - **IP**: Enter the corresponding IP address (e.g., `93.184.216.34`).
     ![Add a host](/images/manual/olares/add-host.png#bordered)
3. Click **Confirm** to save the changes.

:::info DNS cache delay
After editing the hosts file, the changes might not take effect immediately due to DNS caching.
:::

## Verify hosts file change
To ensure that the new domain-to-IP mapping is working as expected, you can use the `nslookup` command. This command allows you to check whether the domain resolves to the IP address you specified in the hosts file.

In your command prompt or terminal, run the following command, and replace `[domain-name]` with the domain you want to verify:
  ```shell
  nslookup [domain-name]
  ```

The output will include the resolved IP address for the domain. For example:
  ```shell
  Name:    example.com
  Address: 93.184.216.34
  ```
If the `Address` field matches the IP address you configured in the hosts file, it means the changes have taken effect.
