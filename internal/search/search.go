package search

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/nscuro/fdnssearch/internal/dataset"
	"github.com/panjf2000/ants"
	"io"
	"strings"
)

type searchWorkerContext struct {
	chunk       string
	domains     *[]string
	types       *[]string
	resultsChan chan<- dataset.Entry
	errorsChan  chan<- error
}

func searchWorker(workerCtx interface{}) {
	ctx, ok := workerCtx.(searchWorkerContext)
	if !ok {
		return
	}

	var entry dataset.Entry
	if err := json.Unmarshal([]byte(ctx.chunk), &entry); err != nil {
		ctx.errorsChan <- err
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

	ctx.resultsChan <- entry
}

type Options struct {
	DatasetReader io.Reader
	Domains       []string
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

func (s Searcher) Search(options Options) (<-chan dataset.Entry, <-chan error, error) {
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

		scanner := bufio.NewScanner(options.DatasetReader)
		for scanner.Scan() {
			err = workerPool.Invoke(searchWorkerContext{
				chunk:       scanner.Text(),
				domains:     &options.Domains,
				types:       &options.Types,
				resultsChan: resultsChan,
				errorsChan:  errorsChan,
			})
			if err != nil {
				errorsChan <- fmt.Errorf("failed to submit chunk to worker pool: %w", err)
				break
			}
		}

		// wait for workers to finish
		for {
			if workerPool.Free() == workerPool.Cap() {
				break
			}
		}

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
