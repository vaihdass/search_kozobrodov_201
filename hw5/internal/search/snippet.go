package search

import "strings"

const snippetLen = 200

// precomputeTokens лемматизирует все токены документов при запуске.
// Сохраняются только токены с леммами из IDF-индекса.
func (e *Engine) precomputeTokens(texts map[string]string) {
	e.docRunes = make(map[string][]rune, len(texts))
	e.docTokens = make(map[string][]docToken, len(texts))

	for id, text := range texts {
		if text == "" {
			continue
		}
		runes := []rune(text)
		e.docRunes[id] = runes

		byteToRune := makeByteToRuneIndex(text)
		locs := tokenRe.FindAllStringIndex(text, -1)
		tokens := make([]docToken, 0, len(locs)/4)

		for _, loc := range locs {
			rStart := byteToRune[loc[0]]
			rEnd := byteToRune[loc[1]]
			word := string(runes[rStart:rEnd])
			lemma := e.lem.Lemmatize(strings.ToLower(word))
			if lemma == "" {
				continue
			}
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

// makeSnippet выбирает окно с максимальным числом различных лемм запроса.
// Возвращает текст сниппета и словоформы для подсветки.
func (e *Engine) makeSnippet(
	docID string, queryVec map[string]float64,
) (string, []string) {
	runes := e.docRunes[docID]
	if len(runes) == 0 {
		return "", nil
	}

	allTokens := e.docTokens[docID]
	hits := make([]docToken, 0, len(queryVec))
	for _, t := range allTokens {
		if _, ok := queryVec[t.lemma]; ok {
			hits = append(hits, t)
		}
	}

	bestStart := 0
	if len(hits) > 0 {
		bestScore := 0
		for _, h := range hits {
			start := h.pos - snippetLen/2
			if start < 0 {
				start = 0
			}
			end := start + snippetLen
			if end > len(runes) {
				end = len(runes)
				start = max(0, end-snippetLen)
			}
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

	end := bestStart + snippetLen
	if end > len(runes) {
		end = len(runes)
	}

	snippet := string(runes[bestStart:end])
	if bestStart > 0 {
		snippet = "..." + snippet
	}
	if end < len(runes) {
		snippet = snippet + "..."
	}

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
