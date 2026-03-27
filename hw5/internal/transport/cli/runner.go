package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/vaihdass/search_kozobrodov_201/search/internal/service"
)

// RunInteractive запускает интерактивный REPL поиска.
func RunInteractive(svc *service.SearchService) {
	fmt.Println(
		"Векторный поиск. " +
			"Для выхода введите \"exit\" " +
			"или нажмите Ctrl+C.",
	)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nЗапрос: ")
		if !scanner.Scan() {
			break
		}

		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		}
		if strings.EqualFold(query, "exit") {
			break
		}

		searchAndPrint(svc, query)
	}

	fmt.Println()
}

// RunOnce выполняет один поиск и завершается.
func RunOnce(svc *service.SearchService, query string) {
	searchAndPrint(svc, query)
}

func searchAndPrint(svc *service.SearchService, query string) {
	results := svc.Search(query, 10)

	if len(results) == 0 {
		fmt.Printf(
			"По запросу \"%s\" ничего не найдено\n",
			query,
		)
		return
	}

	fmt.Printf("Результаты по запросу \"%s\":\n\n", query)
	for i, r := range results {
		fmt.Printf(
			"%d. %s (score: %.4f)\n",
			i+1, r.Title, r.Score,
		)
		fmt.Printf("   %s\n", r.URL)
		fmt.Printf("   %s\n\n", r.Snippet)
	}
}
