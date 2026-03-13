package main

import (
	"fmt"
	"os"

	"github.com/vaihdass/search_kozobrodov_201/search/internal/service"
	"github.com/vaihdass/search_kozobrodov_201/search/internal/transport/cli"
)

func main() {
	svc, err := service.New("data")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка загрузки: %v\n", err)
		os.Exit(1)
	}

	cli.Run(svc)
}
