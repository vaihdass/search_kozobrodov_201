package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/vaihdass/search_kozobrodov_201/search/internal/service"
	"github.com/vaihdass/search_kozobrodov_201/search/internal/transport/web"
)

func main() {
	svc, err := service.New("data")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка загрузки: %v\n", err)
		os.Exit(1)
	}

	handler, err := web.NewHandler(svc, "templates/index.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка шаблона: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	handler.Register(mux)

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Printf("http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
