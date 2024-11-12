package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/Kdizzle8d8/Chatterbox/backend/filters"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type initialSearchResult struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func search(query string, limit int) ([]initialSearchResult, error) {

	apiKey := os.Getenv("SEARCH_API_KEY")
	cx := os.Getenv("SEARCH_CX")

	if apiKey == "" || cx == "" {
		return []initialSearchResult{}, fmt.Errorf("search API key or custom search engine ID is not set")
	}

	encodedQuery := url.QueryEscape(query)
	braveApiKey := os.Getenv("BRAVE_API_KEY")
	if braveApiKey == "" {
		return []initialSearchResult{}, fmt.Errorf("brave api key is not set")
	}

	url := fmt.Sprintf("https://api.search.brave.com/res/v1/web/search?q=%s", encodedQuery)
	fmt.Println("Sending request to:", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []initialSearchResult{}, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Add("X-Subscription-Token", "BSAmZReNAsbJj270ClDtIoR6ypQjH54")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []initialSearchResult{}, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []initialSearchResult{}, fmt.Errorf("search query failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []initialSearchResult{}, err
	}

	var fullRes FullSearch
	if err := json.Unmarshal(body, &fullRes); err != nil {
		return []initialSearchResult{}, err
	}
	results := fullRes.Web.Results
	if len(results) > 0 {
		newResults := []initialSearchResult{}
		for _, item := range results[:limit] {
			newResults = append(newResults, initialSearchResult{Title: item.Title, Link: item.URL})
		}
		return newResults, nil
	}

	return []initialSearchResult{}, nil
}

func fullSearch(queries []string, limit int) {
	resultsChan := make(chan finalResult, len(queries))
	var wg sync.WaitGroup

	for _, query := range queries {
		wg.Add(1)
		go processQuery(query, resultsChan, &wg, limit)
		time.Sleep(2 * time.Second)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var finalResults []finalResult
	for result := range resultsChan {
		finalResults = append(finalResults, result)
	}
	fmt.Println("\nSearch Results:")

	// Create table rows
	var rows [][]string
	for _, result := range finalResults {
		for _, r := range result.Results {
			rows = append(rows, []string{result.Query, r.Title, r.Link})
		}
	}

	// Create and style table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row%2 == 0:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
			default:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
			}
		}).
		Headers("QUERY", "TITLE", "LINK").
		Rows(rows...)

	fmt.Println(t)
	// for _, r := range finalResults {
	// 	for _, rr := range r.Results {
	// 		os.WriteFile(fmt.Sprintf("./results/%s.md", rr.Title), []byte(rr.Markdown), 0644)
	// 	}
	// }
}

func processQuery(query string, resultsChan chan<- finalResult, wg *sync.WaitGroup, limit int) {
	defer wg.Done()

	results, err := search(query, limit)
	if err != nil {
		fmt.Printf("Error searching for query '%s': %v\n", query, err)
		return
	}

	fr := finalResult{
		Query:   query,
		Results: []result{},
	}

	for _, page := range results {
		md := pageToMd(page)
		fr.Results = append(fr.Results, result{
			Title:    page.Title,
			Link:     page.Link,
			Markdown: md,
		})
	}

	resultsChan <- fr
}

func pageToMd(result initialSearchResult) string {
	fullPageResult, err := getFullPage(result)
	if err != nil {
		fmt.Println("Error during search:", err)
		return ""
	}
	text, err := filters.Filter(result.Link, fullPageResult.Html)
	if err != nil {
		fmt.Println("Error during full page search:", err)
		return ""
	}
	markdown, err := htmltomarkdown.ConvertString(text)
	if err != nil {
		fmt.Println("Error during text extraction:", err)
		return ""
	}
	return markdown
}
func getFullPage(result initialSearchResult) (fullPageResult, error) {
	resp, err := http.Get(result.Link)
	if err != nil {
		return fullPageResult{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fullPageResult{}, err
	}
	return fullPageResult{Title: resp.Header.Get("Title"), Link: result.Link, Html: string(body)}, nil
}

type fullPageResult struct {
	Title string `json:"title"`
	Link  string `json:"link"`
	Html  string `json:"html"`
}

type QueryResult struct {
	Query string                `json:"query"`
	Pages []initialSearchResult `json:"pages"`
}
type finalResult struct {
	Query   string   `json:"query"`
	Results []result `json:"results"`
}
type result struct {
	Title    string `json:"title"`
	Link     string `json:"link"`
	Markdown string `json:"markdown"`
}

type FullSearch struct {
	Query Query    `json:"query"`
	Mixed Mixed    `json:"mixed"`
	Type  string   `json:"type"`
	Web   WebClass `json:"web"`
}

type Mixed struct {
	Type string        `json:"type"`
	Main []Main        `json:"main"`
	Top  []interface{} `json:"top"`
	Side []interface{} `json:"side"`
}

type Main struct {
	Type  MainType `json:"type"`
	Index int64    `json:"index"`
	All   bool     `json:"all"`
}

type Query struct {
	Original             string `json:"original"`
	ShowStrictWarning    bool   `json:"show_strict_warning"`
	IsNavigational       bool   `json:"is_navigational"`
	IsNewsBreaking       bool   `json:"is_news_breaking"`
	SpellcheckOff        bool   `json:"spellcheck_off"`
	Country              string `json:"country"`
	BadResults           bool   `json:"bad_results"`
	ShouldFallback       bool   `json:"should_fallback"`
	PostalCode           string `json:"postal_code"`
	City                 string `json:"city"`
	HeaderCountry        string `json:"header_country"`
	MoreResultsAvailable bool   `json:"more_results_available"`
	State                string `json:"state"`
}

type WebClass struct {
	Type           string   `json:"type"`
	Results        []Result `json:"results"`
	FamilyFriendly bool     `json:"family_friendly"`
}

type Result struct {
	Title          string     `json:"title"`
	URL            string     `json:"url"`
	IsSourceLocal  bool       `json:"is_source_local"`
	IsSourceBoth   bool       `json:"is_source_both"`
	Description    string     `json:"description"`
	PageAge        string     `json:"page_age,omitempty"`
	Profile        Profile    `json:"profile"`
	Language       Language   `json:"language"`
	FamilyFriendly bool       `json:"family_friendly"`
	Type           ResultType `json:"type"`
	Subtype        string     `json:"subtype"`
	MetaURL        MetaURL    `json:"meta_url"`
	Thumbnail      *Thumbnail `json:"thumbnail,omitempty"`
	Age            *string    `json:"age,omitempty"`
}

type MetaURL struct {
	Scheme   Scheme `json:"scheme"`
	Netloc   string `json:"netloc"`
	Hostname string `json:"hostname"`
	Favicon  string `json:"favicon"`
	Path     string `json:"path"`
}

type Profile struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	LongName string `json:"long_name"`
	Img      string `json:"img"`
}

type Thumbnail struct {
	Src      string `json:"src"`
	Original string `json:"original"`
	Logo     bool   `json:"logo"`
}

type MainType string

const (
	Web MainType = "web"
)

type Language string

const (
	En Language = "en"
)

type Scheme string

const (
	HTTPS Scheme = "https"
)

type ResultType string

const (
	TypeSearchResult ResultType = "search_result"
)
