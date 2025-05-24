Checkpoints to Cover
Real-time location updates: Accept and process user location updates.

Store user locations: Efficiently store and update user coordinates in Redis Geo.

Instant geo queries: Provide endpoints to fetch users within a given radius.

Continuous geo watcher: Enable subscriptions for interest-based or intent-based nearby user alerts.

Filtered geo search: Support location queries with additional filters (interests, intent, availability).

Publish location events: Emit location update events to pub/sub (e.g., Redis Streams) for other services.

Data TTL: Ensure location data is ephemeral and automatically expires after a set period.

