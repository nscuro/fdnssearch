package search

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/nscuro/fdnssearch/internal/dataset"
	"github.com/panjf2000/ants"
)

type searchWorkerContext struct {
	chunk       string
	domains     *[]string
	exclusions  *[]string
	types       *[]string
	resultsChan chan<- dataset.Entry
	errorsChan  chan<- error
	waitGroup   *sync.WaitGroup
}

func searchWorker(workerCtx interface{}) {
	ctx, ok := workerCtx.(searchWorkerContext)
	if !ok {
		return
	}

	ctx.waitGroup.Add(1)
	defer ctx.waitGroup.Done()

	var entry dataset.Entry
	if err := json.Unmarshal([]byte(ctx.chunk), &entry); err != nil {
		ctx.errorsChan <- fmt.Errorf("failed to decode entry: %w", err)
		return
	}

	// filter by type
	if len(*ctx.types) > 0 {
		found := false
		for _, ttype := range *ctx.types {
			if entry.Type == ttype {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	// filter by domain
	if len(*ctx.domains) > 0 {
		found := false
		for _, domain := range *ctx.domains {
			if entry.Name == domain || strings.HasSuffix(entry.Name, "."+domain) {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	// filter by exclusion
	if len(*ctx.exclusions) > 0 {
		found := false
		for _, exclusion := range *ctx.exclusions {
			if entry.Name == exclusion || strings.HasSuffix(entry.Name, "."+exclusion) {
				found = true
				break
			}
		}
		if found {
			return
		}
	}

	ctx.resultsChan <- entry
}

type Options struct {
	DatasetReader io.Reader
	Domains       []string
	Exclusions    []string
	Types         []string
}

type Searcher struct {
	workerCount int
}

func NewSearcher(workerCount int) *Searcher {
	return &Searcher{
		workerCount: workerCount,
	}
}

func (s Searcher) Search(ctx context.Context, options Options) (<-chan dataset.Entry, <-chan error, error) {
	if err := s.validateOptions(&options); err != nil {
		return nil, nil, fmt.Errorf("invalid options: %w", err)
	}

	resultsChan := make(chan dataset.Entry, s.workerCount)
	errorsChan := make(chan error, s.workerCount)

	go func() {
		defer close(resultsChan)
		defer close(errorsChan)

		workerPool, err := ants.NewPoolWithFunc(s.workerCount, searchWorker)
		if err != nil {
			errorsChan <- err
			return
		}

		// wait group for search workers
		waitGroup := sync.WaitGroup{}

		scanner := bufio.NewScanner(options.DatasetReader)
	scanLoop:
		for scanner.Scan() {
			// handle cancellation via context
			select {
			case <-ctx.Done():
				break scanLoop
			default:
			}

			err = workerPool.Invoke(searchWorkerContext{
				chunk:       scanner.Text(),
				domains:     &options.Domains,
				exclusions:  &options.Exclusions,
				types:       &options.Types,
				resultsChan: resultsChan,
				errorsChan:  errorsChan,
				waitGroup:   &waitGroup,
			})
			if err != nil {
				errorsChan <- fmt.Errorf("failed to submit chunk to worker pool: %w", err)
				break
			}
		}

		waitGroup.Wait()
		workerPool.Release()
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
