// Purpose: Profile command to list and inspect research profiles.
// Docs: cmd/sin-websearch/profile_cmd.doc.md
package main

import (
	"fmt"

	"github.com/OpenSIN-Code/web_search_bundle/internal/profiles"
	"github.com/spf13/cobra"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage research profiles",
	}
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry, err := profiles.NewRegistry("")
			if err != nil {
				return err
			}
			for _, name := range registry.List() {
				fmt.Println(name)
			}
			return nil
		},
	}
	showCmd := &cobra.Command{
		Use:   "show [name]",
		Short: "Show a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry, err := profiles.NewRegistry("")
			if err != nil {
				return err
			}
			p, err := registry.Get(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Name: %s\nDescription: %s\nVersion: %s\nAgents: %d explore / %d librarian\nSources: %v\n",
				p.Name, p.Description, p.Version, p.Agents.Explore.Count, p.Agents.Librarian.Count, p.Sources.Required)
			return nil
		},
	}
	cmd.AddCommand(listCmd)
	cmd.AddCommand(showCmd)
	return cmd
}
