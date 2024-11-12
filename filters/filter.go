package filters

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/PuerkitoBio/goquery"
)

var filters = map[string]func(doc *goquery.Document) (string, error){
	"https://developer.mozilla.org/en-US/docs/Web/": MdnFilter,
	"https://en.wikipedia.org/wiki/":                WikipediaFilter,
	"https://simple.wikipedia.org/wiki/":            WikipediaFilter,
}

func Filter(url, html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}
	doc.Find("script").Remove()
	doc.Find("style").Remove()

	for filterURL, filterFunc := range filters {
		if strings.HasPrefix(url, filterURL) {
			fmt.Println("Filtering with the", filterURL, "filter")
			return filterFunc(doc)
		}
	}

	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("204"))
	finalDoc, err := doc.Html()
	if err != nil {
		return "", err
	}
	fmt.Println(style.Render("No filter found for url:", url))
	return finalDoc, nil
}
