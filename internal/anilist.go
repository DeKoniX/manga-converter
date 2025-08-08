package internal

type AniResponse struct {
	Data struct {
		Media struct {
			Title struct {
				Romaji  string `json:"romaji"`
				English string `json:"english"`
			} `json:"title"`
			Description string   `json:"description"`
			Genres      []string `json:"genres"`
			CoverImage  struct {
				ExtraLarge string `json:"extraLarge"`
			} `json:"coverImage"`
			SiteURL string `json:"siteUrl"`
			Staff   struct {
				Edges []struct {
					Node struct {
						Name struct {
							Full string `json:"full"`
						} `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"staff"`
		} `json:"Media"`
	} `json:"data"`
}

// type Metadata struct {
// 	Title       string
// 	Author      string
// 	Description string
// 	Genres      string
// 	URL         string
// 	CoverURL    string
// }

// func FetchMetadata(title string) (*Metadata, error) {
// 	query := `query ($search: String) {
//       Media(search: $search, type: MANGA) {
//         title { romaji english }
//         description(asHtml: false)
//         genres
//         coverImage { extraLarge }
//         siteUrl
//         staff {
//           edges { node { name { full } } }
//         }
//       }
//     }`

// 	body := map[string]interface{}{
// 		"query": query,
// 		"variables": map[string]string{
// 			"search": title,
// 		},
// 	}
// 	jsonBody, _ := json.Marshal(body)

// 	log.Printf("🔍 Запрос к AniList для: %s", title)
// 	resp, err := http.Post("https://graphql.anilist.co", "application/json", bytes.NewBuffer(jsonBody))
// 	if err != nil {
// 		return nil, fmt.Errorf("ошибка запроса: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != 200 {
// 		return nil, fmt.Errorf("статус ответа %d", resp.StatusCode)
// 	}

// 	var result AniResponse
// 	err = json.NewDecoder(resp.Body).Decode(&result)
// 	if err != nil {
// 		return nil, fmt.Errorf("декодирование ответа: %w", err)
// 	}

// 	m := result.Data.Media
// 	log.Printf("✅ Найдено: %s / %s", m.Title.English, m.Title.Romaji)

// 	return &Metadata{
// 		Title:       chooseFirst(m.Title.English, m.Title.Romaji),
// 		Author:      getAuthor(m.Staff.Edges),
// 		Description: m.Description,
// 		Genres:      joinGenres(m.Genres),
// 		URL:         m.SiteURL,
// 		CoverURL:    m.CoverImage.ExtraLarge,
// 	}, nil
// }

// func chooseFirst(a, b string) string {
// 	if a != "" {
// 		return a
// 	}
// 	return b
// }

// func getAuthor(edges []struct {
// 	Node struct {
// 		Name struct {
// 			Full string `json:"full"`
// 		} `json:"name"`
// 	} `json:"node"`
// }) string {
// 	if len(edges) > 0 {
// 		return edges[0].Node.Name.Full
// 	}
// 	return ""
// }

// func joinGenres(genres []string) string {
// 	return fmt.Sprintf("%s", genres)
// }
