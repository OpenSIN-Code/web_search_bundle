// Purpose: CLI command to initialize a program.md template.
// Docs: alchemist_init_cmd.doc.md
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newAlchemistInitCmd() *cobra.Command {
	var template string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize program.md template",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := os.Getwd()
			path := filepath.Join(repoPath, "program.md")

			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("program.md already exists at %s", path)
			}

			content := getProgramTemplate(template)
			// program.md is a user-visible project file; 0644 is intentional.
			if err := os.WriteFile(path, []byte(content), 0644); err != nil { // #nosec G306
				return err
			}

			fmt.Printf("✓ Created program.md at %s\n", path)
			fmt.Println("\nEdit the hypotheses queue, then run:")
			fmt.Println("  sin-websearch alchemist run --cmd \"your verification command\"")
			return nil
		},
	}
	cmd.Flags().StringVar(&template, "template", "go", "Template: go|python|ml")
	return cmd
}
