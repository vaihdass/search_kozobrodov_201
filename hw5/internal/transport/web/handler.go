package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/vaihdass/search_kozobrodov_201/search/internal/service"
)

// Handler обслуживает HTTP запросы к поисковой системе.
type Handler struct {
	svc  *service.SearchService
	tmpl *template.Template
}

type searchResponse struct {
	Query   string           `json:"query"`
	Results []service.Result `json:"results"`
	Count   int              `json:"count"`
}

// NewHandler создает HTTP handler с шаблоном.
func NewHandler(
	svc *service.SearchService, tmplPath string,
) (*Handler, error) {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, err
	}
	return &Handler{svc: svc, tmpl: tmpl}, nil
}

// Register регистрирует маршруты на переданном mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", h.index)
	mux.HandleFunc("/api/search", h.search)
}

// index обрабатывает запросы к корневому пути и отображает HTML шаблон поисковика.
func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	h.tmpl.Execute(w, nil)
}

// search обрабатывает запросы к /api/search, выполняет поиск и возвращает результаты в формате JSON.
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v <= 200 {
			limit = v
		}
	}

	results := h.svc.Search(query, limit)
	if results == nil {
		results = []service.Result{}
	}

	w.Header().Set(
		"Content-Type", "application/json; charset=utf-8",
	)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(searchResponse{
		Query:   query,
		Results: results,
		Count:   len(results),
	})
}
