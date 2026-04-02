// model/card.go
package model

import "sort"

type Card struct {
	ID       string
	Title    string
	Status   string
	Priority string
	Type     string
	FilePath string
	Meta     map[string]string
}

type Column struct {
	Name  string
	Cards []Card
}

type Board struct {
	Columns []Column
}

func FilterCards(cards []Card, hiddenStatuses []string) []Card {
	if len(hiddenStatuses) == 0 {
		return cards
	}
	hidden := make(map[string]bool, len(hiddenStatuses))
	for _, s := range hiddenStatuses {
		hidden[s] = true
	}
	var result []Card
	for _, c := range cards {
		if !hidden[c.Status] {
			result = append(result, c)
		}
	}
	return result
}

func NewBoard(cards []Card, columnOrder []string) Board {
	groups := make(map[string][]Card)
	for _, c := range cards {
		groups[c.Status] = append(groups[c.Status], c)
	}

	columns := make([]Column, 0, len(groups))
	for name, cards := range groups {
		columns = append(columns, Column{Name: name, Cards: cards})
	}

	if len(columnOrder) > 0 {
		orderIndex := make(map[string]int, len(columnOrder))
		for i, name := range columnOrder {
			orderIndex[name] = i
		}
		sort.Slice(columns, func(i, j int) bool {
			oi, okI := orderIndex[columns[i].Name]
			oj, okJ := orderIndex[columns[j].Name]
			if okI && okJ {
				return oi < oj
			}
			if okI {
				return true
			}
			if okJ {
				return false
			}
			return columns[i].Name < columns[j].Name
		})
	} else {
		sort.Slice(columns, func(i, j int) bool {
			if len(columns[i].Cards) != len(columns[j].Cards) {
				return len(columns[i].Cards) > len(columns[j].Cards)
			}
			return columns[i].Name < columns[j].Name
		})
	}

	return Board{Columns: columns}
}
