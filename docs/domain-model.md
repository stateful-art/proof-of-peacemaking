# Domain Model

## Entity Relationships

```mermaid
classDiagram
    User "1" -- "*" Expression
    User "1" -- "*" Acknowledgement
    Expression "1" -- "*" ExpressionAcknowledgement
    Acknowledgement "1" -- "*" ExpressionAcknowledgement
    ExpressionAcknowledgement "1" -- "1" ProofNFT
    User "1" -- "*" ProofNFT
    User "1" -- "*" UserNotification
    Notification "1" -- "*" UserNotification
    ExpressionAcknowledgement "1" -- "1" ProofRequest

    class User {
        +address: string
        +nonce: number
        +createdAt: datetime
        +updatedAt: datetime
        +subsidizedOperations: string[]
    }

    class Expression {
        +id: ObjectId
        +creator: User
        +content: MediaContent
        +ipfsHash: string
        +onChainId: number
        +status: string
        +createdAt: datetime
        +updatedAt: datetime
        +transactionHash: string
    }

    class Acknowledgement {
        +id: ObjectId
        +acknowledger: User
        +content: MediaContent
        +ipfsHash: string
        +onChainId: number
        +status: string
        +createdAt: datetime
        +updatedAt: datetime
        +transactionHash: string
    }

    class ExpressionAcknowledgement {
        +id: ObjectId
        +expression: Expression
        +acknowledgement: Acknowledgement
        +nftStatus: string
        +createdAt: datetime
        +updatedAt: datetime
        +nftRequestedBy: User
        +nftRequestedAt: datetime
        +nftApprovedBy: User
        +nftApprovedAt: datetime
    }

    class MediaContent {
        +text: string
        +textIpfs: string
        +audio: string
        +audioIpfs: string
        +video: string
        +videoIpfs: string
        +image: string
        +imageIpfs: string
    }

    class ProofNFT {
        +id: ObjectId
        +tokenId: number
        +expressionAcknowledgement: ExpressionAcknowledgement
        +creator: User
        +acknowledger: User
        +ipfsHash: string
        +status: string
        +createdAt: datetime
        +mintedAt: datetime
        +transactionHash: string
    }

    class Notification {
        +id: ObjectId
        +type: string
        +title: string
        +message: string
        +data: Object
        +createdAt: datetime
    }

    class UserNotification {
        +id: ObjectId
        +user: User
        +notification: Notification
        +read: boolean
        +readAt: datetime
        +createdAt: datetime
    }

    class ProofRequest {
        +id: ObjectId
        +expressionAcknowledgement: ExpressionAcknowledgement
        +requestedBy: User
        +requestedAt: datetime
        +approvedBy: User
        +approvedAt: datetime
        +status: string
        +createdAt: datetime
        +updatedAt: datetime
    }
```

## MongoDB Schemas

### User Schema
```javascript
const UserSchema = new Schema({
    address: {
        type: String,
        required: true,
        unique: true,
        lowercase: true
    },
    nonce: {
        type: Number,
        default: () => Math.floor(Math.random() * 1000000)
    },
    subsidizedOperations: [{
        type: String,
        enum: ['EXPRESSION', 'ACKNOWLEDGEMENT', 'NFT']
    }],
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now }
});
```

### Expression Schema
```javascript
const ExpressionSchema = new Schema({
    creator: {
        type: Schema.Types.ObjectId,
        ref: 'User',
        required: true
    },
    content: {
        text: String,
        textIpfs: String,
        audio: String,
        audioIpfs: String,
        video: String,
        videoIpfs: String,
        image: String,
        imageIpfs: String
    },
    ipfsHash: String,
    onChainId: Number,
    status: {
        type: String,
        enum: ['DRAFT', 'PENDING', 'CONFIRMED', 'FAILED'],
        default: 'DRAFT'
    },
    transactionHash: String,
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now }
});
```

### Acknowledgement Schema
```javascript
const AcknowledgementSchema = new Schema({
    acknowledger: {
        type: Schema.Types.ObjectId,
        ref: 'User',
        required: true
    },
    content: {
        text: String,
        textIpfs: String,
        audio: String,
        audioIpfs: String,
        video: String,
        videoIpfs: String,
        image: String,
        imageIpfs: String
    },
    ipfsHash: String,
    onChainId: Number,
    status: {
        type: String,
        enum: ['DRAFT', 'PENDING', 'CONFIRMED', 'FAILED'],
        default: 'DRAFT'
    },
    transactionHash: String,
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now }
});
```

### ExpressionAcknowledgement Schema
```javascript
const ExpressionAcknowledgementSchema = new Schema({
    expression: {
        type: Schema.Types.ObjectId,
        ref: 'Expression',
        required: true
    },
    acknowledgement: {
        type: Schema.Types.ObjectId,
        ref: 'Acknowledgement',
        required: true
    },
    nftStatus: {
        type: String,
        enum: ['NONE', 'REQUESTED', 'APPROVED', 'MINTING', 'MINTED', 'FAILED'],
        default: 'NONE'
    },
    nftRequestedBy: {
        type: Schema.Types.ObjectId,
        ref: 'User'
    },
    nftRequestedAt: Date,
    nftApprovedBy: {
        type: Schema.Types.ObjectId,
        ref: 'User'
    },
    nftApprovedAt: Date,
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now }
});

// Compound unique index to prevent duplicate acknowledgements
ExpressionAcknowledgementSchema.index(
    { expression: 1, acknowledgement: 1 },
    { unique: true }
);
```

### ProofNFT Schema
```javascript
const ProofNFTSchema = new Schema({
    tokenId: {
        type: Number,
        required: true,
        unique: true
    },
    expressionAcknowledgement: {
        type: Schema.Types.ObjectId,
        ref: 'ExpressionAcknowledgement',
        required: true
    },
    creator: {
        type: Schema.Types.ObjectId,
        ref: 'User',
        required: true
    },
    acknowledger: {
        type: Schema.Types.ObjectId,
        ref: 'User',
        required: true
    },
    ipfsHash: String,
    status: {
        type: String,
        enum: ['PENDING', 'MINTED', 'FAILED'],
        default: 'PENDING'
    },
    createdAt: { type: Date, default: Date.now },
    mintedAt: Date,
    transactionHash: String
});
```

### Notification Schema
```javascript
const NotificationSchema = new Schema({
    type: {
        type: String,
        enum: [
            'NEW_ACKNOWLEDGEMENT',
            'PROOF_REQUEST_RECEIVED',
            'PROOF_REQUEST_APPROVED',
            'NFT_MINTED',
            'EXPRESSION_CONFIRMED',
            'ACKNOWLEDGEMENT_CONFIRMED'
        ],
        required: true
    },
    title: String,
    message: String,
    data: {
        type: Map,
        of: Schema.Types.Mixed
    },
    createdAt: { type: Date, default: Date.now }
});
```

### UserNotification Schema
```javascript
const UserNotificationSchema = new Schema({
    user: {
        type: Schema.Types.ObjectId,
        ref: 'User',
        required: true
    },
    notification: {
        type: Schema.Types.ObjectId,
        ref: 'Notification',
        required: true
    },
    read: {
        type: Boolean,
        default: false
    },
    readAt: Date,
    createdAt: { type: Date, default: Date.now }
});

// Index for quick retrieval of user's unread notifications
UserNotificationSchema.index({ user: 1, read: 1 });
```

### ProofRequest Schema
```javascript
const ProofRequestSchema = new Schema({
    expressionAcknowledgement: {
        type: Schema.Types.ObjectId,
        ref: 'ExpressionAcknowledgement',
        required: true,
        unique: true
    },
    requestedBy: {
        type: Schema.Types.ObjectId,
        ref: 'User',
        required: true
    },
    requestedAt: { type: Date, default: Date.now },
    approvedBy: {
        type: Schema.Types.ObjectId,
        ref: 'User'
    },
    approvedAt: Date,
    status: {
        type: String,
        enum: ['PENDING', 'APPROVED', 'REJECTED', 'CANCELLED'],
        default: 'PENDING'
    },
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now }
});
```

## Status Flows

### Expression/Acknowledgement Status Flow
```mermaid
stateDiagram-v2
    [*] --> DRAFT
    DRAFT --> PENDING: Submit to blockchain
    PENDING --> CONFIRMED: Transaction confirmed
    PENDING --> FAILED: Transaction failed
    FAILED --> PENDING: Retry transaction
```

### NFT Status Flow
```mermaid
stateDiagram-v2
    [*] --> PENDING
    PENDING --> MINTED: Minting successful
    PENDING --> FAILED: Minting failed
    FAILED --> PENDING: Retry minting
```

### NFT Request Status Flow
```mermaid
stateDiagram-v2
    [*] --> NONE
    NONE --> REQUESTED: User requests NFT
    REQUESTED --> APPROVED: Other party approves
    APPROVED --> MINTING: Start minting
    MINTING --> MINTED: Minting successful
    MINTING --> FAILED: Minting failed
    FAILED --> MINTING: Retry minting
```

### Proof Request Status Flow
```mermaid
stateDiagram-v2
    [*] --> PENDING: Request initiated
    PENDING --> APPROVED: Other party approves
    PENDING --> REJECTED: Other party rejects
    PENDING --> CANCELLED: Requester cancels
    APPROVED --> [*]: NFT minting starts
``` 