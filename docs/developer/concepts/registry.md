# DID Registry on Blockchains

Once a DID is generated, users need to register their ownership in a **DID Registry**. Although the W3C standards do not prescribe a specific implementation for this registry, Olares ID utilizes [Smart Contracts](/developer/contribute/olares-id/contract/contract.md#smart-contract.md) to facilitate the registration process. This approach offers several advantages:

- **Decentralization**: There is no reliance on centralized providers or authorities.
- **Censorship resistance**: The structure is robust against censorship and interference.
- **Universal discoverability**: DIDs can be easily discovered by users across the network.

Issuers can upload essential DID metadata to the Registry, including the DID itself, [Olares ID](olares-id.md), RSA public keys, and other relevant information.

DID metadata require minimal storage and infrequent updates. This allows existing mainstream blockchain technologies to manage these efficiently. Storing and retrieving DID metadata on blockchain results in low overhead and system efficiency. 

Currently, Olares ID's smart contracts are deployed on [OPTIMISTIC ROLLUPS](https://optimism.io/), an Ethereum Layer 2 blockchain.

::: info
It is noteworthy that the Web5 development team has transitioned from utilizing ION as their registry mechanism to adopting a Distributed Hash Table (DHT) network. DHT networks provide a greater degree of decentralization compared to Layer 2 blockchain solutions. However, managing DHT networks involves complexities in maintaining efficiency and stability, as well as ensuring data integrity across a vast number of nodes.
:::