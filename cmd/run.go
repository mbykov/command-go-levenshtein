package main

import (
    "flag"
    "fmt"
    "log"
    "strings"
	"os"
    "gopkg.in/yaml.v3"

    "github.com/chzyer/readline"
    "github.com/mbykov/command-go-levenshtein" // импорт прямо из корня!
)

func main() {
    configPath := flag.String("config", "cmd/config.yaml", "путь к конфигу")
    flag.Parse()


	// Загружаем конфиг
	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфига: %v", err)
	}

	// Создаем резолвер
    resolver, err := command.NewResolver(cfg.CommandsFile, cfg.Threshold)
	if err != nil {
		log.Fatalf("❌ Ошибка создания резолвера: %v", err)
	}

	fmt.Printf("✅ Резолвер загружен (порог = %d)\n", cfg.Threshold)
	fmt.Printf("📁 Команды из: %s\n", cfg.CommandsFile)
	fmt.Println("📜 Введите фразу или 'exit' для выхода:")

	rl, err := readline.New("> ")
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}

		phrase := strings.TrimSpace(line)
		if phrase == "" {
			continue
		}
		if phrase == "exit" || phrase == "quit" {
			fmt.Println("Выход.")
			break
		}

		cmdName, external := resolver.Resolve(phrase)
		if cmdName == "" {
			fmt.Printf("❌ Команда не найдена.\n")
		} else {
			cmd := resolver.GetCommand(cmdName)
			fmt.Printf("✅ Найдена команда: %s\n", cmdName)
			if cmd != nil {
				fmt.Printf("   Синонимы: %v\n", cmd.Synonyms)
			}
			fmt.Printf("   Внешняя: %v\n", external)
		}
	}
}

// Config для example (не экспортируется)
type config struct {
	CommandsFile string `yaml:"commands_file"`
	Threshold    int    `yaml:"threshold"`
}

func loadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Если файла нет, используем значения по умолчанию
		return &config{
			CommandsFile: "./data/commands.json",
			Threshold:    3,
		}, nil
	}

	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
