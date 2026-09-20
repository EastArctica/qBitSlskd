package qbittorrent

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/EastArctica/qbitslskd/models"
)

func CategoriesHandler(w http.ResponseWriter, req *http.Request, cache *models.Cache) {
	cache.Mutex.Lock()
	data, err := json.Marshal(cache.Categories)
	cache.Mutex.Unlock()

	if err != nil {
		// TODO: idk qbit says it returns "200 in all scenarios" but this is most definitely a failure
		fmt.Fprintf(w, "{}")
		return
	}

	w.Header().Set("response-type", "application/json")
	fmt.Fprint(w, string(data))
}

func CreateCategoryHandler(w http.ResponseWriter, req *http.Request, cache *models.Cache) {
	req.ParseMultipartForm(10 << 10) // 10MB
	category := req.PostForm.Get("category")
	if category == "" {
		http.Error(w, "Category cannot be empty", 400)
		return
	}

	// Check if category already exists
	cache.Mutex.Lock()
	_, ok := cache.Categories[category]
	cache.Mutex.Lock()

	if !ok {
		http.Error(w, "Unable to create category", http.StatusConflict)
		return
	}

	// TODO: In qBit, the category has other restrictions. In this we'll ignore them.

	cache.Mutex.Lock()
	cache.Categories[category] = models.Category{
		Name:     category,
		SavePath: req.Form.Get("savePath"),
	}
	cache.Mutex.Unlock()

	CategoriesHandler(w, req, cache)
}
