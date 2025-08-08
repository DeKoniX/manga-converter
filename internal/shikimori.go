package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Metadata struct {
	Title       string
	Author      string
	Description string
	Genres      string
	URL         string
	CoverURL    string
}

type shikimoriResponse struct {
	Name    string `json:"name"`
	Russian string `json:"russian"`
	URL     string `json:"url"`
	Image   struct {
		Original string `json:"original"`
	} `json:"image"`
	Description string   `json:"description"`
	Genres      []string `json:"genres"`
}

func FetchMetadata(name string) (*Metadata, error) {
	query := strings.ReplaceAll(name, "_", " ")
	log.Printf("🔎 Запрос Shikimori по имени: %s", query)

	req, err := http.NewRequest("GET", "https://shikimori.one/api/mangas", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("search", query)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", "manga-converter")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("статус ответа %d", resp.StatusCode)
	}

	var results []shikimoriResponse
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("манга не найдена")
	}

	manga := results[0]
	genres := strings.Join(manga.Genres, ", ")

	return &Metadata{
		Title:       manga.Russian,
		Author:      "", // Shikimori не всегда указывает
		Description: manga.Description,
		Genres:      genres,
		URL:         "https://shikimori.one" + manga.URL,
		CoverURL:    "https://shikimori.one" + manga.Image.Original,
	}, nil
}
