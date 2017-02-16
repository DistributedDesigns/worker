Worker
====
[![Build Status](https://travis-ci.org/DistributedDesigns/worker.svg?branch=master)](https://travis-ci.org/DistributedDesigns/worker)

Executes transactions for users. Stores the user account information. It's got a lot more going on, too.

## Installing
```sh
git clone https://github.com/DistributedDesigns/worker.git

.scripts/install

# Assumes there's a RMQ and Redis instance running in docker
# Run with one of
$GOPATH/bin/worker -n 1
go run *.go -n 1
```
Runtime flags of interest
- `-l, --log-level` Sets the... log level.
- `-n, --worker-num` **Mandatory** ID for the worker. Yes, each worker should have a unique ID.
- `-a, --no-audit` Don't send messages to the audit server.

### Hot tips!
#### Bypass the workload generator
```sh
cat workload.txt | grep QUOTE | head -n 10 | xargs -d '\n' redis-cli -h localhost -p 44431 RPUSH worker:1:pendingtx
```
- MacOS (BSD) `xargs` doesn't have `-d` option. Install `findutils` from Homebrew then substitute `gxargs`
