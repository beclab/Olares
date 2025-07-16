---
outline: [2, 3]
description: Core principles of Olares account system, including synchronization mechanisms, account stages and unified authentication. Covers multi-factor authentication and multi-device sync fundamentals.
---

# Olares account

This document covers concepts and designs related to account system in Olares.

## Account synchronization

Accounts in LarePass, Olares, and Olares Space stay synchronized as described below:

- Creating an Olares requires providing an Olares ID and activate it using the LarePass logged in with that Olares ID.
- To log into Olares Space, you need to scan a QR code with LarePass.

## Understand the stage of account

Each account has three stages.

### Not bound to an Olares ID (DID stage)
An unbound account represents the initial stage where you have basic credentials created locally. This includes your mnemonic phrase, private key, and DID, but no Olares ID yet. 

During this stage, you can export and back up your mnemonic phrase and access Olares Space to request an organization domain name. 

However, importing to other LarePass clients isn't possible at this point.
:::tip
In the LarePass app, when you tap **Create an account**, your account enters the DID stage.
:::
### Bound to an Olares ID
When your account is bound to an Olares ID, the system records the connection between your Olares ID and DID on the blockchain.

This enables you to request and activate an Olares through Olares Space. 

At this stage, you gain the ability to import your account to other devices using your exported mnemonic phrase, supporting unified authentication across applications.

### Bound to an Olares
The final stage occurs when your account is linked to an Olares device. This enables full participation in the Olares ecosystem, including monitoring system resources for your device.

## Unified account system

Olares supports unified authentication for a multi-user system. 

1. After the user logs in on the login page, all future requests automatically include authentication details.
2. Each user request first goes through the Authelia service for authentication.
3. If authentication fails, the application redirects the user to the login page to re-authenticate.
4. If authentication succeeds, the [Backend for Launcher (BFL)](https://github.com/beclab/bfl) attaches the user's basic information and forwards the request to the application service. This relieves the application from handling the authentication itself.
5. For [shared applications](./application.md#shared-applications), developers need to build an additional `Auth Server` to connect the application's account with the BFL account.

## Multi-factor authentication (MFA)

Olares integrates a variety of authentication factors with different security levels to ensure the security of user identity authentication in the system.

### Password

When a user is first created, Olares generates a random password for initial setup. After completing identity verification, the user is prompted to replace this initial password with a stronger, custom password.

### One-time password

When users perform sensitive operations such as login, Olares requires users to enter the one-time two-factor authentication code generated in LarePass.

## Learn more

### Users

- [Create Olares ID](../get-started/create-olares-id)
- [User roles and permissions](../olares/settings/roles-permissions.md)

### Developers

- [Account system callback](../../developer/develop/advanced/account.md)