package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/anyjava/kbn/config"
	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a .kbn.yml config file interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(".kbn.yml"); err == nil {
				return fmt.Errorf(".kbn.yml already exists in current directory")
			}

			scanner := bufio.NewScanner(os.Stdin)

			vault := prompt(scanner, "Obsidian vault path", "")
			path := prompt(scanner, "Folder path within vault (relative)", "")
			statusField := prompt(scanner, "Frontmatter field for status", "status")
			idField := prompt(scanner, "Frontmatter field for ID (leave empty to use filename)", "")
			titleField := prompt(scanner, "Frontmatter field for title (leave empty to use filename)", "")
			priorityField := prompt(scanner, "Frontmatter field for priority (optional)", "")
			typeField := prompt(scanner, "Frontmatter field for type (optional)", "")

			hiddenInput := prompt(scanner, "Statuses to hide (comma-separated, e.g. Closed,Archived)", "")
			var hidden []string
			if hiddenInput != "" {
				for _, s := range strings.Split(hiddenInput, ",") {
					s = strings.TrimSpace(s)
					if s != "" {
						hidden = append(hidden, s)
					}
				}
			}

			orderInput := prompt(scanner, "Column order (comma-separated, e.g. TODO,In Progress,Done)", "")
			var order []string
			if orderInput != "" {
				for _, s := range strings.Split(orderInput, ",") {
					s = strings.TrimSpace(s)
					if s != "" {
						order = append(order, s)
					}
				}
			}

			cfg := config.Config{
				Vault: vault,
				Path:  path,
				Glob:  "*.md",
				Fields: config.Fields{
					ID:       idField,
					Title:    titleField,
					Status:   statusField,
					Priority: priorityField,
					Type:     typeField,
				},
				HiddenStatuses: hidden,
				ColumnOrder:    order,
			}

			data, err := yaml.Marshal(&cfg)
			if err != nil {
				return fmt.Errorf("marshaling config: %w", err)
			}

			if err := os.WriteFile(".kbn.yml", data, 0644); err != nil {
				return fmt.Errorf("writing .kbn.yml: %w", err)
			}

			fmt.Println("\n✓ Created .kbn.yml")
			fmt.Println("  Run 'kbn' to view your kanban board.")
			return nil
		},
	}
}

func prompt(scanner *bufio.Scanner, label string, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}

	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return defaultVal
	}
	return input
}
