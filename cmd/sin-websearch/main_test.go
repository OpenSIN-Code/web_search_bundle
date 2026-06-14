// SPDX-License-Identifier: MIT
// Purpose: Smoke tests for the CLI command builders.
// Docs: main_test.doc.md
package main

import (
	"testing"

	"github.com/spf13/cobra"
)

// commandBuilders returns every command factory in the package.
func commandBuilders() []*cobra.Command {
	return []*cobra.Command{
		newSearchCmd(),
		newPulseCmd(),
		newBriefCmd(),
		newResolveCmd(),
		newMonitorCmd(),
		newWatchCmd(),
		newVBriefCmd(),
		newVBriefPromptCmd(),
		newMissionCmd(),
		newVerifyCmd(),
		newProfileCmd(),
		newSecretsCmd(),
		newServeCmd(),
		newHTTPCmd(),
		newConfigCmd(),
		newAlchemistCmd(),
		newCompletionCmd(),
	}
}

func TestCommandBuildersReturnNonNil(t *testing.T) {
	for _, cmd := range commandBuilders() {
		if cmd == nil {
			t.Fatal("command builder returned nil")
		}
	}
}

func TestCommandBuildersHaveUseAndShort(t *testing.T) {
	for _, cmd := range commandBuilders() {
		if cmd.Use == "" {
			t.Errorf("%s: Use is empty", cmd.Name())
		}
		if cmd.Short == "" {
			t.Errorf("%s: Short is empty", cmd.Name())
		}
	}
}

func TestCompletionCmdValidArgs(t *testing.T) {
	cmd := newCompletionCmd()
	if len(cmd.ValidArgs) != 4 {
		t.Fatalf("completion valid args = %d, want 4", len(cmd.ValidArgs))
	}
	seen := map[string]bool{}
	for _, v := range cmd.ValidArgs {
		seen[v] = true
	}
	for _, want := range []string{"bash", "zsh", "fish", "powershell"} {
		if !seen[want] {
			t.Errorf("missing shell %q", want)
		}
	}
}

func TestCompletionCmdRunE(t *testing.T) {
	cmd := newCompletionCmd()
	if err := cmd.RunE(cmd, []string{"tcsh"}); err == nil {
		t.Error("expected unsupported shell to fail RunE")
	}
}

func TestRootCmd(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "sin-websearch",
		Short: "Unified Intelligence Gateway for OpenSIN",
	}
	for _, cmd := range commandBuilders() {
		rootCmd.AddCommand(cmd)
	}
	if len(rootCmd.Commands()) != 17 {
		t.Fatalf("root command has %d subcommands, want 17", len(rootCmd.Commands()))
	}
}
