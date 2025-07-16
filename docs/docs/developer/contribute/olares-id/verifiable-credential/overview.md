# VC Service

The purpose of this documentation is to help you understand how to configure various schemas and create Issuers and Verifiers for your own scenarios using:
- [verifiable-credential-gate](https://github.com/Above-Os/verifiable-credential-gate) and [did-gate](https://github.com/Above-Os/did-gate) by Terminus.
- [SSI Service](https://github.com/TBD54566975/ssi-service) by the tbd team.

::: tip
If you want to engage in lower-level development, read the [protocol standard](#reference) together with the source code in [SSI SDK](https://github.com/TBD54566975/ssi-sdk).
:::

## Introduction

We have learned about [VC](/manual/concepts/vc.md) and the basic process for applying VCs.

Before we get into the implementation details, we can familiarize ourselves with the terms that we will encounter in the real communication process of Wallet, Verifier and Issuer.
| Term                    | Definition                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Holder                  | Holders are entities that have one or more verifiable credentials in their possession. Holders are also the entities that submit proofs to Verifiers to satisfy the requirements described in a Presentation Definition.                                                                                                                                                                                                                                                                                                        |
| Issuer                  | A role an entity can perform by asserting claims about one or more subjects, creating a verifiable credential from these claims, and transmitting the verifiable credential to a holder.                                                                                                                                                                                                                                                                                                                                        |
| Verifier                | Verifiers are entities that define what proofs they require from a Holder (via a Presentation Definition) in order to proceed with an interaction.                                                                                                                                                                                                                                                                                                                                                                              |
| Verifiable Credential   | Is a tamper-evident credential that has authorship that can be cryptographically verified. Verifiable credentials can be used to build Verifiable Presentations, which can also be cryptographically verified. The claims in a credential can be about different subjects. PEX accepts Verifiable credential in 3 forms: 1. JSON_LD which is known in our system as IVerifiableCredential, 2. JWT-Wrapped VC which is known in our system as JwtWrappedVerifiableCredential or string which is a valid Verifiable credential jwt |
| Verifiable Presentation | s a tamper-evident presentation encoded in such a way that authorship of the data can be trusted after a process of cryptographic verification.                                                                                                                                                                                                                                                                                                                                                                                 |
| Manifest                | Credential Manifests are used to describe which credentials are available for issuance.                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| Application             | The format provided by Holder to Issuer, including                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| Presentation Definition | Presentation Definitions are objects that articulate what proofs a Verifier requires.                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| Presentation            | Data derived from one or more verifiable credentials, issued by one or more issuers                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| Submission              | TBC                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| Definition              | TBC                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| Schema                  | All different Manifest, Application, Credential, Presentation, Definition need to define Schema with JSON. The service will verify the correctness of submitted data and then go into business process.                                                                                                                                                                                                                                                                                                                         |

## Reference

The following reference materials come from the [SSI SDK](https://github.com/TBD54566975/ssi-sdk) project.

### Specifications

Here are a set of references to specifications that this library currently supports. It is a dynamic set that will change as the library evolves.

- [Decentralized Identifiers (DIDs) v1.0](https://www.w3.org/TR/did-core/) W3C Proposed Recommendation 03 August 2021
- [Verifiable Credentials Data Model v1.1](https://www.w3.org/TR/vc-data-model/) W3C Recommendation 09 November 2021
  - Supports [Linked Data Proof](https://www.w3.org/TR/vc-data-model/#data-integrity-proofs) formats.
  - Supports [VC-JWT and VP-JWT](https://www.w3.org/TR/vc-data-model/#json-web-token) formats.
- [Verifiable Credentials JSON Schema Specification](https://w3c-ccg.github.io/vc-json-schemas/v2/index.html) Draft Community Group Report, 21 September 2021
- [Presentation Exchange 2.0.0](https://identity.foundation/presentation-exchange/) Working Group Draft, March 2022
- [Wallet Rendering Strawman](https://identity.foundation/wallet-rendering/), June 2022
- [Credential Manifest](https://identity.foundation/credential-manifest/) Strawman, June 2022
- [Status List 2021](https://w3c-ccg.github.io/vc-status-list-2021/) Draft Community Group Report 04 April 2022

### Signing Methods

> - [Data Integrity 1.0](https://w3c.github.io/vc-data-integrity/) Draft Community Group Report
> - [Linked Data Cryptographic Suite Registry](https://w3c-ccg.github.io/ld-cryptosuite-registry/) Draft Community Group Report 29 December 2020
> - [JSON Web Signature 2020](https://w3c-ccg.github.io/lds-jws2020/) Draft Community Group Report 09 February 2022
>   - [VC Proof Formats Test Suite, VC Data Model with JSON Web Signatures](https://identity.foundation/JWS-Test-Suite/) Unofficial Draft 09 March 2022 This implementation's compliance with the JWS Test Suite can be found here.
>   - Supports both JWT and Linked Data proof formats with [JOSE compliance](https://jose.readthedocs.io/en/latest/).
