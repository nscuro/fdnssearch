package search

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/nscuro/fdnssearch/internal/dataset"
	"github.com/valyala/fastjson"
)

type searchWorkerContext struct {
	chunk          string
	domains        *[]string
	exclusions     *[]string
	types          *[]string
	jsonParserPool *fastjson.ParserPool
	resultsChan    chan<- dataset.Entry
	errorsChan     chan<- error
	waitGroup      *sync.WaitGroup
}

func searchWorker(ctx searchWorkerContext) {
	defer ctx.waitGroup.Done()

	entry, err := filter(ctx.chunk, ctx.types, ctx.domains, ctx.exclusions, ctx.jsonParserPool)
	if err != nil {
		ctx.errorsChan <- err
		return
	} else if entry == nil {
		return
	}

	ctx.resultsChan <- *entry
}

func filter(chunk string, types *[]string, domains *[]string, exclusions *[]string, jsonParserPool *fastjson.ParserPool) (*dataset.Entry, error) {
	// prevent the necessity to decode entries that definitely
	// do not match the given search criteria. decoding json is
	// drastically more computationally expensive than this simple
	// loop.
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

	jsonParser := jsonParserPool.Get()
	parsedEntry, err := jsonParser.Parse(chunk)
	if err != nil {
		jsonParserPool.Put(jsonParser)
		return nil, fmt.Errorf("failed to parse entry: %w", err)
	}

	// parse everything we need in advance so jsonParser can
	// be put back into the pool as fast as possible
	entryName := string(parsedEntry.GetStringBytes("name"))
	entryValue := string(parsedEntry.GetStringBytes("value"))
	entryType := string(parsedEntry.GetStringBytes("type"))
	jsonParserPool.Put(jsonParser)

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
	jsonParserPool fastjson.ParserPool
}

func NewSearcher() *Searcher {
	return &Searcher{}
}

func (s Searcher) Search(ctx context.Context, options Options) (<-chan dataset.Entry, <-chan error, error) {
	if err := s.validateOptions(&options); err != nil {
		return nil, nil, fmt.Errorf("invalid options: %w", err)
	}

	resultsChan := make(chan dataset.Entry, 10)
	errorsChan := make(chan error)

	go func() {
		defer close(resultsChan)
		defer close(errorsChan)

		// wait group for search workers
		waitGroup := sync.WaitGroup{}

		// pool for fastjson.Parser to encourage reusing
		// of instances without causing race conditions
		jsonParserPool := fastjson.ParserPool{}

		scanner := bufio.NewScanner(options.DatasetReader)
	scanLoop:
		for scanner.Scan() {
			// handle cancellation via context
			select {
			case <-ctx.Done():
				break scanLoop
			default:
			}

			waitGroup.Add(1)
			go searchWorker(searchWorkerContext{
				chunk:          scanner.Text(),
				domains:        &options.Domains,
				exclusions:     &options.Exclusions,
				types:          &options.Types,
				jsonParserPool: &jsonParserPool,
				resultsChan:    resultsChan,
				errorsChan:     errorsChan,
				waitGroup:      &waitGroup,
			})
		}

		waitGroup.Wait()
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
