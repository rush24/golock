# Golock

Go lib implements distributed lock based on redis of single node, you can also try [redsync](github.com/go-redsync/redsync) for high availability with multiple redis nodes.

## Usage

```
// recommend to use client globally for sharing redis connection pool
cli := mutex.NewClient("127.0.0.1:6379", SetNodeNum(1))
	
if m := cli.NewMutex(key, time.Second); m.Lock() {
	defer m.Unlock()
	// biz logic ...
}
```

## Requirement

Redis server 2.6.12 or later for the `SETNX` command with expiration

## Features

**Matually Exclusive & Avoiding Deadlock**: using redis `SETNX` with atomic expiration  
**Safety**: using lua script to guarantee unlocking is in a transaction and safe  
**Correctness**: using [snowflake](github.com/bwmarrin/snowflake) uuid to guarantee unlocking is correct  
