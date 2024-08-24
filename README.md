# dialogue-service

* [1 Prerequisites](#1-prerequisites)
* [2 Prepare the environment](#2-prepare-the-environment)
* [3 Run the service](#3-run-the-service)
  * [3.1 Run locally](#31-run-locally)
  * [3.2 Run using Docker](#32-run-using-docker)
* [3 Scaling sharded database](#3-scaling-sharded-database)

## 1 Prerequisites

* [Go](https://go.dev/) (`v1.22` or later)
* [Docker](https://www.docker.com/)
* [goose](https://github.com/pressly/goose) (for running migrations)

`goose` can be installed by the following command (Go language must be already installed on the machine):

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

`goose` can also be installed using `brew` (for MacOS):

```bash
brew install goose
```

## 2 Prepare the environment

#### Step 1 - Up containers

Run the following command to up all infrastructure containers.

```bash
docker compose -f ./docker-compose-infra.yml up -d
```

#### Step 2 - Run migrations

Execute the commands below to apply migrations:

```bash
goose -dir ./migrations/ postgres "host=localhost port=25432 user=postgres dbname=postgres" up
```

## 3 Run the service

Depending on your preferences, you can run the service using one of the following ways:

* Run locally (Go must be installed)
* Run using Docker

### 3.1 Run locally

Run the following command:

```bash
go run ./cmd/app/main.go --config ./config/local.yml
```

### 3.2 Run using Docker

Run the following command:

```bash
docker compose -f ./docker-compose-service.yml up -d
```

## 3 Scaling sharded database

> NOTE
>
> By default WAL level is `replica` which should be changed to `logical`. Provided docker compose file sets WAL level to `logical` at database initialization step.

The database for dialogues is sharded by Citus. You can scale the number of workers, DB instances that store dialogue messages, using the following command:

```bash
docker compose -f .\docker-compose-infra.yml up -d --scale worker=3
```
Make sure that the master node can see new workers. Go inside the container and perform the commands bellow.

```bash
docker exec -it social-network-service-master pgsql -U postgres
select master_get_active_worker_nodes();
```

The result should look like below:

```
     master_get_active_worker_nodes     
----------------------------------------
 (social-network-service-worker-1,5432)
 (social-network-service-worker-2,5432)
 (social-network-service-worker-3,5432)
```

Then it is necessary to run the following command to run rebalancing process.

```
select citus_rebalance_start();
```

You can observe the status of rebalancing executing the following query.

```
select * from citus_rebalance_status();
```

When rebalancing is finished successfully, this query return the response like this (the format of the response is changed to make it more readable).

```
-[ RECORD 1 ]-------------------------------------------------
job_id      | 1
state       | finished
job_type    | rebalance
description | Rebalance all colocation groups
started_at  | 2024-07-21 09:39:47.47983+00
finished_at | 2024-07-21 09:40:24.939598+00
details     | {"tasks": [], "task_state_counts": {"done": 21}}
```