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
	// docs хранит TF-IDF вектор каждого документа: docID -> (лемма -> вес).
	// Используется при ранжировании для вычисления скалярного
	// произведения с вектором запроса.
	docs map[string]map[string]float64

	// idf хранит IDF-вес каждой леммы по всему корпусу.
	// Нужен для построения TF-IDF вектора запроса
	// с теми же весами, что и у документов.
	idf map[string]float64

	titles map[string]string
	urls   map[string]string

	// docRunes хранит полный текст документа как []rune.
	// Используется для извлечения сниппета по позициям токенов.
	docRunes map[string][]rune

	// docTokens - предвычисленные токены каждого документа
	// с позициями и леммами. Нужны для быстрого выбора сниппета:
	// находим токены, совпавшие с запросом, и выбираем окно
	// с максимальным покрытием.
	docTokens map[string][]docToken

	// docNorms - предвычисленная L2-норма TF-IDF вектора
	// каждого документа. Считается один раз при загрузке,
	// чтобы не пересчитывать при каждом запросе
	// (cosine = dot / (normQuery * normDoc)).
	docNorms map[string]float64

	lem *Lemmatizer
}

type docInfo struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	URL   string `json:"url"`
}

// docToken - токен документа с позицией в тексте (в рунах).
// Позиции нужны для вырезания сниппета из docRunes.
type docToken struct {
	pos   int    // начало токена в []rune
	end   int    // конец токена в []rune
	lemma string // нормальная форма слова
	word  string // оригинальная форма (для подсветки в сниппете)
}

// New загружает данные из dataDir и подготавливает движок к поиску.
func New(dataDir string) (*Engine, error) {
	e := &Engine{}

	// Загружаем TF-IDF вектора документов (результат gen_data.py).
	if err := loadJSON(filepath.Join(dataDir, "index.json"), &e.docs); err != nil {
		return nil, err
	}

	// Загружаем IDF-веса лемм - нужны для построения
	// вектора запроса с теми же весами, что у документов.
	if err := loadJSON(filepath.Join(dataDir, "idf.json"), &e.idf); err != nil {
		return nil, err
	}

	// Загружаем метаданные документов (заголовок, текст, URL)
	// и раскладываем по отдельным картам для быстрого доступа.
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

	// Инициализируем лемматизатор (SteosMorphy).
	// При первом запуске собирает словарь morph.dawg из module cache.
	lem, err := NewLemmatizer(filepath.Join(dataDir, "morph.dawg"))
	if err != nil {
		return nil, err
	}
	e.lem = lem

	// Предвычисляем L2-нормы документов и токенизируем тексты -
	// оба шага выполняются один раз, чтобы не тратить время
	// при каждом поисковом запросе.
	e.precomputeNorms()
	e.precomputeTokens(texts)
	return e, nil
}

// loadJSON читает JSON-файл и десериализует в target.
func loadJSON(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
