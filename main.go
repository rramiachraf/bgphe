package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func main() {
	query := flag.String("q", "", "BGP query")
	output := flag.String("o", "", "File output")
	flag.Parse()

	if *query == "" {
		fmt.Println("[ERR] query argument wasn't passed")
		return
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	url := fmt.Sprintf("https://bgp.he.net/search?search%%5Bsearch%%5D=%s&commit=Search", *query)

	var page string
	err := chromedp.Run(
		ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(".w100p > tbody"),
		chromedp.InnerHTML("html", &page),
	)

	if err != nil {
		fmt.Println("[ERR]" + err.Error())
	}

	document, _ := goquery.NewDocumentFromReader(strings.NewReader(page))

	var f *os.File
	defer f.Close()

	if *output != "" {
		f, err = os.Create(*output)

		if err != nil {
			fmt.Println("[ERR]", err.Error())
		}
	}

	document.Find(".w100p > tbody").Children().Each(func(i int, s *goquery.Selection) {
		el := s.Children()

		ASN := el.Find("a").Text()
		Org := el.Last().Text()
		Location, _ := el.Find("img").Attr("alt")

		str := fmt.Sprintf("[%s] %s [%s]\n", ASN, Org, Location)

		if i != 0 {
			fmt.Print(str)
			f.Write([]byte(str))
		}
	})
}
