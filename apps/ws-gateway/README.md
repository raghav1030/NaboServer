# 🧭 WebSocket Gateway Service — Architecture & Contribution Guide

## 📌 Overview

The **WebSocket Gateway** serves as the real-time entry point for all client communications. It maintains persistent WebSocket connections with clients and acts as a central relay, routing messages to various backend microservices such as:

- **Location Service**
- **Chat Service**
- **Notification Service**

---

## 🧱 High-Level Architecture

```mermaid
graph TD
    Client[Client (WebSocket)] -->|Connects via WS| Gateway[WebSocket Gateway]
    Gateway -->|gRPC| LocationService
    Gateway -->|gRPC| ChatService
    Gateway -->|gRPC| NotificationService



Clients establish persistent WebSocket connections to the Gateway.

The Gateway decodes messages and forwards them to appropriate backend services via gRPC.

Backend responses are sent back to clients over the same WebSocket connection.


graph TD
    subgraph Gateway Internals
        A[Connection Map]
        B[Read Pump (goroutine)]
        C[Write Pump (goroutine)]
        D[Message Channel]
        E[gRPC Clients]
    end
    Client --> A
    A --> B
    A --> C
    B --> D
    D --> C
    B --> E
    C --> Client


Components:
Connection Map: Maintains a map of active WebSocket connections for broadcasting and managing lifecycle events.

Read Pump: Reads and decodes incoming WebSocket messages using a goroutine.

Write Pump: Sends outgoing messages from the backend to clients over WebSocket.

Message Channel: A per-connection buffered channel ensuring non-blocking writes.

gRPC Clients: Enables communication with microservices using strongly typed Protobuf contracts.


Message Flow

sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant R as ReadPump
    participant GRPC as gRPC Backend
    participant W as WritePump

    C->>G: Sends WebSocket message
    G->>R: Read & decode message
    R->>GRPC: Forward to backend via gRPC
    GRPC->>W: Return response
    W->>C: Send message to client
    C-->>G: Disconnect
    G->>G: Clean up connection


| Area                  | Responsibility                                    | File Path                       |
| --------------------- | ------------------------------------------------- | ------------------------------- |
| WebSocket Handler     | Upgrade HTTP to WS, initialize state, start pumps | `internal/handler/websocket.go` |
| Connection Management | Add/remove connections from the map               | `internal/service/gateway.go`   |
| Message Routing       | Parse and route messages to backend services      | `internal/service/gateway.go`   |
| Router Setup          | Register WebSocket endpoints and routes           | `internal/router/router.go`     |
| Proto Definitions     | Define message structures using Protobuf          | `libs/proto/`                   |
| gRPC Client Logic     | Initialize and invoke RPCs to microservices       | `internal/service/gateway.go`   |
| Utilities & Helpers   | Reusable logic and common utilities               | `pkg/utils/`                    |
```
