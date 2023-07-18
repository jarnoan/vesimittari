package scraper

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jarnoan/vesimittari/meter"
)

var (
	ctrRegex  = regexp.MustCompile(`(\d{1,2}\.\d{1,2}\.\d{4})\s+(\d+)`)
	custRegex = regexp.MustCompile(`(\w*)\**`)
	datefmt   = "2.1.2006"
)

// Scraper can scrape consumption records from web.
type Scraper struct{}

// New constructs a new scraper.
func New() *Scraper {
	return &Scraper{}
}

// ReadConsumption reads the consumption data of a meter.
func (s *Scraper) ReadMeter(site meter.SiteNumber, num meter.Number) (meter.Reading, error) {
	ctx := context.Background()

	jar, err := cookiejar.New(nil)
	if err != nil {
		return meter.Reading{}, fmt.Errorf("create cookie jar: %w", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	// get login form page because it sets some cookies, I don't know whether they are required or not
	if err := s.getLoginPage(ctx, client); err != nil {
		return meter.Reading{}, fmt.Errorf("get login page: %w", err)
	}

	frontPageDoc, err := s.postLoginForm(ctx, client, site, num)
	if err != nil {
		return meter.Reading{}, fmt.Errorf("get login page: %w", err)
	}

	cntrPageURL, err := s.counterPageURL(frontPageDoc)
	if err != nil {
		return meter.Reading{}, fmt.Errorf("counter page url: %w", err)
	}

	cntrPageDoc, err := s.getCounterPage(ctx, client, cntrPageURL)
	if err != nil {
		return meter.Reading{}, fmt.Errorf("get counter page: %w", err)
	}

	consData, err := s.consumptionData(cntrPageDoc)
	if err != nil {
		return meter.Reading{}, err
	}

	return consData, nil
}

func (s *Scraper) getLoginPage(ctx context.Context, client *http.Client) error {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://www.kulutus-web.com/Nokia/vesi/Suomi/",
		nil,
	)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-ok response: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func (s *Scraper) postLoginForm(ctx context.Context, client *http.Client, site meter.SiteNumber, num meter.Number) (*goquery.Document, error) {
	data := url.Values{
		"mittarinro":         {string(num)},
		"kpiste":             {string(site)},
		"laitosid":           {"4"}, // TODO: read these from login page
		"toimialaid":         {"2"},
		"MenuToTheLeftFrame": {"no"},
		"kieli":              {"suomi"},
	}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://www.kulutus-web.com/common/logincheck_old.asp",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-ok response: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("new document: %w", err)
	}

	return doc, nil
}

func (s *Scraper) counterPageURL(doc *goquery.Document) (string, error) {
	ctrLink := doc.Find("#menuItem2 a")
	if ctrLink.Length() != 1 {
		return "", fmt.Errorf("found %d links", ctrLink.Length())
	}
	ctrURL, ok := ctrLink.First().Attr("href")
	if !ok {
		return "", fmt.Errorf("no href in link")
	}

	return "https://www.kulutus-web.com/common/" + ctrURL, nil
}

func (s *Scraper) getCounterPage(ctx context.Context, client *http.Client, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-ok response: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("new document: %w", err)
	}

	return doc, nil
}

func (s *Scraper) consumptionData(doc *goquery.Document) (meter.Reading, error) {
	td := doc.Find(`form[name="ilmoituslomake"] > table > tbody > tr:nth-child(4) > td:nth-child(2)`)
	if td.Length() != 1 {
		return meter.Reading{}, fmt.Errorf("found %d counter tds", td.Length())
	}

	tdText := td.First().Text()
	ms := ctrRegex.FindStringSubmatch(tdText)
	if len(ms) != 3 {
		return meter.Reading{}, fmt.Errorf("found %d matches", len(ms))
	}

	date, err := time.Parse(datefmt, ms[1])
	if err != nil {
		return meter.Reading{}, fmt.Errorf("invalid date: %s", ms[1])
	}

	ctr, err := strconv.Atoi(ms[2])
	if err != nil {
		return meter.Reading{}, fmt.Errorf("invalid counter value: %s", ms[2])
	}

	custDiv := doc.Find("#asiakasContent")
	if custDiv.Length() != 1 {
		return meter.Reading{}, fmt.Errorf("found %d customer divs", td.Length())
	}

	custText := custDiv.First().Text()
	ms = custRegex.FindStringSubmatch(custText)
	if len(ms) != 2 {
		return meter.Reading{}, fmt.Errorf("found %d customer names in %s", len(ms), custText)
	}

	return meter.Reading{
		Counter:  ctr,
		Date:     date,
		Customer: ms[1],
	}, nil
}
