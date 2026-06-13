# Kafka-Lite: High-Throughput Messaging System

A Kafka-inspired messaging system in Go with persistent TCP connections, custom message framing, producer/consumer routing, consumer groups, partitioned queues, and file-backed offset recovery.

## Highlights

- Built a broker that coordinates producers, consumers, topics, consumer groups, and partitions.
- Implemented a compact TCP protocol for registration, publishing, delivery, and acknowledgements.
- Added partitioned queue storage with fixed-size buffers for predictable memory usage.
- Supported concurrent consumer-group delivery with partition-level synchronization.
- Persisted queue offsets to recover topic, partition, and group progress after restart.
- Local benchmark results: **470K+ msgs/sec**, **1–3% protocol overhead**, and **10K+ buffered messages per partition**.

## Architecture

The system runs in three modes:

```bash
go run . server
go run . producer <producer_port> <topic_id>
go run . consumer <consumer_port> <topic_id> <group_id>
```

The broker receives producer and consumer registrations, creates topics and consumer groups as needed, routes published messages into partition queues, and delivers messages to consumers in each group.

### Flow 1: High-Level Architecture

![High-Level Architecture]()

> Placeholder: add a diagram showing Producer → Broker → Topic → Consumer Group → Partitions → Consumers.

### Flow 2: Producer Publish Flow

![Producer Publish Flow]()

> Placeholder: add a diagram showing producer registration, topic lookup/creation, message publish, broker routing, queue append, and acknowledgement.

### Flow 3: Consumer Delivery and Recovery Flow

![Consumer Delivery and Recovery Flow]()

> Placeholder: add a diagram showing consumer registration, group/partition assignment, message delivery, offset persistence, and restart recovery.

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
