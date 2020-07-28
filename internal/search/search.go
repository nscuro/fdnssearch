package search

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nscuro/fdnssearch/internal/dataset"
	"github.com/valyala/fastjson"
)

func filter(chunk string, types *[]string, domains *[]string, exclusions *[]string, jsonParser *fastjson.Parser) (*dataset.Entry, error) {
	possibleMatch := false
	for _, domain := range *domains {
		if strings.Contains(chunk, domain) {
			possibleMatch = true
			break
		}
	}
	if !possibleMatch {
		return nil, nil
	}

	parsedEntry, err := jsonParser.Parse(chunk)
	if err != nil {
		return nil, fmt.Errorf("failed to parse entry: %w", err)
	}
	entryName := string(parsedEntry.GetStringBytes("name"))
	entryValue := string(parsedEntry.GetStringBytes("value"))
	entryType := string(parsedEntry.GetStringBytes("type"))

	// filter by type
	if len(*types) > 0 {
		found := false
		for _, ttype := range *types {
			if entryType == ttype {
				found = true
				break
			}
		}
		if !found {
			return nil, nil
		}
	}

	// filter by domain
	if len(*domains) > 0 {
		found := false
		for _, domain := range *domains {
			if entryName == domain || strings.HasSuffix(entryName, "."+domain) {
				found = true
				break
			}
		}
		if !found {
			return nil, nil
		}
	}

	// filter by exclusion
	if len(*exclusions) > 0 {
		found := false
		for _, exclusion := range *exclusions {
			if entryName == exclusion || strings.HasSuffix(entryName, "."+exclusion) {
				found = true
				break
			}
		}
		if found {
			return nil, nil
		}
	}

	return &dataset.Entry{
		Name:  entryName,
		Type:  entryType,
		Value: entryValue,
	}, nil
}

type Options struct {
	DatasetReader io.Reader
	Domains       []string
	Exclusions    []string
	Types         []string
}

type Searcher struct {
	jsonParser fastjson.Parser
}

func NewSearcher() *Searcher {
	return &Searcher{}
}

func (s Searcher) Search(ctx context.Context, options Options) (<-chan dataset.Entry, <-chan error, error) {
	if err := s.validateOptions(&options); err != nil {
		return nil, nil, fmt.Errorf("invalid options: %w", err)
	}

	resultsChan := make(chan dataset.Entry, 100)
	errorsChan := make(chan error)

	go func() {
		defer close(resultsChan)
		defer close(errorsChan)

		scanner := bufio.NewScanner(options.DatasetReader)
	scanLoop:
		for scanner.Scan() {
			// handle cancellation via context
			select {
			case <-ctx.Done():
				break scanLoop
			default:
			}

			entry, err := filter(scanner.Text(), &options.Types, &options.Domains, &options.Exclusions, &s.jsonParser)
			if err != nil {
				errorsChan <- err
				continue
			} else if entry == nil {
				continue
			}
			resultsChan <- *entry
		}
	}()

	return resultsChan, errorsChan, nil
}

func (s Searcher) validateOptions(options *Options) error {
	if options.DatasetReader == nil {
		return fmt.Errorf("no dataset reader provided")
	}
	if options.Domains == nil || len(options.Domains) == 0 {
		return fmt.Errorf("no domain filter provided")
	}
	if options.Types == nil || len(options.Types) == 0 {
		return fmt.Errorf("no type filter provided")
	}
	return nil
}
