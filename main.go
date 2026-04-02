// main.go
package main

import (
	"fmt"
	"os"

	"github.com/anyjava/kbn/cmd"
	"github.com/anyjava/kbn/config"
	"github.com/anyjava/kbn/model"
	"github.com/anyjava/kbn/parser"
	"github.com/anyjava/kbn/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func main() {
	var (
		showAll      bool
		configPath   string
		pathOverride string
	)

	rootCmd := &cobra.Command{
		Use:   "kbn",
		Short: "Obsidian vault kanban board TUI viewer",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}

			if pathOverride != "" {
				cfg.Path = pathOverride
			}

			cards, errs := parser.ParseCards(cfg.FullPath(), cfg.Glob, cfg.Fields)
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "warning: %v\n", e)
			}

			if !showAll {
				cards = model.FilterCards(cards, cfg.HiddenStatuses)
			}

			board := model.NewBoard(cards, cfg.ColumnOrder)
			app := tui.NewApp(board, cfg.ColumnOrder, cfg.PreviewLayout)

			p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
			if _, err := p.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.Flags().BoolVar(&showAll, "all", false, "Show all cards including hidden statuses")
	rootCmd.Flags().StringVar(&configPath, "config", "", "Path to config file")
	rootCmd.Flags().StringVar(&pathOverride, "path", "", "Override vault path")

	rootCmd.AddCommand(cmd.NewInitCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
