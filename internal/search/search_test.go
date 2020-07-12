package search

import (
	"sync"
	"testing"

	"github.com/nscuro/fdnssearch/internal/dataset"
	"github.com/stretchr/testify/assert"
)

func TestSearchWorker(t *testing.T) {
	resultsChan := make(chan dataset.Entry, 1)
	errorsChan := make(chan error, 1)
	defer close(resultsChan)
	defer close(errorsChan)

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	go searchWorker(searchWorkerContext{
		chunk:       "{\"name\":\"acme.example.com\",\"value\":\"1.1.1.1\",\"type\":\"a\"}",
		domains:     &[]string{"example.com"},
		exclusions:  &[]string{},
		types:       &[]string{"a"},
		resultsChan: resultsChan,
		errorsChan:  errorsChan,
		waitGroup:   &waitGroup,
	})

	waitGroup.Wait()

	assert.Len(t, resultsChan, 1)
	assert.Len(t, errorsChan, 0)

	result := <-resultsChan
	assert.Equal(t, "acme.example.com", result.Name)
	assert.Equal(t, "1.1.1.1", result.Value)
	assert.Equal(t, "a", result.Type)
}

func TestSearchWorkerFilterByDomain(t *testing.T) {
	resultsChan := make(chan dataset.Entry, 1)
	errorsChan := make(chan error, 1)
	defer close(resultsChan)
	defer close(errorsChan)

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	go searchWorker(searchWorkerContext{
		chunk:       "{\"name\":\"acme.example.com\",\"value\":\"1.1.1.1\",\"type\":\"a\"}",
		domains:     &[]string{"example.de"},
		exclusions:  &[]string{},
		types:       &[]string{},
		resultsChan: resultsChan,
		errorsChan:  errorsChan,
		waitGroup:   &waitGroup,
	})

	waitGroup.Wait()

	assert.Len(t, resultsChan, 0)
	assert.Len(t, errorsChan, 0)
}

func TestSearchWorkerFilterByType(t *testing.T) {
	resultsChan := make(chan dataset.Entry, 1)
	errorsChan := make(chan error, 1)
	defer close(resultsChan)
	defer close(errorsChan)

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	go searchWorker(searchWorkerContext{
		chunk:       "{\"name\":\"acme.example.com\",\"value\":\"1.1.1.1\",\"type\":\"a\"}",
		domains:     &[]string{"example.com"},
		exclusions:  &[]string{},
		types:       &[]string{"aaaa"},
		resultsChan: resultsChan,
		errorsChan:  errorsChan,
		waitGroup:   &waitGroup,
	})

	waitGroup.Wait()

	assert.Len(t, resultsChan, 0)
	assert.Len(t, errorsChan, 0)
}

func TestSearchWorkerFilterByExclusion(t *testing.T) {
	resultsChan := make(chan dataset.Entry, 1)
	errorsChan := make(chan error, 1)
	defer close(resultsChan)
	defer close(errorsChan)

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	go searchWorker(searchWorkerContext{
		chunk:       "{\"name\":\"acme.example.com\",\"value\":\"1.1.1.1\",\"type\":\"a\"}",
		domains:     &[]string{"example.com"},
		exclusions:  &[]string{"acme.example.com"},
		types:       &[]string{},
		resultsChan: resultsChan,
		errorsChan:  errorsChan,
		waitGroup:   &waitGroup,
	})

	waitGroup.Wait()

	assert.Len(t, resultsChan, 0)
	assert.Len(t, errorsChan, 0)
}

func TestSearchWorkerError(t *testing.T) {
	resultsChan := make(chan dataset.Entry, 1)
	errorsChan := make(chan error, 1)
	defer close(resultsChan)
	defer close(errorsChan)

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	go searchWorker(searchWorkerContext{
		chunk:       "invalidjson",
		domains:     &[]string{"example.com"},
		exclusions:  &[]string{},
		types:       &[]string{},
		resultsChan: resultsChan,
		errorsChan:  errorsChan,
		waitGroup:   &waitGroup,
	})

	waitGroup.Wait()

	assert.Len(t, resultsChan, 0)
	assert.Len(t, errorsChan, 1)
}
