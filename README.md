# Kafka-Lite: High-Throughput Messaging System

A Kafka-inspired messaging system in Go with persistent TCP connections, custom message framing, producer/consumer routing, consumer groups, partitioned queues, and file-backed offset recovery.

## Highlights

- Built a broker that coordinates producers, consumers, topics, consumer groups, and partitions.
- Implemented a compact TCP protocol for registration, publishing, delivery, and acknowledgements.
- Added partitioned queue storage with fixed-size buffers for predictable memory usage.
- Supported concurrent consumer-group delivery with partition-level synchronization.
- Persisted queue offsets to recover topic, partition, and group progress after restart.
- Local benchmark: **470K+ msgs/sec**, **1–3% protocol overhead**, and **10K+ buffered messages per partition**.

## Architecture

The system runs in three modes:

```bash
go run . server
go run . producer <producer_port> <topic_id>
go run . consumer <consumer_port> <topic_id> <group_id>
```

The broker receives producer and consumer registrations, creates topics and consumer groups as needed, routes published messages into partition queues, and delivers messages to consumers in each group.

### Flow 1: High-Level Architecture

![High-Level Architecture](https://github-production-user-asset-6210df.s3.amazonaws.com/130151173/607518574-d20f93bd-35c6-45ee-9200-7b0456c12b4f.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAVCODYLSA53PQK4ZA%2F20260613%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20260613T171300Z&X-Amz-Expires=300&X-Amz-Signature=658390108d926103f50c5b3056ebc9677c2aee02fb0106aedb410b2897dac983&X-Amz-SignedHeaders=host&response-content-type=image%2Fpng)


### Flow 2: Producer Publish Flow

![Producer Publish Flow](https://github-production-user-asset-6210df.s3.amazonaws.com/130151173/607519094-54c760ab-1e1b-4a76-8ac3-73507ec71171.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAVCODYLSA53PQK4ZA%2F20260613%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20260613T171934Z&X-Amz-Expires=300&X-Amz-Signature=933a622bb672d3d7f7c5e91646f5505638c36b2a63fe063b2ac7826c82daf0f7&X-Amz-SignedHeaders=host&response-content-type=image%2Fpng)

### Flow 3: Consumer Delivery and Recovery Flow

![Consumer Delivery and Recovery Flow](https://github-production-user-asset-6210df.s3.amazonaws.com/130151173/607519147-9d023401-d7a3-4eed-9543-80ac6b55a763.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAVCODYLSA53PQK4ZA%2F20260613%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20260613T172014Z&X-Amz-Expires=300&X-Amz-Signature=6c1e76d8ff274485afea794dfeb285274d32d6d6066adbbf6c5e928b60f8063c&X-Amz-SignedHeaders=host&response-content-type=image%2Fpng)

## Core Components

| Component | Role |
|---|---|
| Broker | Coordinates registration, topic routing, partition assignment, and message delivery. |
| Producer | Publishes messages to a topic through a persistent TCP connection. |
| Consumer | Receives messages through a consumer group and assigned partition. |
| Topic | Logical stream that producers publish into. |
| Consumer Group | Allows multiple consumers to share work while preserving group-level progress. |
| Partition | Owns a synchronized queue for concurrent delivery. |
| Queue | Fixed-capacity file-backed circular buffer with persisted head/tail offsets. |

## Message Protocol

Each TCP message uses a compact length-prefixed frame:

```text
+-------------+---------------+------------------+
| 1B Length   | 1B Msg Type    | Payload Bytes    |
+-------------+---------------+------------------+
```

## Storage Design

Each partition uses a fixed-capacity file-backed queue:

```go
maxMessageSize = 255
queueCapacity  = 10000
```

The queue tracks:

- `head`: next message to consume
- `tail`: next write position
- payload storage
- message-size storage
- persisted metadata

This keeps memory usage predictable and allows broker progress to recover after restart.

## Getting Started

```bash
git clone https://github.com/quocbaodo205/kafka_go.git
cd kafka_go
go mod tidy
```

Start the broker:

```bash
go run . server
```

Start a producer:

```bash
go run . producer 10001 1
```

Start a consumer:

```bash
go run . consumer 10002 1 1
```
