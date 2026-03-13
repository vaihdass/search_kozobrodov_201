package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/vaihdass/search_kozobrodov_201/search/internal/service"
)

// Run выполняет поиск и выводит результаты в stdout.
func Run(svc *service.SearchService) {
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run ./cmd/cli \"запрос\"")
		os.Exit(1)
	}

	query := strings.Join(os.Args[1:], " ")
	results := svc.Search(query, 10)

	if len(results) == 0 {
		fmt.Printf("По запросу \"%s\" ничего не найдено\n", query)
		return
	}

	fmt.Printf("Результаты по запросу \"%s\":\n\n", query)
	for i, r := range results {
		fmt.Printf("%d. %s (score: %.4f)\n", i+1, r.Title, r.Score)
		fmt.Printf("   %s\n", r.URL)
		fmt.Printf("   %s\n\n", r.Snippet)
	}
}
