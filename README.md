# Toy Docker 


# Features
* Implement the basic Docker functionality. 

# Build
* First, install `x86_64-linux-musl-gcc`
```
arch -x86_64 brew install FiloSottile/musl-cross/musl-cross
```

* Second, build it.
```bash
CC=x86_64-linux-musl-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64  go build -ldflags "-linkmode external -extldflags -static" -o docker /cmd/main.go
```

# Usage


Licensed under of either of

* MIT license ([LICENSE-MIT](LICENSE) or http://opensource.org/licenses/MIT)
