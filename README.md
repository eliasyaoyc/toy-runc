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

```bash
USAGE:
   runC [global options] command [command options] [arguments...]

COMMANDS:
   exec     exec a command into container
   init     Init container process run user's process in container. Do not call it outside
   ps       list all containers
   logs     print logs of a container
   rm       remove unused container
   run      create a container: my-docker run -ti [command]
   stop     stop a container
   network  container network commands
   commit   commit a container to image
   image    image commands
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```
Licensed under of either of

* MIT license ([LICENSE-MIT](LICENSE) or http://opensource.org/licenses/MIT)
