# About

Example [lxd client library](https://pkg.go.dev/github.com/lxc/lxd/client) to execute a command inside an lxd container while capturing its output and forwarding the Ctrl+C signal.

# Usage

```bash
export LXD_SOCKET='/var/snap/lxd/common/lxd/unix.socket'

# launch the example container.
lxc delete --force lxd-exec-example
lxc launch images:debian/11 lxd-exec-example

# build this example binary, upload it to the container, and execute it.
go build
lxc file push lxd-exec-example lxd-exec-example/lxd-exec-example
./lxd-exec-example
```
