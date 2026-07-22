package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/martins6/acolyte/internal/config"
	"github.com/martins6/acolyte/internal/daemon"
	"github.com/martins6/acolyte/internal/workspace"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:    "__daemon",
	Short:  "internal worker invoked by the service manager",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.SingletonConfigPath()
		if err != nil {
			return fmt.Errorf("locate singleton config: %w", err)
		}
		cfg, err := config.LoadIfExists(path)
		if err != nil {
			return fmt.Errorf("load singleton config: %w", err)
		}
		workspacePath := cfg.Workspace.Path
		if workspacePath == "" {
			return fmt.Errorf("singleton config has no workspace.path")
		}
		if err := workspace.StrictValidate(workspacePath); err != nil {
			return err
		}

		rt, err := daemon.New(daemon.Options{
			Workspace:  workspacePath,
			ConfigPath: path,
		})
		if err != nil {
			return fmt.Errorf("init daemon: %w", err)
		}

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		return rt.Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
