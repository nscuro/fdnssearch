package search

import (
	"fmt"
	"io"
)

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

func (s Searcher) Search(options Options) error {
	if err := s.validateOptions(&options); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	return nil
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
