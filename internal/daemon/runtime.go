package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/martins6/acolyte/internal/bot"
	"github.com/martins6/acolyte/internal/config"
	"github.com/martins6/acolyte/internal/database"
	"github.com/martins6/acolyte/internal/logger"
	acolyteruntime "github.com/martins6/acolyte/internal/runtime"
	"github.com/martins6/acolyte/internal/scheduler"
	"github.com/martins6/acolyte/internal/updater"
	"github.com/martins6/acolyte/internal/workspace"
)

type Options struct {
	Workspace  string
	ConfigPath string
	BinaryPath string
}

type Exec interface {
	LookPath(file string) (string, error)
}

type stdExec struct{}

func (stdExec) LookPath(file string) (string, error) { return exec.LookPath(file) }

func New(opts Options) (*Runtime, error) {
	if opts.Workspace == "" {
		return nil, fmt.Errorf("daemon: workspace path is required")
	}
	abs, err := filepath.Abs(opts.Workspace)
	if err != nil {
		return nil, fmt.Errorf("daemon: resolve workspace: %w", err)
	}
	abs = filepath.Clean(abs)
	opts.Workspace = abs

	stateDir, err := acolyteruntime.Dir()
	if err != nil {
		return nil, fmt.Errorf("daemon: resolve runtime dir: %w", err)
	}
	bin := opts.BinaryPath
	if bin == "" {
		exe, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("daemon: locate executable: %w", err)
		}
		bin = exe
	}
	return &Runtime{
		opts:       opts,
		stateDir:   stateDir,
		exec:       stdExec{},
		binaryGet:  os.Executable,
		currentPID: os.Getpid,
		startedAt:  time.Now().UTC(),
		binary:     bin,
	}, nil
}

type Runtime struct {
	opts       Options
	stateDir   string
	exec       Exec
	binaryGet  func() (string, error)
	currentPID func() int
	startedAt  time.Time
	binary     string
	cancel     context.CancelFunc
}

func (r *Runtime) Run(ctx context.Context) error {
	if err := workspace.StrictValidate(r.opts.Workspace); err != nil {
		return err
	}

	if _, err := r.exec.LookPath("opencode"); err != nil {
		log.Printf("daemon: warning: opencode CLI not found in PATH: %v", err)
	}

	if err := logger.Initialize(r.opts.Workspace); err != nil {
		log.Printf("daemon: logger init failed: %v", err)
	}
	logger.LogDebug("daemon: starting worker in workspace %s", r.opts.Workspace)

	if err := database.Init(r.opts.Workspace); err != nil {
		return fmt.Errorf("daemon: init database: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		var err error
		cfg, err = config.Load("")
		if err != nil {
			return fmt.Errorf("daemon: load config: %w", err)
		}
	}

	bot.SetConfig(cfg)

	if err := r.writeState(false); err != nil {
		log.Printf("daemon: failed to write initial state: %v", err)
	}

	runCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	defer cancel()

	telegramBot, err := bot.Initialize(cfg)
	if err != nil {
		return fmt.Errorf("daemon: init bot: %w", err)
	}
	if telegramBot != nil {
		bot.RegisterHandlers(telegramBot)
	}

	if err := bot.Start(runCtx, telegramBot); err != nil {
		return fmt.Errorf("daemon: start bot: %w", err)
	}

	if err := bot.StartNotifier(runCtx, telegramBot); err != nil {
		log.Printf("daemon: notifier failed to start: %v", err)
	}

	if err := scheduler.StartScheduler(runCtx, telegramBot, r.opts.Workspace); err != nil {
		log.Printf("daemon: scheduler failed to start: %v", err)
	}

	go r.updateChecker(runCtx)

	if err := acolyteruntime.TouchReady(r.stateDir); err != nil {
		log.Printf("daemon: touch ready state: %v", err)
	}
	logger.LogDebug("daemon: ready")

	<-runCtx.Done()

	logger.LogDebug("daemon: stopping")
	if err := acolyteruntime.MarkStopping(r.stateDir); err != nil {
		log.Printf("daemon: mark stopping: %v", err)
	}
	bot.StopNotifier()
	scheduler.Stop()

	logger.Close()
	if err := acolyteruntime.Remove(r.stateDir); err != nil {
		log.Printf("daemon: remove state: %v", err)
	}
	return nil
}

func (r *Runtime) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *Runtime) writeState(ready bool) error {
	state := acolyteruntime.State{
		PID:        r.currentPID(),
		Workspace:  r.opts.Workspace,
		StartedAt:  r.startedAt,
		Ready:      ready,
		Stopping:   false,
		BinaryPath: r.binary,
	}
	return acolyteruntime.Write(r.stateDir, state)
}

func (r *Runtime) updateChecker(parent context.Context) {
	checkCtx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	latest, outdated, err := updater.IsOutdated(checkCtx, updater.Options{})
	if err != nil || !outdated {
		return
	}
	logger.LogDebug("daemon: update available: %s", latest)
}
