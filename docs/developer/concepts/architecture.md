---
description: Overview of Olares BEC architecture, explaining distributed node implementation for data storage and security. Details three core components Olares ID, Olares OS and LarePass.
---
# Architecture of Olares

Olares introduces a next-generation decentralized Internet framework through its Blockchain-Edge-Client (BEC) architecture. BEC decentralizes data storage and enhances security by distributing information across suitable platforms.

![BEC](/images/manual/concepts/network-topology.jpeg)

- **Blockchain**: Storage of high-value data, including decentralized identifiers (DIDs) and transactions on the blockchain via smart contracts. This enables transparency, immutability, and enhanced discoverability of data.
- **Edge**: The decentralized web node for users, hosting personal data such as documents, chat logs, and photos on private edge servers. Data remains within the user’s control on the edge, ensuring privacy and local data sovereignty.
- **Client**: The identity wallet app that ensures users can securely manage their identities and interact with their self-hosted systems while maintaining ownership and privacy over their digital credentials.

## Core components of Olares

Corresponding to the BEC architecture, Olares comprises the following core components:

- [**Olares ID**](https://docs.snowinning.com/protocol/overview.html): A decentralized identity and reputation system that integrates decentralized identifiers (DIDs), verifiable credentials (VCs), and reputation data. It enhances trust by enabling transparent and verifiable interactions within decentralized environments.
- [**Olares OS**](https://github.com/beclab/Olares): A comprehensive, self-hosted operating system designed for edge devices. It allows users to host and manage their own data and applications, transforming personal edge devices into robust, sovereign cloud systems.
- [**LarePass**](https://www.olares.com/larepass): A secure, unified interface software that connects users to their Olares systems. It offers key functionalities, including identity management, remote access, device management, and data storage, ensuring seamless interactions with Olares.

## Learn more

- [Self-sovereign network](https://docs.snowinning.com/protocol/network.html)