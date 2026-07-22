package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/martins6/acolyte/internal/config"
	"github.com/martins6/acolyte/internal/service"
	"github.com/spf13/cobra"
)

var stopForever bool

var serviceFactory = func() (service.Manager, error) { return service.New(nil) }

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Acolyte user service",
	Long: `Stop the running Acolyte service. Use --forever to also disable
the service so it does not start at the next login.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		mgr, err := serviceFactory()
		if err != nil {
			return err
		}

		if err := mgr.Stop(ctx); err != nil {
			return err
		}

		if stopForever {
			if err := mgr.Disable(ctx); err != nil {
				return fmt.Errorf("disable service: %w", err)
			}
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Acolyte stopped.")
		return nil
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Acolyte user service",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		mgr, err := serviceFactory()
		if err != nil {
			return err
		}
		if err := mgr.Restart(ctx); err != nil {
			return err
		}
		if _, err := mgr.WaitReady(ctx, 5*time.Second); err != nil {
			return fmt.Errorf("service did not become ready in time: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Acolyte restarted.")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print Acolyte service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		mgr, err := serviceFactory()
		if err != nil {
			return err
		}
		st, err := mgr.Status(ctx)
		if err != nil {
			return err
		}

		state := "stopped"
		if st.Loaded {
			state = "running"
		}
		autostart := "disabled"
		if st.Autostart {
			autostart = "enabled"
		}

		workspace := st.Workspace
		if workspace == "" {
			if path, err := config.SingletonConfigPath(); err == nil {
				if cfg, err := config.LoadIfExists(path); err == nil && cfg.Workspace.Path != "" {
					abs, _ := filepath.Abs(cfg.Workspace.Path)
					workspace = abs
				}
			}
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Status:    %s\n", state)
		if workspace != "" {
			fmt.Fprintf(out, "Workspace: %s\n", workspace)
		}
		fmt.Fprintf(out, "Autostart: %s\n", autostart)
		if st.Loaded {
			fmt.Fprintf(out, "PID:       %d\n", st.PID)
		}
		if st.Reason != "" {
			fmt.Fprintf(out, "Reason:    %s\n", st.Reason)
		}

		if !st.Loaded {
			return ExitCode(ExitStopped)
		}
		return nil
	},
}

func init() {
	stopCmd.Flags().BoolVar(&stopForever, "forever", false, "also disable the service so it does not start at next login")
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(statusCmd)
}
