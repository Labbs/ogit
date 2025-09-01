package main

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

// TestSSH_UploadPackAdvertisedRefs spins up the SSH server and checks that
// git-upload-pack advertises refs over the SSH exec channel.
func TestSSH_UploadPackAdvertisedRefs(t *testing.T) {
	t.Parallel()

	// Isolate repositories dir
	tmpBase := t.TempDir()
	oldBase := BaseDir
	BaseDir = tmpBase
	t.Cleanup(func() { BaseDir = oldBase })

	// Create an empty bare repo (required: no auto-creation)
	if err := os.MkdirAll(BaseDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := initBare(filepath.Join(BaseDir, "demo.git")); err != nil {
		t.Fatalf("init bare: %v", err)
	}

	// Start SSH server on 127.0.0.1:0
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	srv, err := NewGitSSHServer(addr)
	if err != nil {
		t.Fatalf("ssh server: %v", err)
	}
	go func() { _ = srv.Serve(ln) }()
	time.Sleep(100 * time.Millisecond)

	// SSH client config: password auth accepted by server, ignore host key
	cfg := &gossh.ClientConfig{User: "git", HostKeyCallback: gossh.InsecureIgnoreHostKey(), Auth: []gossh.AuthMethod{gossh.Password("x")}}
	client, err := gossh.Dial("tcp", addr, cfg)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	defer sess.Close()

	// Start command and read some output
	stdout, err := sess.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	sess.Stderr = io.Discard
	if err := sess.Start("git-upload-pack '/demo.git'"); err != nil {
		t.Fatalf("start: %v", err)
	}

	done := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 256)
		n, _ := stdout.Read(buf)
		done <- buf[:n]
	}()
	select {
	case data := <-done:
		if len(data) == 0 {
			t.Fatalf("no advertised refs received over SSH")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for advertised refs over SSH")
	}
	// Close session; ignore exit status errors
	_ = sess.Close()
}
