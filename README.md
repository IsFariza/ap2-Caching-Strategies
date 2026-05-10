# AP2 Assignment 4 - Caching and Background Jobs

This project extends the Assignment 3 medical scheduling platform with Redis caching, Redis-backed gRPC rate limiting, notification job queue, and simulated external notification mock gateway. Domain models, protobuf contracts, PostgreSQL repositories, migrations, and the NATS event subjects are preserved.
Here is a simplified, student-friendly version of the README. It keeps all the technical requirements but explains them in a more "hands-on" way.



## How it Works 
Clients talk to the Doctor and Appointment services via gRPC.  
Redis handles fast storage for caching and tracking how many requests user makes.  
NATS acts as the messenger. When an appointment is finished, a message is sent to NATS.  
Notification Service listens to NATS and simulates to "email" the patient via a Mock Gateway.

## Cache Strategy

Doctor Service:
- GetDoctor: cache-aside using `doctor:id`. On miss, read PostgreSQL and best-effort `SET`.
- ListDoctors: cache-aside using `doctors:list`.
- CreateDoctor: write-through style decorator. The database write happens first, then `doctor:<id>` is refreshed and `doctors:list` is evicted.

Appointment Service:
- GetAppointment: cache-aside using `appointment:<id>`.
- `ListAppointments`: cache-aside using `appointments:list`.
- `CreateAppointment`: write-around. The row is written to PostgreSQL and only `appointments:list` is evicted.
- `UpdateAppointmentStatus`: write-through refresh of `appointment:<id>` plus eviction of `appointments:list`.

`CACHE_TTL_SECONDS` configures TTL; default is 60 seconds. Cache misses and Redis write failures never fail the gRPC request.

## Rate Limiting

Both gRPC services use a unary server interceptor. The implementation is a Redis sliding-window log per client IP using sorted sets:

`rate_limit:<service>:<client-ip>`

Each request removes entries older than 60 seconds with `ZREMRANGEBYSCORE`, counts the current window with `ZCARD`, and adds the current request timestamp with `ZADD` when the request is allowed. The key is expired after 60 seconds of inactivity. The default limit is 5 requests per minute and can be changed with `RATE_LIMIT_RPM`. Exceeded requests return `codes.ResourceExhausted` with a retry-after hint based on the oldest request still in the window.

Central Redis counters avoid two common horizontal-scaling problems: each service instance having a separate local counter, and clients bypassing limits by being routed to different instances.

## Job Queue

The Notification Service has three separated concerns:
- `internal/subscriber`: NATS subscription and routing.
- `internal/logger`: one structured JSON event log per received event.
- `internal/jobqueue`: buffered channel, worker pool, idempotency, gateway retry, and dead-letter logging.

When `appointments.status_updated` has `new_status = "done"`, the subscriber logs the event and enqueues a job. The job queue size is `WORKER_POOL_SIZE * 10`; worker count is configured by `WORKER_POOL_SIZE` and defaults to 3. If the channel is full, the job is written to stderr as a dead-letter entry.

Idempotency keys are SHA-256 hex strings of:

`event_type + id + occurred_at`

They are stored in Redis as `notification:idempotency:<key>` with a 24 hour TTL. Values move from `queued` to `done`. Replayed events with the same key are dropped without a second gateway call.

Gateway failures or network errors are retried up to 3 attempts with exponential backoff logs. After the final failure, the worker writes a structured JSON `dead_letter` line to stderr.

## Mock Gateway

Run from `mock-gateway` with:

```powershell
go run .
```

It exposes:

```http
POST /notify
```

Request:

```json
{"idempotency_key":"...","channel":"email","recipient":"patient@clinic.kz","message":"..."}
```

New keys return `{"status":"accepted"}`. Repeated keys return `{"status":"duplicate"}`. About 20% of requests return HTTP 503 to exercise retry logic. Every request is logged to stdout as JSON.

## Infrastructure Setup

Docker Compose starts NATS, Redis, both PostgreSQL databases, the three platform services, and the mock gateway:

```powershell
docker compose up --build
```

PostgreSQL must use the same `DATABASE_URL` values shown in `docker-compose.yml`.

## Environment Variables

- All services: `REDIS_URL`, `CACHE_TTL_SECONDS`
- Doctor/Appointment: `DATABASE_URL`, `NATS_URL`, `GRPC_PORT`, `RATE_LIMIT_RPM`
- Appointment only: `DOCTOR_ADDR`
- Notification: `NATS_URL`, `REDIS_URL`, `GATEWAY_URL`, `WORKER_POOL_SIZE`
- Mock Gateway: `GATEWAY_PORT`

## Startup Order

```powershell
docker compose up --build
```

If manually, start Redis, PostgreSQL, and NATS before the services.


## Consistency Trade-offs

Redis is treated as optional. If Redis is unavailable at startup, services log a warning and continue with PostgreSQL reads. Cached reads can be stale until TTL expiry if an invalidation fails, but write paths evict or refresh keys immediately after successful database writes. Redis Cluster would improve availability, but consistency still depends on key routing, failover behavior, and successful invalidation commands.

## Rate-limiting trade-offs 
Per-instance rate limiting has two major flaws:
- Because a load balancer may route a client to a different server for every request, no single instance has a complete view of the client's activity, leading to inaccurate blocking.
- If there are 5 instances and a limit of 100 RPM, a user can bypass the limit by sending 500 RPM distributed across all instances.