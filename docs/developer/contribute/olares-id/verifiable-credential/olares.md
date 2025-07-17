# Use case with Olares

While DIDs solve the issue of identity in a decentralized network, they are typically difficult for humans to remember or use in daily situations. Olares ID provides a familiar, easy-to-remember format similar to email addresses, while still leveraging the power and security of DIDs. Each Olares ID is bound to a DID.

## Potential fairness issues

The **first-come, first-served** system in Olares ID registration may present some fairness issues, such as:

1. **Fraud:** For instance, the Olares ID `elonmusk@myterminus.com` might be registered by someone who isn't actually Elon Mask.
2. **Speculation:** Pre-registration of popular names could lead to speculation, potentially boosting early network activity, but at the cost of fairness.

## VC process for Olares ID 

To address the potential faireness concerns, we adopted principles from **Self-Sovereign Identity (SSI)** services proposed by the Web5 team, along with the [VC process](/manual/concepts/vc.md#verification-process) of Olares ID. This led us to design an **Issuer and Verifier** process to assist users in applying for a **Olares ID**.

![alt text](/images/developer/contribute/vc-process.jpeg)

### Gmail issuer service

We utilize Google's OAuth process to facilitate the issuance of **Verifiable Credentials (VCs)**. The simplified process is as follows:

1. Alice logs into her Gmail account via OAuth in LarePass, the wallet client. 
2. Google returns the OAuth credentials to LarePass.
3. LarePass submits the OAuth credentials to the Issuer.
4. The Issuer confirms the validity of the credentials with Google's servers and retrieves basic information (e.g., email name).
5. The Issuer issues a VC to Alice that matches the local part of her Gmail address.

Alice can now store the issued VC in LarePass.

:::tip NOTE
- Throughout the process, Alice only reveals basic account data within the scope of the credential authorization to LarePass and the Issuer service, with password and privacy protection ensured by Google's OAuth protocol.
- All the code for setting up a Gmail Issuer Service or other Web2 service Issuer Services are open sourced on GitHub.
:::

### Olares ID verifier service

Here's how the **Verifier Service** works on the Olares end:

1. Alice packages her DID, Olares ID, and Gmail VC into a **Verifiable Presentation (VP)** and submits the VP with its signature to the Verifier Service.
2. The Verifier Service checks:
    - The signature's validity.
    - The validity of the VC in the VP.
    - Whether the Olares ID can be registered on the blockchain (conflicts may arise if multiple channels, such as Gmail and Twitter, are used for VC information).
3. After all checks pass, the Verifier Service submits Alice's information to the blockchain and covers the Gas fees.

At this point, Alice successfully obtains her **Olares ID**. For example, if you apply with the Gmail address "hello@gmail.com", you'll receive the Olares ID "hello@myterminus.com" once all checks are completed.