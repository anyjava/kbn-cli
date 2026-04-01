# kbn - Obsidian Vault Kanban Board TUI

## Overview

Obsidian vault의 마크다운 파일을 읽어 터미널에서 인터랙티브 칸반 보드로 보여주는 CLI 도구. 각 마크다운 파일의 YAML frontmatter에 있는 status 필드를 기준으로 칼럼을 자동 생성하고, bubbletea 기반 TUI로 탐색/미리보기/편집을 지원한다.

## Goals

- Obsidian vault 내 마크다운 파일을 칸반 보드 형태로 터미널에서 조회
- 범용: 어떤 vault/폴더 구조든 설정 파일로 대응
- 인터랙티브 TUI: 키보드로 칼럼/카드 탐색, 마크다운 미리보기, 에디터 연동

## Non-Goals

- 티켓 생성/수정/삭제 (v1 범위 외)
- 멀티 프로젝트 전환
- 외부 서비스 연동 (Jira, Trello 등)
- 모바일/웹 지원

## Tech Stack

| 용도 | 라이브러리 |
|------|-----------|
| 언어 | Go 1.22+ |
| CLI 파싱 | spf13/cobra |
| TUI 프레임워크 | charmbracelet/bubbletea |
| TUI 스타일링 | charmbracelet/lipgloss |
| 마크다운 렌더링 | charmbracelet/glamour |
| YAML 파싱 | gopkg.in/yaml.v3 |
| Frontmatter 파싱 | adrg/frontmatter |

## Project Structure

```
kbn/
├── main.go              # 엔트리포인트, CLI 파싱
├── config/
│   └── config.go        # .kbn.yml 로드/파싱
├── parser/
│   └── parser.go        # md 파일 스캔, frontmatter 파싱
├── model/
│   └── card.go          # Card, Column, Board 도메인 모델
├── tui/
│   ├── app.go           # bubbletea 메인 Model
│   ├── board.go         # 칸반 보드 뷰 (칼럼 수평 배치)
│   ├── preview.go       # 마크다운 미리보기 패널
│   └── styles.go        # lipgloss 스타일 정의
└── .kbn.yml             # 예시 설정 파일
```

## Configuration

파일 위치 우선순위: 현재 디렉토리 `.kbn.yml` > `~/.config/kbn/config.yml`

```yaml
# .kbn.yml
vault: "/path/to/obsidian/vault"
path: "개발일/Underlog"          # vault 내 상대 경로
glob: "*.md"                     # 스캔할 파일 패턴

# frontmatter 필드 매핑
fields:
  id: "ticket_id"
  title: "title"
  status: "status"
  priority: "priority"
  type: "type"

# 기본 숨김 상태 (--all로 해제)
hidden_statuses:
  - "Closed"
```

### Fields Mapping

- `id`: 카드에 표시할 식별자 필드. 없으면 파일명 사용.
- `title`: 카드 제목. 없으면 파일명에서 ID를 제거한 나머지 사용.
- `status`: 칼럼 그룹핑 기준. 필수.
- `priority`, `type`: 카드에 보조 정보로 표시. 선택.

## Data Model

```go
type Card struct {
    ID       string            // fields.id 매핑값
    Title    string            // fields.title 매핑값
    Status   string            // fields.status 매핑값
    Priority string            // fields.priority 매핑값
    Type     string            // fields.type 매핑값
    FilePath string            // md 파일 절대 경로
    Meta     map[string]string // 매핑되지 않은 나머지 frontmatter
}

type Board struct {
    Columns []Column
}

type Column struct {
    Name  string
    Cards []Card
}
```

## Parsing Flow

1. `vault/path` 디렉토리에서 `glob` 패턴에 맞는 `.md` 파일 수집
2. 각 파일의 YAML frontmatter 파싱 -> `fields` 매핑에 따라 Card 생성
3. `hidden_statuses`에 해당하는 Card 필터링 (`--all`이면 스킵)
4. `status` 값 기준으로 Card를 그룹핑 -> Column 생성
5. Column 정렬: 카드 수 내림차순

## TUI Layout

```
┌─ kbn ─────────────────────────────────────────────────────────┐
│ Backlog (3)     │ In Progress (2) │ Done (5)    │ Preview     │
│─────────────────│─────────────────│─────────────│─────────────│
│ > UL-017 도서…  │   UL-016 Notion │   UL-001 …  │ ## 개요     │
│   UL-015 CRM…  │   UL-014 Notion │   UL-002 …  │             │
│   UL-009 멀티…  │                 │   UL-003 …  │ 사용자가…   │
│                 │                 │   UL-004 …  │             │
│                 │                 │   UL-005 …  │ ## 작업 내용 │
│                 │                 │             │ - [x] ...   │
└─────────────────────────────────────────────────────────────────┘
  ←→ 칼럼 이동  ↑↓ 카드 이동  Enter 에디터 열기  q 종료  ? 도움말
```

### Key Bindings

| 키 | 동작 |
|---|------|
| `←` `→` / `h` `l` | 칼럼 간 이동 |
| `↑` `↓` / `j` `k` | 칼럼 내 카드 이동 |
| `Enter` | `$EDITOR`로 파일 열기 |
| `p` | 미리보기 패널 토글 |
| `/` | 카드 검색 (제목/ID 필터링) |
| `q` / `Ctrl+C` | 종료 |
| `?` | 키 바인딩 도움말 |

### Preview Panel

- 화면 우측 약 35% 차지
- charmbracelet/glamour로 마크다운 렌더링
- 커서 이동 시 해당 카드의 md 파일 본문 표시
- `p` 키로 토글 (숨기면 보드 영역 확장)

## CLI Interface

```bash
kbn                    # TUI 보드 실행
kbn --all              # Closed 포함 전체 표시
kbn --config path.yml  # 설정 파일 경로 지정
kbn --path "다른/경로" # vault 내 경로 오버라이드
kbn init               # 대화형으로 .kbn.yml 생성
```

## Risks

| 리스크 | 대응 |
|--------|------|
| 터미널 폭이 좁아 칼럼이 잘림 | 최소 폭 미만 시 칼럼 수 줄이거나 스크롤 지원 |
| 마크다운 파일 수가 많아 파싱 느림 | 파일 수 표시 + 필요시 glob 패턴으로 범위 축소 |
| frontmatter 형식이 비표준 | 파싱 에러 시 해당 파일 스킵 + 경고 표시 |
| iCloud 경로 동기화 지연 | 로컬 파일만 읽으므로 동기화 완료된 상태에서 사용 |

## Success Criteria

- `kbn` 실행 시 칸반 보드가 TUI로 표시된다
- 키보드로 칼럼/카드 간 자유롭게 이동할 수 있다
- 카드 선택 시 마크다운 본문이 미리보기 패널에 렌더링된다
- Enter 키로 `$EDITOR`에서 파일을 열 수 있다
- `.kbn.yml` 설정으로 다른 vault/폴더에도 적용할 수 있다
- Closed 상태가 기본적으로 숨겨지고, `--all`로 표시할 수 있다
