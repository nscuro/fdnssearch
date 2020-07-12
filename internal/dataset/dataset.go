package dataset

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var datasetRegex = regexp.MustCompile("^[\\d]{4}-[\\d]{2}-[\\d]{2}-[\\d]+-fdns_(?P<types>[\\w-]+)\\.json\\.gz$")

type Dataset struct {
	Fingerprint string
	Types       []string
	URL         string
}

type Entry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

func FetchDatasets(ctx context.Context) ([]Dataset, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://opendata.rapid7.com/sonar.fdns_v2/", nil)
	if err != nil {
		return nil, fmt.Errorf("preparing request failed: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching dataset page failed: %w", err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing dataset page failed: %w", err)
	}

	datasets := make([]Dataset, 0)
	doc.Find("tr.ungated").Each(func(_ int, selection *goquery.Selection) {
		fingerprint := strings.TrimSpace(selection.Find("div.fingerprint > div.sha").Text())

		url := selection.Find("td:nth-child(1) > a").AttrOr("href", "")
		if url[0] == '/' {
			url = "https://opendata.rapid7.com" + url
		}

		types, err := getDatasetTypes(strings.TrimSpace(selection.Find("td:nth-child(1) > a").Text()))
		if err != nil {
			return
		}

		datasets = append(datasets, Dataset{
			Fingerprint: fingerprint,
			Types:       types,
			URL:         url,
		})
	})

	return datasets, nil
}

func getDatasetTypes(datasetName string) ([]string, error) {
	matches := datasetRegex.FindStringSubmatch(datasetName)
	if matches == nil {
		return nil, fmt.Errorf("regex did not match")
	}
	if len(matches) != 2 {
		return nil, fmt.Errorf("unexpected match count")
	}

	return strings.Split(matches[1], "_"), nil
}
