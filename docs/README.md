# Technical Documentation

This directory contains technical documentation for the Proof of Peacemaking project.

For general project information, installation instructions, and overview, please see the [main README](../README.md).

## Contents

### Smart Contract Architecture

The project uses the Diamond Pattern for upgradeable smart contracts:

```
contracts/
├── Diamond.sol                 # Main diamond contract
├── facets/
│   ├── DiamondCutFacet.sol    # Handles upgrades
│   ├── DiamondLoupeFacet.sol  # Contract inspection
│   ├── ExpressionFacet.sol    # Expression functionality
│   ├── AcknowledgementFacet.sol # Acknowledgement functionality
│   ├── POPNFTFacet.sol        # NFT minting functionality
│   └── PermissionsFacet.sol   # Permission management
├── libraries/
│   ├── LibDiamond.sol         # Diamond storage & core functions
│   ├── LibStorage.sol         # Shared storage structure
│   └── LibPermissions.sol     # Permission & subsidy logic
└── interfaces/
    ├── IDiamondCut.sol        # Diamond upgrade interface
    └── IDiamondLoupe.sol      # Diamond inspection interface
```

### Data Flow Diagrams

#### Expression Creation and Acknowledgement
```mermaid
sequenceDiagram
    participant User1 as Expression Creator
    participant User2 as Acknowledger
    participant D as Diamond Proxy
    participant EF as Expression Facet
    participant AF as Acknowledgement Facet
    participant NF as NFT Facet
    participant IPFS

    Note over User1,IPFS: 1. Expression Creation
    User1->>IPFS: Upload content
    IPFS-->>User1: Return IPFS hash
    User1->>D: createExpression()
    D->>EF: delegate call
    Note right of EF: Check subsidization
    EF-->>User1: ExpressionCreated event

    Note over User2,IPFS: 2. Acknowledgment
    User2->>IPFS: Upload content
    IPFS-->>User2: Return IPFS hash
    User2->>D: createAcknowledgement()
    D->>AF: delegate call
    Note right of AF: Check subsidization
    AF-->>User2: AcknowledgementCreated event

    Note over User1,User2: 3. NFT Minting
    User1->>D: Sign for NFT
    User2->>D: Sign for NFT
    D->>NF: mintProofs()
    NF-->>User1: Mint NFT
    NF-->>User2: Mint NFT
```

### API Documentation

For detailed API documentation, see [API.md](API.md).

### Database Schema

For database schema documentation, see [DATABASE.md](DATABASE.md).

### Development Guidelines

For development guidelines and best practices, see [DEVELOPMENT.md](DEVELOPMENT.md).
