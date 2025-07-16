# Decentralized Identifier (DID)

Decentralized Identifiers (DIDs), as defined by the [W3C](https://www.w3.org/TR/did-core/), are self-sovereign globally unique identifiers that are verifiable without centralized registry. A DID can refer to any subject (e.g., a person, organization, thing, data model, abstract entity, etc.) as determined by its controller.

The primary objective of DIDs is to empower individuals and organizations with control over their identity information, enabling them to selectively and securely share this information without reliance on centralized authorities.

## Structure of a DID

A DID is structured as a text string that comprises three distinct components:

- **DID URI Scheme Identifier**: This specifies the identifier's format.
- **DID Method Identifier**: This indicates the specific protocol or method used to create the DID.
- **DID Method-Specific Identifier**: This is a unique identifier specific to the method used.

![DID Structure](/images/manual/concepts/did.png)

## DID derivation

In the context of the Olares ID, DIDs are self-generated using a mnemonic-based algorithm similar to blockchain addresses. The derivation process follows this sequence:

> Mnemonic -> Private Key -> Public Key -> Blockchain Address on DID

1. **Mnemonic**: A randomly generated string consisting of 12 words that serves as the basis for generating the private key.
2. **Private Key**: Derived from the mnemonic, this key is kept confidential and is essential for creating the public key and blockchain address.
3. **Public Key and Blockchain Address**: Both are generated from the private key and can be shared publicly.

::: warning IMPORTANT
It is crucial to keep the private key and mnemonic secure and confidential to prevent unauthorized access and ensure recoverability.
:::


