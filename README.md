# simple win32 service executable in Go

## What

A simple *Windows service* executable in Go (blueprint for copy + paste).    
    
Uses the pkg [`golang.org/x/sys/windows/svc`](https://pkg.go.dev/golang.org/x/sys/windows/svc), which is CGO-free. It's very similar to its [example](https://pkg.go.dev/golang.org/x/sys/windows/svc/example), but more generic, without the management functionality (service executable only) and all signals ready to be wired).

## service behaviour

- logs all signals received to the windows `Application` event log (i.e. start, stop, shutdown, pause, resume and more)
- sleeps 5 seconds during startup before it signals running state (init demo)
- sleeps 5 seconds when it receives stop/shutdown signal (graceful shutdown demo)
- runs a http server on `http://localhost:8080/` with cancellation wired to service start/stop.
  - warning: other signals like pause, resume, etc are not wired to the http server (only log)
- every minute the total elapsed time is logged to the Windows event log
- after 3 minutes runtime the service shuts down gracefully

This should give basic building blocks for simple use cases.

## Build and install

```
GOOS=windows GOARCH=amd64 go build .
```

install and start with admin privs:

```
.\install.cmd
.\start.cmd
```

query status:
```
.\status.cmd
```

(or use the `services.msc` GUI)

