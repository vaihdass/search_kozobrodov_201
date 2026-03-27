package search

import "strings"

// snippetLen - длина окна сниппета в рунах.
const snippetLen = 200

// precomputeTokens лемматизирует тексты всех документов при запуске.
// Результат сохраняется в e.docTokens и e.docRunes.
//
// Делаем это один раз заранее, чтобы при каждом запросе не тратить
// время на повторную токенизацию и лемматизацию текстов.
// Сохраняем только токены, чьи леммы есть в IDF-индексе -
// остальные слова всё равно не будут совпадать ни с одним запросом.
func (e *Engine) precomputeTokens(texts map[string]string) {
	e.docRunes = make(map[string][]rune, len(texts))
	e.docTokens = make(map[string][]docToken, len(texts))

	for id, text := range texts {
		if text == "" {
			continue
		}

		// Преобразуем текст в []rune, чтобы работать с позициями
		// символов, а не байт. В utf-8 один кириллический символ
		// занимает 2 байта, поэтому срезы по байтам ломают текст.
		runes := []rune(text)
		e.docRunes[id] = runes

		// Регэксп tokenRe работает с байтовыми смещениями (loc[0], loc[1]).
		// Чтобы перевести их в позиции рун (нужны для вырезания сниппета),
		// строим таблицу: байтовая позиция -> позиция руны.
		byteToRune := makeByteToRuneIndex(text)
		locs := tokenRe.FindAllStringIndex(text, -1)

		// Резервируем память под ~25% токенов от общего числа -
		// большинство слов не будут в IDF-индексе.
		tokens := make([]docToken, 0, len(locs)/4)

		for _, loc := range locs {
			// Переводим байтовые позиции в позиции рун
			rStart := byteToRune[loc[0]]
			rEnd := byteToRune[loc[1]]
			word := string(runes[rStart:rEnd])

			lemma := e.lem.Lemmatize(strings.ToLower(word))
			if lemma == "" {
				continue
			}

			// Пропускаем слова, которых нет в IDF-индексе.
			// Такие леммы не встречались ни в одном документе
			// при построении индекса - их вес при поиске равен 0.
			if _, ok := e.idf[lemma]; !ok {
				continue
			}

			tokens = append(tokens, docToken{
				rStart, rEnd, lemma, word,
			})
		}
		e.docTokens[id] = tokens
	}
}

// makeSnippet выбирает наиболее релевантный фрагмент текста документа.
// Возвращает строку сниппета и словоформы из запроса для подсветки.
//
// Алгоритм: скользящее окно шириной snippetLen рун.
// Для каждого токена, совпавшего с запросом (hit), центрируем окно на нём
// и считаем сколько различных лемм запроса попало в окно.
// Побеждает окно с максимальным покрытием.
func (e *Engine) makeSnippet(
	docID string, queryVec map[string]float64,
) (string, []string) {
	runes := e.docRunes[docID]
	if len(runes) == 0 {
		return "", nil
	}

	// Оставляем только те токены документа, чьи леммы есть в запросе.
	// Это кандидаты для центрирования окна сниппета.
	allTokens := e.docTokens[docID]
	hits := make([]docToken, 0, len(queryVec))
	for _, t := range allTokens {
		if _, ok := queryVec[t.lemma]; ok {
			hits = append(hits, t)
		}
	}

	// Перебираем каждый hit как потенциальный центр сниппета.
	// Для каждого варианта считаем уникальные леммы запроса в окне.
	// Запоминаем начало окна с наибольшим числом уникальных лемм.
	bestStart := 0
	if len(hits) > 0 {
		bestScore := 0
		for _, h := range hits {
			// Центрируем окно на текущем hit
			start := h.pos - snippetLen/2
			if start < 0 {
				start = 0
			}
			end := start + snippetLen
			if end > len(runes) {
				end = len(runes)
				start = max(0, end-snippetLen)
			}

			// Считаем сколько различных лемм запроса попало в окно
			seen := map[string]bool{}
			for _, hh := range hits {
				if hh.pos >= start && hh.pos < end {
					seen[hh.lemma] = true
				}
			}
			if len(seen) > bestScore {
				bestScore = len(seen)
				bestStart = start
			}
		}
	}

	// Вырезаем финальный сниппет по лучшей найденной позиции
	end := bestStart + snippetLen
	if end > len(runes) {
		end = len(runes)
	}

	snippet := string(runes[bestStart:end])
	// Добавляем многоточие если сниппет вырезан из середины текста
	if bestStart > 0 {
		snippet = "..." + snippet
	}
	if end < len(runes) {
		snippet = snippet + "..."
	}

	// Собираем уникальные словоформы для подсветки в UI.
	// Отдаём оригинальные формы (не леммы), как они написаны в тексте.
	seen := map[string]bool{}
	var highlights []string
	for _, h := range hits {
		if h.pos >= bestStart && h.pos < end && !seen[h.word] {
			seen[h.word] = true
			highlights = append(highlights, h.word)
		}
	}

	return snippet, highlights
}

// makeByteToRuneIndex строит таблицу перевода байтовых позиций в позиции рун.
//
// Нужна потому что regexp.FindAllStringIndex возвращает байтовые смещения,
// а нам нужны позиции рун для работы с []rune текста документа.
// Таблица имеет размер len(s)+1 и для каждого байта хранит индекс руны.
func makeByteToRuneIndex(s string) []int {
	idx := make([]int, len(s)+1)
	ri := 0
	for bi := range s {
		idx[bi] = ri
		ri++
	}
	idx[len(s)] = ri
	return idx
}
