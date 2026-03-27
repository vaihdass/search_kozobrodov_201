package search

import (
	"math"
	"sort"
	"strings"
)

// Result - результат поиска по одному документу.
type Result struct {
	Title      string
	URL        string
	Snippet    string
	Highlights []string
	Score      float64
}

type scoredDoc struct {
	docID string
	score float64
}

// Search - основной метод поиска.
// 1. Строит TF-IDF вектор запроса
// 2. Считает косинусное сходство с каждым документом
// 3. Сортирует по убыванию релевантности и берет топ-k
// 4. Для каждого результата генерирует сниппет с подсветкой
func (e *Engine) Search(query string, topK int) []Result {
	// Шаг 1: лемматизация запроса и построение TF-IDF вектора
	queryVec, queryNorm := e.buildQueryVector(query)
	if queryNorm == 0 {
		return nil
	}

	// Шаг 2: ранжирование документов по косинусному сходству
	candidates := e.rankDocuments(queryVec, queryNorm)

	// Шаг 3: оставляем только топ-k
	candidates = e.topK(candidates, topK)

	// Шаг 4: формируем результаты со сниппетами
	return e.buildResults(candidates, queryVec)
}

// rankDocuments оценивает каждый документ по близости к запросу.
//
// Косинусное сходство показывает насколько два вектора "смотрят в одну сторону":
// берем сумму произведений совпадающих компонент (скалярное произведение)
// и делим на произведение длин обоих векторов (нормализация).
// Результат от 0 (ничего общего) до 1 (полное совпадение направлений).
//
// Документы, у которых нет ни одного общего слова с запросом, пропускаются.
func (e *Engine) rankDocuments(
	queryVec map[string]float64, queryNorm float64,
) []scoredDoc {
	var candidates []scoredDoc

	for docID, docVec := range e.docs {
		// Скалярное произведение: суммируем произведения весов
		// совпадающих лемм в запросе и документе
		dot := 0.0
		for lemma, qVal := range queryVec {
			if dVal, ok := docVec[lemma]; ok {
				dot += qVal * dVal
			}
		}
		if dot == 0 {
			continue
		}

		// Делим на длины векторов чтобы длинные документы
		// не получали преимущество просто за счет размера
		cosine := dot / (queryNorm * e.docNorms[docID])
		candidates = append(candidates, scoredDoc{docID, cosine})
	}

	return candidates
}

// topK сортирует кандидатов по убыванию релевантности и обрезает до topK.
// topK <= 0 означает "вернуть все".
func (e *Engine) topK(
	candidates []scoredDoc, topK int,
) []scoredDoc {
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if topK > 0 && len(candidates) > topK {
		candidates = candidates[:topK]
	}
	return candidates
}

// buildResults собирает финальные результаты для отображения пользователю.
// Для каждого отранжированного документа подбирает наиболее релевантный
// фрагмент текста (сниппет) и собирает список слов для подсветки.
func (e *Engine) buildResults(
	candidates []scoredDoc, queryVec map[string]float64,
) []Result {
	results := make([]Result, len(candidates))
	for i, c := range candidates {
		snippet, highlights := e.makeSnippet(c.docID, queryVec)
		results[i] = Result{
			Title:      e.titles[c.docID],
			URL:        e.urls[c.docID],
			Snippet:    snippet,
			Highlights: highlights,
			Score:      c.score,
		}
	}
	return results
}

// precomputeNorms предвычисляет длины TF-IDF векторов всех документов.
// Длина вектора - это корень из суммы квадратов всех его компонент.
// Нужна для нормализации при вычислении косинусного сходства,
// чтобы длинные и короткие документы сравнивались на равных.
func (e *Engine) precomputeNorms() {
	e.docNorms = make(map[string]float64, len(e.docs))
	for id, vec := range e.docs {
		sum := 0.0
		for _, v := range vec {
			sum += v * v
		}
		e.docNorms[id] = math.Sqrt(sum)
	}
}

// buildQueryVector строит TF-IDF вектор для поискового запроса.
//
// TF (частота термина) - как часто слово встречается в запросе,
// считается как доля: количество вхождений / общее число слов.
//
// IDF (обратная документная частота) - насколько слово редкое в коллекции.
// Чем реже слово встречается среди документов, тем выше его вес.
// Берется из предвычисленного индекса.
//
// TF-IDF = TF * IDF. Частое в запросе и редкое в коллекции слово
// получает наибольший вес.
//
// Возвращает вектор (лемма -> вес) и его длину для нормализации.
func (e *Engine) buildQueryVector(
	query string,
) (map[string]float64, float64) {
	// Выделяем из запроса кириллические слова
	words := tokenRe.FindAllString(query, -1)
	if len(words) == 0 {
		return nil, 0
	}

	// Приводим каждое слово к нормальной форме (лемме)
	// и считаем сколько раз каждая лемма встретилась
	counts := make(map[string]int)
	total := 0
	for _, w := range words {
		lemma := e.lem.Lemmatize(strings.ToLower(w))
		if lemma == "" {
			continue
		}
		// Пропускаем леммы которых нет в индексе -
		// они не встречались ни в одном документе
		if _, ok := e.idf[lemma]; !ok {
			continue
		}
		counts[lemma]++
		total++
	}
	if total == 0 {
		return nil, 0
	}

	// Считаем TF-IDF вес для каждой леммы запроса
	vec := make(map[string]float64, len(counts))
	norm := 0.0
	for lemma, cnt := range counts {
		tf := float64(cnt) / float64(total)
		tfidf := tf * e.idf[lemma]
		vec[lemma] = tfidf
		norm += tfidf * tfidf
	}
	return vec, math.Sqrt(norm)
}
