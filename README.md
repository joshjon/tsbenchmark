# ⚡ Timescale Benchmark

Implement a command line tool that can be used to benchmark `SELECT` query performance across multiple workers/clients
against a TimescaleDB instance. The tool should take as its input a CSV file (whose format is specified below) and a
flag to specify the number of concurrent workers. After processing all the queries specified by the parameters in the
CSV file, the tool should output a summary with the following stats:

- Number of queries processed
- Total processing time across all queries
- Minimum query time (for a single query)
- Median query time
- Average query time
- Maximum query time

## 🧰 Prerequisites

- Go 1.18.1
- Docker 20.10.10 CE

## 🚀 Running

1. Start and migrate TimescaleDB:
   ```
   docker-compose -p timescaledb up -d --build
   ```
2. Build tsbenchmark image:
   ```
   docker build -t local/tsbenchmark .
   ```
3. Run tsbenchmark:
   > ℹ️ Use `-h` flag to output help for all available options.
   ```
   docker run --rm --name tsbenchmark local/tsbenchmark filename.csv -m 10
   ```
4. Stop TimescaleDB
   ```
   docker-compose -p timescaledb down
   ```

## 🔬 Testing

- Unit tests: `make unit`

## 🔍 Design

#### High level flow chart

```mermaid
flowchart LR;
    Client--queue task-->pool;
    subgraph worker pool
        pool
        w1
        w2
        w...n
    end
    pool--queue unallocated task-->task_queue;
    pool--queue allocated task-->w1;
    pool--queue allocated task-->w2;
    pool--queue allocated task-->w...n;
    task_queue--receive task-->w1;
    task_queue--receive task-->w2;
    task_queue--receive task-->w...n;

```

#### Low level sequence diagram

```mermaid
sequenceDiagram
   participant client as Client (main)
   participant pool as Pool
   participant task_queue as Task Queue
   participant worker as Worker... n
   participant file as CSV File
   participant db as Timescale DB
   
    par read and queue
     loop while not EOF
        client->>file: read query
        file-->>client: query
        client-)pool: submit query task to pool wait queue
    end
    and pool dispatch
        loop while len(pool_wait_queue) > 0
            pool->> pool: receive task from wait queue
            pool->>worker: find worker for task route key (host name)
            alt worker found
               worker-->>pool: worker
               pool-)worker:queue task
            else
                 pool-)worker: start new
                 pool->>task_queue: queue task
            end
        end
    and worker processing
        loop while len(worker_queue) > 0 or task_queue open
            alt len(worker_queue) > 0
                worker->>worker: receive task  
            else
                task_queue->>worker: receive task
                worker->>worker: allocate route key (host name)
            end
               worker->>db: perform query task
               activate db
                  db-->>worker: result
               deactivate db
               worker->>worker: process result benchmark
        end
    end
    worker--)pool: send completion result
    pool--)client: send completion result
```