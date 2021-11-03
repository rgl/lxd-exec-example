// Copyright 2021 Rui Lopes (ruilopes.com). All rights reserved.

package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	lxd "github.com/lxc/lxd/client"
	lxdApi "github.com/lxc/lxd/shared/api"
)

type logWriter struct {
	p string
	b bytes.Buffer
}

func newLogWriter(prefix string) *logWriter {
	return &logWriter{p: prefix}
}

func (w *logWriter) Write(p []byte) (int, error) {
	n, err := w.b.Write(p)
	if err != nil {
		return 0, err
	}
	for {
		line, err := w.b.ReadString('\n')
		if err == io.EOF {
			break
		}
		line = line[0 : len(line)-1]
		log.Printf("%s: %s", w.p, line)
	}
	return n, nil
}

func (lb *logWriter) Close() error {
	if lb.b.Len() > 0 {
		line := lb.b.String()
		log.Printf("%s: %s", lb.p, line)
	}
	lb.b.Reset()
	return nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "console" {
		console()
		return
	}

	instanceName := "lxd-exec-example"

	ctx, ctxCancel := context.WithCancel(context.Background())

	c, err := lxd.ConnectLXDUnix("", nil)
	if err != nil {
		log.Fatalf("failed to create the lxd client: %v", err)
	}

	// cancel ctx when a SIGINT or SIGTERM is received by this process.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)
	go func() {
		s := <-ch
		log.Printf("Received signal %d. Going to cancel ctx.", s)
		ctxCancel()
	}()

	execRequest := lxdApi.InstanceExecPost{
		Command:   []string{"/" + instanceName, "console"},
		WaitForWS: true,
	}
	execStdout := newLogWriter("console app stdout")
	execStderr := newLogWriter("console app stderr")
	execArgs := lxd.InstanceExecArgs{
		Stdin:  nil,
		Stdout: execStdout,
		Stderr: execStderr,
		Control: func(control *websocket.Conn) {
			closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
			defer control.WriteMessage(websocket.CloseMessage, closeMsg)

			// wait for the context to be canceled.
			<-ctx.Done()

			log.Println("Sending SIGINT to console app")
			err := control.WriteJSON(lxdApi.InstanceExecControl{
				Command: "signal",
				Signal:  int(syscall.SIGINT),
			})
			if err != nil {
				log.Println("ERROR Failed to send SIGINT to console app: %w", err)
			}

			log.Println("Exiting the Control loop")
		},
	}
	op, err := c.ExecInstance(instanceName, execRequest, &execArgs)
	if err != nil {
		log.Fatalf("failed to start the exec: %v", err)
	}
	err = op.Wait()
	if err != nil {
		log.Fatalf("failed to wait for the exec: %v", err)
	}

	exitCode := int(op.Get().Metadata["return"].(float64))
	log.Printf("Console app terminated with exitCode %d", exitCode)
}
