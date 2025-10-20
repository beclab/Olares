# Identity Wallet App

A Digital Identity Wallet is an essential tool in the self-sovereign identity ecosystem, allowing users to manage their decentralized identifiers (DIDs), credentials, and interactions with digital services, without relying on centralized entities for authentication or data sharing. 

LarePass is the identity wallet app for Olares, the decentralized self-hosted OS based on Olares ID. LarePass allows users easily and securely manage DIDs, Olares IDs, and enables seamless access to Olares.

## Manage Olares ID

When users creates a Olares account, a DID is automatically generated in the beginning and then bound to the new Olares ID. LarePass facilitates the Olares ID management with the following functions:

![Olares ID management](/images/manual/concepts/create-terminus-name.png)

- Olares IDs creation
  - Fast creation without VC binding
  - Advanced creation with VC binding (via the Gmail Issuer service)
- Backup/Import Olares IDs with a mnemonic phrase for quick setup and account recovery
- Manage multiple DIDs/Olares IDs 

See [Manage Accounts with LarePass](https://docs.olares.com/how-to/LarePass/account/) for more detailed information.

## Manage VCs

LarePass allows users to manage their VCs that are bound to Olares IDs, enabling users to interact with VCs through signing, verification, discovery, and presentation processes.

![VC management](/images/manual/concepts/vc-management.png)

::: tip NOTE
These are just the core implementations of LarePass that are closely related to the identity wallet. For more details on LarePass and its usages, refer to the [LarePass documentation](https://docs.olares.com/how-to/LarePass/overview.html).
