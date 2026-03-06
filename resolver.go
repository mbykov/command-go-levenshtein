package command

import (
	"encoding/json"
	"io"
	"os"
	"log"
	"sort"
	"strings"

	"github.com/agnivade/levenshtein"
)

// CommandDefinition определяет одну команду
type CommandDefinition struct {
	Name     string   `json:"name"`
	Synonyms []string `json:"synonyms"`
	External bool     `json:"external,omitempty"` // true для команд, требующих внешнего выполнения
}

// CommandResolver определяет команды в тексте
type CommandResolver struct {
	commands  []CommandDefinition
	threshold int
	synonyms  []synonymMapping // для быстрого поиска
}

type synonymMapping struct {
	Text     string
	CmdName  string
	External bool
}

// NewResolver создает новый резолвер команд
func NewResolver(filepath string, threshold int) (*CommandResolver, error) {
	// Загружаем JSON файл
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var commands []CommandDefinition
	if err := json.Unmarshal(data, &commands); err != nil {
		return nil, err
	}

	// Преобразуем в плоский список синонимов для быстрого поиска
	var synonyms []synonymMapping
	for _, cmd := range commands {
		for _, s := range cmd.Synonyms {
			synonyms = append(synonyms, synonymMapping{
				Text:     strings.ToLower(strings.TrimSpace(s)),
				CmdName:  cmd.Name,
				External: cmd.External,
			})
		}
	}

	// Сортируем от длинных к коротким (более точные совпадения в начале)
	sort.Slice(synonyms, func(i, j int) bool {
		return len(synonyms[i].Text) > len(synonyms[j].Text)
	})

	return &CommandResolver{
		commands:  commands,
		threshold: threshold,
		synonyms:  synonyms,
	}, nil
}

// Resolve определяет команду в тексте (проверяет всю строку целиком)
func (r *CommandResolver) Resolve(text string) (string, bool) {
	cleanText := strings.ToLower(strings.TrimSpace(text))
	// log.Printf("[COMMAND] Checking: '%s' (cleaned: '%s')", text, cleanText)
	// log.Printf("[COMMAND] Available synonyms: %d", len(r.synonyms))

	// 1. Точное совпадение (вся строка равна синониму)
	for _, mapping := range r.synonyms {
		// log.Printf("[COMMAND] Comparing '%s' with synonym '%s'", cleanText, mapping.Text)
		if cleanText == mapping.Text {
			// log.Printf("[COMMAND] ✓ Exact match found: %s -> %s", cleanText, mapping.CmdName)
			return mapping.CmdName, mapping.External
		}
	}

	// 2. Нечеткое совпадение по Левенштейну (вся строка целиком)
	bestMatch := ""
	bestExternal := false
	minDist := r.threshold + 1

	for _, mapping := range r.synonyms {
		dist := levenshtein.ComputeDistance(cleanText, mapping.Text)
		// log.Printf("[COMMAND] Levenshtein '%s' vs '%s' = %d (threshold: %d)", cleanText, mapping.Text, dist, r.threshold)

		if dist <= r.threshold && dist < minDist {
			minDist = dist
			bestMatch = mapping.CmdName
			bestExternal = mapping.External
		}
	}

	if bestMatch != "" {
		log.Printf("[COMMAND] ✓ Fuzzy match found: %s -> %s (dist=%d)", cleanText, bestMatch, minDist)
	} else {
		log.Printf("[COMMAND] ✗ No match found for: '%s'", cleanText)
	}

	return bestMatch, bestExternal
}

// GetCommand возвращает полное определение команды по имени
func (r *CommandResolver) GetCommand(name string) *CommandDefinition {
	for _, cmd := range r.commands {
		if cmd.Name == name {
			return &cmd
		}
	}
	return nil
}
