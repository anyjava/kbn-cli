package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

			fmt.Println("🗂  kbn init — 칸반 보드 설정을 생성합니다.")
			fmt.Println()

			// Vault path
			fmt.Println("  Obsidian vault의 루트 경로를 입력하세요.")
			fmt.Println("  예: /Users/you/Documents/MyVault")
			home, _ := os.UserHomeDir()
			defaultVault := filepath.Join(home, "Documents")
			vault := prompt(scanner, "Vault path", defaultVault)

			// Folder path
			fmt.Println()
			fmt.Println("  Vault 안에서 칸반 카드(md 파일)가 있는 폴더 경로.")
			fmt.Println("  vault 루트에 바로 있으면 비워두세요.")
			fmt.Println("  예: 프로젝트/MyApp")
			path := prompt(scanner, "Folder path", "")

			// Status field
			fmt.Println()
			fmt.Println("  마크다운 frontmatter에서 상태를 나타내는 필드명.")
			fmt.Println("  예: status: In Progress → 'status'")
			statusField := prompt(scanner, "Status field", "status")

			// ID field
			fmt.Println()
			fmt.Println("  카드 ID로 사용할 frontmatter 필드.")
			fmt.Println("  비워두면 파일명을 ID로 사용합니다.")
			fmt.Println("  예: ticket_id: UL-001 → 'ticket_id'")
			idField := prompt(scanner, "ID field", "")

			// Title field
			fmt.Println()
			fmt.Println("  카드 제목으로 사용할 frontmatter 필드.")
			fmt.Println("  비워두면 파일명을 제목으로 사용합니다.")
			fmt.Println("  예: title: Widget 개발 → 'title'")
			titleField := prompt(scanner, "Title field", "")

			// Priority field
			fmt.Println()
			fmt.Println("  우선순위 필드. 카드에 뱃지로 표시됩니다.")
			fmt.Println("  없으면 비워두세요.")
			fmt.Println("  예: priority: High → 'priority'")
			priorityField := prompt(scanner, "Priority field", "priority")

			// Type field
			fmt.Println()
			fmt.Println("  타입 필드. 카드에 뱃지로 표시됩니다.")
			fmt.Println("  없으면 비워두세요.")
			fmt.Println("  예: type: Feature → 'type'")
			typeField := prompt(scanner, "Type field", "type")

			// Hidden statuses
			fmt.Println()
			fmt.Println("  기본으로 숨길 상태. kbn --all 로 다시 볼 수 있습니다.")
			fmt.Println("  여러 개는 쉼표로 구분. 없으면 비워두세요.")
			fmt.Println("  예: Closed,Archived")
			hiddenInput := prompt(scanner, "Hidden statuses", "Closed")
			var hidden []string
			if hiddenInput != "" {
				for _, s := range strings.Split(hiddenInput, ",") {
					s = strings.TrimSpace(s)
					if s != "" {
						hidden = append(hidden, s)
					}
				}
			}

			// Column order
			fmt.Println()
			fmt.Println("  칼럼 표시 순서. 지정하지 않으면 카드 수 순으로 정렬됩니다.")
			fmt.Println("  여러 개는 쉼표로 구분. 없으면 비워두세요.")
			fmt.Println("  예: TODO,In Progress,Done")
			orderInput := prompt(scanner, "Column order", "")
			var order []string
			if orderInput != "" {
				for _, s := range strings.Split(orderInput, ",") {
					s = strings.TrimSpace(s)
					if s != "" {
						order = append(order, s)
					}
				}
			}

			// Preview layout
			fmt.Println()
			fmt.Println("  미리보기 패널 위치.")
			fmt.Println("  right: 오른쪽 칼럼 / bottom: 아래쪽 분할")
			previewLayout := prompt(scanner, "Preview layout", "bottom")
			if previewLayout != "right" && previewLayout != "bottom" {
				previewLayout = "bottom"
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
				PreviewLayout:  previewLayout,
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
		fmt.Printf("  → %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("  → %s: ", label)
	}

	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return defaultVal
	}
	return input
}
