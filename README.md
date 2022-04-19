# ToyRunC 


# Features
* Implement the basic runC functionality. 

# Build
* First, install `x86_64-linux-musl-gcc`
```
arch -x86_64 brew install FiloSottile/musl-cross/musl-cross
```

* Second, build it.
```bash
CC=x86_64-linux-musl-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64  go build -ldflags "-linkmode external -extldflags -static" -o runC /cmd/main.go
```

# Usage
**Ensure that the operating system runs linux.**
* run a sample bash.
```bash
./runC run -it ls -l

./runC run -it bin/sh
```

* run with resource limit(cgroup)
> **Optional:** you can choose stress in the usage below to start a resource hog program or provide it yourself
> 
> 💡 if you choose stress then you can use command `yum install stress` install it. 
```bash
# Limit memory
./runC run -it -m 100m stress --vm-bytes 200m --vm--keep -m 1
# Limit cpu ratio
./runC run -it -cpushare 512 stress --vm-bytes 200m --vm--keep -m 1
# Limit cpu 
./runC run -it -cpu 1 stress --vm-bytes 200m --vm-keep -m 1
```
Licensed under of either of

* MIT license ([LICENSE-MIT](LICENSE) or http://opensource.org/licenses/MIT)
