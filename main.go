package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/pterm/pterm"
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
		return
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

	table := pterm.TableData{{"ASN/IP", "Prefixes", "Organization", "Location"}}

	document.Find(".w100p > tbody").Children().Each(func(i int, s *goquery.Selection) {
		el := s.Children()

		ASN := el.Find("a").Text()
		Org := el.Last().Text()
		Location, _ := el.Find("img").Attr("alt")
		var Prefixes string

		params, urlExists := s.Find("a").Attr("href")
		url = fmt.Sprintf("https://bgp.he.net%s", params)
		exp := `Prefixes Originated \(all\): (\d+)`
		var body string

		if strings.HasPrefix(ASN, "AS") && urlExists {
			chromedp.Run(
				ctx,
				chromedp.Navigate(url),
				chromedp.WaitReady("body"),
				chromedp.InnerHTML("body", &body),
			)

			rgx, _ := regexp.Compile(exp)
			matches := rgx.FindStringSubmatch(body)
			Prefixes = string(matches[1])
		}

		str := fmt.Sprintf("%s (%s) | %s [%s]\n", ASN, Prefixes, Org, Location)

		if Prefixes == "" {
			str = fmt.Sprintf("%s | %s [%s]\n", ASN, Org, Location)
		}

		if i != 0 {
			table = append(table, []string{ASN, Prefixes, Org, Location})
			f.Write([]byte(str))
		}
	})

	pterm.DefaultTable.WithHasHeader().WithData(table).Render()
}
