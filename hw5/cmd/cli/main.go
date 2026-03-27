package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vaihdass/search_kozobrodov_201/search/internal/service"
	"github.com/vaihdass/search_kozobrodov_201/search/internal/transport/cli"
)

func main() {
	query := flag.String("i", "", "one-shot search query")
	flag.Parse()

	svc, err := service.New("data")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка загрузки: %v\n", err)
		os.Exit(1)
	}

	if *query != "" {
		cli.RunOnce(svc, *query)
		return
	}

	cli.RunInteractive(svc)
}
