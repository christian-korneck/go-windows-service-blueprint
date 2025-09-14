//go:build windows

package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog *eventlog.Log

type myservice struct{}


func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	time.Sleep(5 * time.Second)

	// Start HTTP server

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		elog.Info(1, "svctest1 starting HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			elog.Error(1, fmt.Sprintf("HTTP server error: %v", err))
		}
	}()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	elog.Info(1, "svctest1 service started successfully")

	startTime := time.Now()
	keepaliveTicker := time.NewTicker(1 * time.Minute)
	shutdownTimer := time.NewTimer(3 * time.Minute)
	defer keepaliveTicker.Stop()
	defer shutdownTimer.Stop()

loop:
	//lint:ignore S1000 need select to handle multiple channels
	for {
		select {
		case <-keepaliveTicker.C:
			elapsed := int(time.Since(startTime).Seconds())
			elog.Info(1, fmt.Sprintf("keepalive status: active since %d seconds", elapsed))
		case <-shutdownTimer.C:
			elog.Info(1, "svctest1 service shutting down after 3 minutes")
			changes <- svc.Status{State: svc.StopPending}
			break loop
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				elog.Info(1, "svctest1 received Interrogate signal")
				changes <- c.CurrentStatus
			case svc.Stop:
				elog.Info(1, "svctest1 received Stop signal")
				changes <- svc.Status{State: svc.StopPending}
				elog.Info(1, "svctest1 performing shutdown cleanup")
				break loop
			case svc.Shutdown:
				elog.Info(1, "svctest1 received Shutdown signal")
				changes <- svc.Status{State: svc.StopPending}
				elog.Info(1, "svctest1 performing shutdown cleanup")
				break loop
			case svc.Pause:
				elog.Info(1, "svctest1 received Pause signal")
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				elog.Info(1, "svctest1 received Continue signal")
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			case svc.ParamChange:
				elog.Info(1, "svctest1 received ParamChange signal")
			case svc.NetBindAdd:
				elog.Info(1, "svctest1 received NetBindAdd signal")
			case svc.NetBindRemove:
				elog.Info(1, "svctest1 received NetBindRemove signal")
			case svc.NetBindEnable:
				elog.Info(1, "svctest1 received NetBindEnable signal")
			case svc.NetBindDisable:
				elog.Info(1, "svctest1 received NetBindDisable signal")
			case svc.DeviceEvent:
				elog.Info(1, "svctest1 received DeviceEvent signal")
			case svc.HardwareProfileChange:
				elog.Info(1, "svctest1 received HardwareProfileChange signal")
			case svc.PowerEvent:
				elog.Info(1, "svctest1 received PowerEvent signal")
			case svc.SessionChange:
				elog.Info(1, "svctest1 received SessionChange signal")
			case svc.PreShutdown:
				elog.Info(1, "svctest1 received PreShutdown signal")
				break loop
			default:
				elog.Error(1, fmt.Sprintf("svctest1 received unexpected signal %d", c.Cmd))
			}
		}
	}

	// Shutdown HTTP server gracefully
	elog.Info(1, "svctest1 shutting down HTTP server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		elog.Error(1, fmt.Sprintf("HTTP server shutdown error: %v", err))
	} else {
		elog.Info(1, "svctest1 HTTP server shutdown complete")
	}

	changes <- svc.Status{State: svc.StopPending}
	return
}

func setupEventLog(name string) error {
	err := eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func main() {
	const svcName = "svctest1"

	// Setup event log source first
	// (in the original golang.org/x/sys/windows/svc/example app, this part is in
	// the installer part (installService), but we need this here in the service)
	err := setupEventLog(svcName)
	if err != nil {
		// Try to continue anyway - maybe it already exists
	}

	elog, err = eventlog.Open(svcName)
	if err != nil {
		return
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", svcName))
	err = svc.Run(svcName, &myservice{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", svcName, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", svcName))
}
