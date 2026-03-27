package search

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	steosmorphy "github.com/steosofficial/steosmorphy/analyzer"
)

const steosModule = "github.com/steosofficial/steosmorphy"

// Lemmatizer оборачивает морфологический анализатор SteosMorphy
// с кешем уже обработанных слов.
//
// SteosMorphy работает со словарём morph.dawg (~445 МБ) -
// конечным автоматом, который хранит все словоформы русского языка
// и позволяет быстро находить нормальную форму (лемму) любого слова.
type Lemmatizer struct {
	morph *steosmorphy.MorphAnalyzer
	// cache позволяет не обращаться к словарю дважды для одной словоформы.
	// В реальных текстах одни и те же слова встречаются много раз,
	// поэтому кеш существенно ускоряет предвычисление токенов.
	cache map[string]string
}

// NewLemmatizer инициализирует SteosMorphy из файла словаря dawgPath.
// Если morph.dawg не найден - собирает его из Go module cache.
//
// Словарь разбит на части (morph_a, morph_b, ...) прямо внутри
// Go-модуля steosmorphy, чтобы обойти ограничение git на размер файлов.
// При первом запуске части склеиваются в один файл morph.dawg.
func NewLemmatizer(dawgPath string) (*Lemmatizer, error) {
	if _, err := os.Stat(dawgPath); err != nil {
		// Файл не найден - собираем из частей
		fmt.Println("morph.dawg not found, building from Go module cache...")
		if buildErr := buildMorphDawg(dawgPath); buildErr != nil {
			return nil, fmt.Errorf("build morph.dawg: %w", buildErr)
		}
	}

	// Передаём путь к словарю через переменную окружения -
	// так требует API SteosMorphy
	os.Setenv("STEOSMORPHY_DICT_PATH", dawgPath)
	morph, err := steosmorphy.LoadMorphAnalyzer()
	if err != nil {
		return nil, err
	}

	return &Lemmatizer{morph: morph, cache: make(map[string]string)}, nil
}

// Lemmatize возвращает нормальную форму (лемму) русского слова.
// Например: "курицей" -> "курица", "приготовленного" -> "приготовленный".
//
// Результаты кешируются: повторный вызов с тем же словом
// возвращает сохранённое значение без обращения к словарю.
func (l *Lemmatizer) Lemmatize(word string) string {
	if v, ok := l.cache[word]; ok {
		return v
	}

	// Analyze возвращает все возможные разборы слова.
	// Берём первый - он обычно наиболее вероятный по частоте.
	var lemma string
	parses, _ := l.morph.Analyze(word)
	if len(parses) > 0 {
		lemma = parses[0].Lemma
	}

	l.cache[word] = lemma
	return lemma
}

// buildMorphDawg собирает morph.dawg из частей в Go module cache.
//
// SteosMorphy хранит словарь разбитым на файлы morph_a, morph_b, ...
// в директории analyzer/ внутри своего модуля. Находим модуль через
// "go list", сортируем части по имени и склеиваем в один файл.
func buildMorphDawg(targetPath string) error {
	// Находим директорию модуля steosmorphy в локальном кеше Go
	modDir, err := goModuleDir(steosModule)
	if err != nil {
		return fmt.Errorf("locate module: %w", err)
	}

	// Ищем все файлы с префиксом morph_ в поддиректории analyzer/
	partsDir := filepath.Join(modDir, "analyzer")
	entries, err := os.ReadDir(partsDir)
	if err != nil {
		return fmt.Errorf("read parts dir: %w", err)
	}

	var parts []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "morph_") {
			parts = append(parts, filepath.Join(partsDir, e.Name()))
		}
	}
	if len(parts) == 0 {
		return fmt.Errorf("no morph_* parts found in %s", partsDir)
	}
	// Сортируем по имени, чтобы части склеились в правильном порядке
	sort.Strings(parts)

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Последовательно копируем части в выходной файл
	for _, p := range parts {
		f, fErr := os.Open(p)
		if fErr != nil {
			return fErr
		}
		_, fErr = io.Copy(out, f)
		f.Close()
		if fErr != nil {
			return fErr
		}
	}

	fmt.Printf(
		"building morph.dawg from %d parts (first run only)...\n",
		len(parts),
	)
	return nil
}

// goModuleDir возвращает путь к директории Go-модуля в локальном кеше.
// Использует "go list -m -json" для получения метаданных модуля.
func goModuleDir(module string) (string, error) {
	out, err := exec.Command("go", "list", "-m", "-json", module).Output()
	if err != nil {
		return "", err
	}

	var info struct{ Dir string }
	if err = json.Unmarshal(out, &info); err != nil {
		return "", err
	}
	if info.Dir == "" {
		return "", fmt.Errorf("module %s: Dir is empty", module)
	}
	return info.Dir, nil
}
