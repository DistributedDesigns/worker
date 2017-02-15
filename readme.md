Worker
====
Executes transactions for users. Stores the user account information. It's got a lot more going on, too.

### Hot tips!
#### Bypass the workload generator
```sh
cat workload.txt | grep QUOTE | head -n 10 | xargs -d '\n' redis-cli -h localhost -p 44431 RPUSH worker:1:pendingtx
```
- MacOS (BSD) `xargs` doesn't have `-d` option. Install `findutils` from Homebrew then substitute `gxargs`
