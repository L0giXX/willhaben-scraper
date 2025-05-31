package main

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"

	"strings"
)

type WillhabenScraper struct {
	Collector *colly.Collector
	Listings  []*Listing
}

type Listing struct {
	Title       string
	Price       string
	Location    string
	Description string
	Seller      string
	URL         string
	PostedDate  string
}

func NewWillhabenScraper() *WillhabenScraper {
	c := colly.NewCollector(
		colly.AllowedDomains("willhaben.at", "www.willhaben.at"),
	)

	return &WillhabenScraper{
		Collector: c,
		Listings:  []*Listing{},
	}
}

func (s *WillhabenScraper) Scrape() {
	page := "1"

	s.Collector.OnHTML("span[data-testid^='ad-detail-ad-edit-date-top']", func(e *colly.HTMLElement) {
		currentURL := e.Request.URL.String()
		if !strings.Contains(currentURL, "/iad/immobilien/d/") {
			return
		}
		postedDate := strings.TrimSpace(e.Text)
		postedDate = strings.TrimLeft(postedDate, "Zuletzt geändert:")
		postedDate = strings.TrimSpace(postedDate)

		for i, listing := range s.Listings {
			if listing.URL == currentURL {
				s.Listings[i].PostedDate = postedDate
				break
			}
		}
	})

	s.Collector.OnHTML("a[id^='search-result-entry-header-']", func(e *colly.HTMLElement) {
		listing := &Listing{}

		link := e.Attr("href")
		if link != "" {
			listing.URL = "https://www.willhaben.at" + link
		}

		titleElement := e.DOM.Find("h3")
		if titleElement.Length() > 0 {
			listing.Title = titleElement.Text()
		}

		priceElement := e.DOM.Find("span[data-testid^='search-result-entry-price-']")
		if priceElement.Length() > 0 {
			listing.Price = strings.TrimSpace(priceElement.Text())
		}

		locationElement := e.DOM.Find("span[aria-label^='Ort']")
		if locationElement.Length() > 0 {
			listing.Location = strings.TrimSpace(locationElement.Text())
		}

		sellerElement := e.DOM.Find("span[data-testid^='search-result-entry-seller-information-']")
		if sellerElement.Length() > 0 {
			listing.Seller = strings.TrimSpace(sellerElement.Text())
		}

		// Extract attributes (like room count, size, etc.)
		attributeElements := e.DOM.Find("div[data-testid^='search-result-entry-teaser-attributes-'] div")
		var attributes []string
		attributeElements.Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				attributes = append(attributes, text)
			}
		})
		if len(attributes) > 0 {
			attributesText := strings.Join(attributes, " | ")
			if listing.Description == "" {
				listing.Description = attributesText
			} else {
				listing.Description += " | " + attributesText
			}
		}

		dateElement := e.DOM.Find("span[data-testid^='ad-detail-ad-edit-date-top']")
		if dateElement.Length() > 0 {
			listing.PostedDate = strings.TrimSpace(dateElement.Text())
		}

		if listing.URL != "" || listing.Title != "" {
			s.Listings = append(s.Listings, listing)
		}
	})

	s.Collector.Visit("https://www.willhaben.at/iad/immobilien/mietwohnungen/mietwohnung-angebote?sfId=b31ce01d-432e-46ea-9b79-45f94596adc1&isNavigation=true&page=" + page + "&rows=30")

	s.Collector.Wait()

	for _, listing := range s.Listings {
		s.Collector.Visit(listing.URL)
	}

	s.Collector.Wait()
}

func (s *WillhabenScraper) PrintListings() {
	for _, listing := range s.Listings {
		fmt.Printf("Title: %s\nPrice: %s\nLocation: %s\nSeller: %s\nURL: %s\nDescription: %s\nDate: %s\n\n",
			listing.Title, listing.Price, listing.Location, listing.Seller, listing.URL, listing.Description, listing.PostedDate)
	}
}

func main() {
	fmt.Println("Starting scraping...")

	scraper := NewWillhabenScraper()
	scraper.Scrape()
	scraper.PrintListings()

	fmt.Println("Scraping completed.")
}
