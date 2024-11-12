package filters

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

func MdnFilter(doc *goquery.Document) (string, error) {
	mainContent := doc.Find("#content > article").First()

	if mainContent.Length() == 0 {
		return "", fmt.Errorf("main content not found")
	}

	content, err := mainContent.Html()
	if err != nil {
		return "", err
	}

	return content, nil
}
