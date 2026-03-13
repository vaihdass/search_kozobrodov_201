package search

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
)

var tokenRe = regexp.MustCompile(`[а-яёА-ЯЁ]+`)

// Engine - поисковый движок на основе TF-IDF векторов.
type Engine struct {
	docs      map[string]map[string]float64
	idf       map[string]float64
	titles    map[string]string
	urls      map[string]string
	docRunes  map[string][]rune
	docTokens map[string][]docToken
	docNorms  map[string]float64
	lem       *Lemmatizer
}

type docInfo struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	URL   string `json:"url"`
}

type docToken struct {
	pos   int
	end   int
	lemma string
	word  string
}

// New загружает данные из dataDir и подготавливает движок к поиску.
func New(dataDir string) (*Engine, error) {
	e := &Engine{}

	if err := loadJSON(filepath.Join(dataDir, "index.json"), &e.docs); err != nil {
		return nil, err
	}
	if err := loadJSON(filepath.Join(dataDir, "idf.json"), &e.idf); err != nil {
		return nil, err
	}

	var docs map[string]docInfo
	if err := loadJSON(filepath.Join(dataDir, "docs.json"), &docs); err != nil {
		return nil, err
	}
	e.titles = make(map[string]string, len(docs))
	e.urls = make(map[string]string, len(docs))
	texts := make(map[string]string, len(docs))
	for id, d := range docs {
		e.titles[id] = d.Title
		e.urls[id] = d.URL
		texts[id] = d.Text
	}

	lem, err := NewLemmatizer(filepath.Join(dataDir, "morph.dawg"))
	if err != nil {
		return nil, err
	}
	e.lem = lem

	e.precomputeNorms()
	e.precomputeTokens(texts)
	return e, nil
}

func loadJSON(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
