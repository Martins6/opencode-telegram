package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/martins6/acolyte/internal/updater"
	"github.com/spf13/cobra"
)

var (
	updateCheck      bool
	updatePinVersion string
	updateRestart    bool
	updateYes        bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates and replace the running binary",
	Long: `Check for newer releases on GitHub, verify the downloaded archive
against the published checksums.txt, and replace the running binary in place.

The 'acolyte start' command will warn you when an update is available without
applying it. Run 'acolyte update' explicitly to perform the upgrade.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		opts := updater.Options{TargetVersion: updatePinVersion}

		rel, err := updater.Latest(ctx, opts)
		if err != nil {
			return fmt.Errorf("fetch latest release: %w", err)
		}

		current := updater.Current()
		newer, err := updater.IsNewer(rel.Tag, current)
		if err != nil {
			return err
		}

		fmt.Printf("Current version: %s\n", current)
		fmt.Printf("Latest version:  %s\n", rel.Tag)

		if updateCheck {
			if newer {
				fmt.Println("Status: update available.")
			} else {
				fmt.Println("Status: up to date.")
			}
			return nil
		}

		if !newer {
			fmt.Println("Already on the latest version. Nothing to do.")
			return nil
		}

		fmt.Printf("Download URL:    %s\n", rel.BrowserDownloadURL)

		if !updateYes && !confirmPrompt(os.Stdin, os.Stdout, fmt.Sprintf("Install %s now?", rel.Tag)) {
			fmt.Println("Aborted.")
			return nil
		}

		restartArgs := filterUpdateArg(os.Args[1:])
		applyOpts := updater.Options{
			TargetVersion:   updatePinVersion,
			Restart:         updateRestart,
			RestartArgs:     restartArgs,
			SignMac:         true,
			APITimeout:      30 * time.Second,
			DownloadTimeout: 2 * time.Minute,
		}

		if err := updater.Apply(ctx, rel, applyOpts); err != nil {
			if isPermissionError(err) {
				return fmt.Errorf("%w\n\nhint: reinstall via install.sh into a writable prefix (e.g. ~/.local/bin) or run with sudo", err)
			}
			return err
		}

		fmt.Printf("Updated to %s.\n", rel.Tag)
		if !updateRestart {
			exe, _ := os.Executable()
			fmt.Printf("Restart the bot to use the new version. (binary at %s)\n", exe)
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Only check and report; do not download or replace")
	updateCmd.Flags().StringVar(&updatePinVersion, "version", "", "Pin to a specific release tag (e.g. v1.2.3)")
	updateCmd.Flags().BoolVar(&updateRestart, "restart", true, "After replacement, re-exec the new binary")
	updateCmd.Flags().BoolVar(&updateYes, "yes", false, "Skip the interactive confirmation prompt")
	rootCmd.AddCommand(updateCmd)
}

func filterUpdateArg(argv []string) []string {
	out := make([]string, 0, len(argv))
	skipNext := false
	for _, a := range argv {
		if skipNext {
			skipNext = false
			continue
		}
		if a == "update" {
			continue
		}
		switch a {
		case "--version", "-v":
			skipNext = true
			continue
		case "--check":
			continue
		case "--restart=false":
			continue
		case "--yes", "-y":
			continue
		}
		out = append(out, a)
	}
	return out
}

func confirmPrompt(in io.Reader, out io.Writer, prompt string) bool {
	fmt.Fprintf(out, "%s [y/N]: ", prompt)
	rdr := bufio.NewReader(in)
	line, err := rdr.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false
	}
	ans := strings.ToLower(strings.TrimSpace(line))
	return ans == "y" || ans == "yes"
}

func isPermissionError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "permission denied") {
		return true
	}
	if strings.Contains(msg, "read-only file system") {
		return true
	}
	if errors.Is(err, os.ErrPermission) {
		return true
	}
	return false
}
