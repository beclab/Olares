# Verifiable Credential

DIDs represent both tangible or intangible entities across physical and digital realms. Each entity can have multiple claims, and the documents that support these claims are called **Verifiable Credentials (VCs)**. VCs are fully ratified W3C standard designed to work together with DIDs to enable trustless, secure interactions.

Consider this example: Alice, an entity, has educational qualifications that form a claim. Her diploma acts as the credential that verifies this claim. The diploma is issued by the university she attended and can be securely authenticated using cryptographic methods. When Alice applies for jobs, potential employers can verify this credential during interview.

## Roles related to VC

The following outlines the transformation of a diploma from a simple credential to a VC:

![Verifiable Credential Process](/images/manual/concepts/vc-diploma.jpeg)

This process involves three roles:

- **Issuer**

  > A role an entity can perform by asserting claims about one or more subjects. It creates a verifiable credential from these claims, and transmit the verifiable credential to a holder.

  In the above example, the Issuer is the university Alice attended.

- **Holder**

  > Holders are entities that have one or more verifiable credentials in their possession. Holders are also the entities that submit proofs to Verifiers to satisfy the requirements described in a Presentation Definition.

  In this case, the Holder is Alice. She stores and manages her VCs using a wallet app, which is TermiPass in this case.

- **Verifier**

  > Verifiers are entities that define what proofs they require from a Holder via the Presentation Definition in order to proceed with an interaction.

  The Verifier in this scenario is the company Alice interviews with.

## Verification process

The verification process involves six structured steps:

1. The Issuer registers their information on the DID Registry.
2. The Holder submits a verification request to the Issuer, indicating the need for credential issuance.
3. The Issuer issues a VC to the Holder, embedding a claim about her educational qualifications.
4. The Holder securely stores this VC in TermiPass, ensuring its availability for future verification.
5. The Holder sends a **Verifiable Presentation (VP)** that encapsulates the VC to a Verifier, initiating the verification process.
6. The Verifier checks the authenticity of the VC's and VP's signatures via the DID Registry, confirming the validity of the claim, thereby completing the verification.

::: tip NOTE
VCs are not stored on the blockchain. 
:::