Checkpoints to Cover:

WebSocket connection handling: Accept and manage client WebSocket connections, including authentication via JWT.

Message routing: Parse incoming messages, determine action/type, and route to the correct backend service (location, matching, chat, notification) using gRPC.

Protocol Buffers: Encode/decode all WebSocket messages using protobuf for efficient, typed communication.

Bi-directional streaming: Maintain real-time, two-way communication between clients and backend services.

Multiplexing: Support multiple logical channels (e.g., location updates, chat, match notifications) over a single WebSocket connection.
