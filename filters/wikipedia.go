package filters

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

func WikipediaFilter(doc *goquery.Document) (string, error) {
	mainContent := doc.Find("#bodyContent").First()
	if mainContent.Length() == 0 {
		fmt.Errorf("Main content not found, falling back to body")
		mainContent = doc.Find("body").First()
	}

	mainContent.Find("a").Each(func(i int, a *goquery.Selection) {
		a.ReplaceWithHtml(a.Text())
	})

	return mainContent.Html()
}
