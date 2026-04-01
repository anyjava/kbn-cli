// parser/parser.go
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/anyjava/kbn/config"
	"github.com/anyjava/kbn/model"
)

func ParseCards(dir string, glob string, fields config.Fields) ([]model.Card, []error) {
	pattern := filepath.Join(dir, glob)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, []error{fmt.Errorf("glob %q: %w", pattern, err)}
	}

	var cards []model.Card
	var errs []error

	for _, path := range matches {
		card, err := parseFile(path, fields)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", filepath.Base(path), err))
			continue
		}
		if card.Status == "" {
			continue // skip files without status field
		}
		cards = append(cards, card)
	}

	return cards, errs
}

func parseFile(path string, fields config.Fields) (model.Card, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.Card{}, err
	}
	defer f.Close()

	var meta map[string]interface{}
	_, err = frontmatter.Parse(f, &meta)
	if err != nil {
		return model.Card{}, err
	}

	card := model.Card{
		FilePath: path,
		Meta:     make(map[string]string),
	}

	mapped := map[string]*string{
		fields.ID:       &card.ID,
		fields.Title:    &card.Title,
		fields.Status:   &card.Status,
		fields.Priority: &card.Priority,
		fields.Type:     &card.Type,
	}

	for key, val := range meta {
		str := fmt.Sprintf("%v", val)
		if target, ok := mapped[key]; ok && key != "" {
			*target = str
		} else {
			card.Meta[key] = str
		}
	}

	// Fallback: use filename as ID/Title if not mapped
	if card.ID == "" {
		base := strings.TrimSuffix(filepath.Base(path), ".md")
		card.ID = base
	}
	if card.Title == "" {
		base := strings.TrimSuffix(filepath.Base(path), ".md")
		// Remove ID prefix if present (e.g., "UL-001 Task Name" -> "Task Name")
		if card.ID != "" && strings.HasPrefix(base, card.ID) {
			card.Title = strings.TrimSpace(strings.TrimPrefix(base, card.ID))
		}
		if card.Title == "" {
			card.Title = base
		}
	}

	return card, nil
}
