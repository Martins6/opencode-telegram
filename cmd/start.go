package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/martins6/acolyte/internal/config"
	"github.com/martins6/acolyte/internal/service"
	"github.com/martins6/acolyte/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	startWorkspace string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Install, enable and start the Acolyte user service",
	Long: `Configure and start the Acolyte singleton user service.

Use ` + "`acolyte start --workspace PATH`" + ` to validate and persist a new
workspace before starting. The workspace must already be initialized with
` + "`acolyte new PATH`" + `. The flag is rejected while the service is running.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		bin, err := os.Executable()
		if err != nil {
			return fmt.Errorf("locate acolyte binary: %w", err)
		}

		mgr, err := serviceFactory()
		if err != nil {
			return err
		}

		path, err := config.SingletonConfigPath()
		if err != nil {
			return fmt.Errorf("locate singleton config: %w", err)
		}

		cfg, err := config.LoadIfExists(path)
		if err != nil {
			if errors.Is(err, config.ErrConfigNotFound) {
				return fmt.Errorf("singleton config not found at %s\n\nfix: run `acolyte new %s` and edit the config token", path, filepath.Dir(path))
			}
			return fmt.Errorf("load singleton config: %w", err)
		}

		desired := cfg.Workspace.Path
		if desired == "" {
			desired = filepath.Join(filepath.Dir(path))
		}

		if startWorkspace != "" {
			abs, err := filepath.Abs(startWorkspace)
			if err != nil {
				return fmt.Errorf("resolve --workspace: %w", err)
			}
			if err := workspace.StrictValidate(abs); err != nil {
				return err
			}
			status, err := mgr.Status(ctx)
			if err == nil && status.Loaded {
				return fmt.Errorf("cannot switch workspace while Acolyte is running\n\nfirst stop the service with `acolyte stop` and try again")
			}
			if err := config.WriteWorkspacePath(abs); err != nil {
				return fmt.Errorf("persist workspace: %w", err)
			}
			desired = abs
		} else if desired == "" {
			return fmt.Errorf("no workspace configured\n\nfix: run `acolyte new %s`", filepath.Dir(path))
		} else {
			if err := workspace.StrictValidate(desired); err != nil {
				return fmt.Errorf("configured workspace is invalid\n\n%w", err)
			}
		}

		cfgSvc := service.ServiceConfig{Workspace: desired, Binary: bin}

		if _, err := os.Stat(mgr.UnitPath()); err != nil {
			if err := mgr.Install(ctx, cfgSvc); err != nil {
				return fmt.Errorf("install service: %w", err)
			}
		} else {
			if err := mgr.Install(ctx, cfgSvc); err != nil {
				return fmt.Errorf("update service: %w", err)
			}
		}

		if err := mgr.Enable(ctx); err != nil {
			return fmt.Errorf("enable service: %w", err)
		}
		if err := mgr.Start(ctx); err != nil {
			return fmt.Errorf("start service: %w", err)
		}

		if _, err := mgr.WaitReady(ctx, 5*time.Second); err != nil {
			return fmt.Errorf("service did not become ready in time: %w", err)
		}

		fmt.Println("Acolyte started.")
		fmt.Println("status:  acolyte status")
		fmt.Println("logs:    acolyte logs")
		return nil
	},
}

func init() {
	startCmd.Flags().StringVar(&startWorkspace, "workspace", "", "set a new workspace path (only when stopped)")
	rootCmd.AddCommand(startCmd)
}
