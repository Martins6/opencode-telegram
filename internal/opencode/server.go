package opencode

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"
)

type Server struct {
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
	port   string
}

var server *Server

func NewServer(port, password, workspace string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	server = &Server{
		ctx:  ctx,
		port: port,
	}
	server.cancel = cancel
	return server
}

func (s *Server) Start(workspace string) error {
	args := []string{"serve", "--port", s.port, "--hostname", "127.0.0.1"}

	cmd := exec.CommandContext(s.ctx, "opencode", args...)
	cmd.Dir = workspace

	if s.port != "4096" {
		log.Printf("Starting OpenCode server on port %s", s.port)
	} else {
		log.Printf("Starting OpenCode server")
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start OpenCode server: %w", err)
	}

	s.cmd = cmd

	if err := s.waitForServer(); err != nil {
		return err
	}

	log.Println("OpenCode server started successfully")
	return nil
}

func (s *Server) waitForServer() error {
	url := fmt.Sprintf("http://127.0.0.1:%s/global/health", s.port)

	for i := 0; i < 30; i++ {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		default:
		}

		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for OpenCode server to start")
}

func (s *Server) Stop() error {
	log.Println("Stopping OpenCode server...")

	if s.cancel != nil {
		s.cancel()
	}

	if s.cmd != nil && s.cmd.Process != nil {
		if err := s.cmd.Process.Kill(); err != nil {
			log.Printf("Error killing OpenCode server: %v", err)
		}
		s.cmd.Wait()
	}

	log.Println("OpenCode server stopped")
	return nil
}

func GetServer() *Server {
	return server
}
