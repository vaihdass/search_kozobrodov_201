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

// Lemmatizer wraps SteosMorphy morphological analyzer with an in-memory cache.
type Lemmatizer struct {
	morph *steosmorphy.MorphAnalyzer
	cache map[string]string
}

// NewLemmatizer initializes SteosMorphy.
// If morph.dawg not found at dawgPath, builds it from Go module cache.
func NewLemmatizer(dawgPath string) (*Lemmatizer, error) {
	if _, err := os.Stat(dawgPath); err != nil {
		fmt.Println("morph.dawg not found, building from Go module cache...")
		if buildErr := buildMorphDawg(dawgPath); buildErr != nil {
			return nil, fmt.Errorf("build morph.dawg: %w", buildErr)
		}
	}

	os.Setenv("STEOSMORPHY_DICT_PATH", dawgPath)
	morph, err := steosmorphy.LoadMorphAnalyzer()
	if err != nil {
		return nil, err
	}

	return &Lemmatizer{morph: morph, cache: make(map[string]string)}, nil
}

// Lemmatize returns the lemma (normal form) of a Russian word.
// Results are cached to avoid repeated calls for the same word form.
func (l *Lemmatizer) Lemmatize(word string) string {
	if v, ok := l.cache[word]; ok {
		return v
	}
	var lemma string
	parses, _ := l.morph.Analyze(word)
	if len(parses) > 0 {
		lemma = parses[0].Lemma
	}
	l.cache[word] = lemma
	return lemma
}

// buildMorphDawg merges morph_a* parts from Go module cache into target path.
func buildMorphDawg(targetPath string) error {
	modDir, err := goModuleDir(steosModule)
	if err != nil {
		return fmt.Errorf("locate module: %w", err)
	}

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
	sort.Strings(parts)

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

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

	fmt.Printf("building morph.dawg from %d parts (first run only)...\n", len(parts))
	return nil
}

// goModuleDir returns the cached directory for a Go module via "go list -m -json".
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
