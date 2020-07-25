package search

import (
	"sync"
	"testing"

	"github.com/nscuro/fdnssearch/internal/dataset"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fastjson"
)

func TestSearchWorker(t *testing.T) {
	resultsChan := make(chan dataset.Entry, 1)
	errorsChan := make(chan error, 1)
	defer close(resultsChan)
	defer close(errorsChan)

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	go searchWorker(searchWorkerContext{
		chunk:          "{\"name\":\"acme.example.com\",\"value\":\"1.1.1.1\",\"type\":\"a\"}",
		domains:        &[]string{"example.com"},
		exclusions:     &[]string{},
		types:          &[]string{"a"},
		jsonParserPool: &fastjson.ParserPool{},
		resultsChan:    resultsChan,
		errorsChan:     errorsChan,
		waitGroup:      &waitGroup,
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
		chunk:          "{\"name\":\"acme.example.com\",\"value\":\"1.1.1.1\",\"type\":\"a\"}",
		domains:        &[]string{"example.de"},
		exclusions:     &[]string{},
		types:          &[]string{},
		jsonParserPool: &fastjson.ParserPool{},
		resultsChan:    resultsChan,
		errorsChan:     errorsChan,
		waitGroup:      &waitGroup,
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
		chunk:          "{\"name\":\"acme.example.com\",\"value\":\"1.1.1.1\",\"type\":\"a\"}",
		domains:        &[]string{"example.com"},
		exclusions:     &[]string{},
		types:          &[]string{"aaaa"},
		jsonParserPool: &fastjson.ParserPool{},
		resultsChan:    resultsChan,
		errorsChan:     errorsChan,
		waitGroup:      &waitGroup,
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
		chunk:          "{\"name\":\"acme.example.com\",\"value\":\"1.1.1.1\",\"type\":\"a\"}",
		domains:        &[]string{"example.com"},
		exclusions:     &[]string{"acme.example.com"},
		types:          &[]string{},
		jsonParserPool: &fastjson.ParserPool{},
		resultsChan:    resultsChan,
		errorsChan:     errorsChan,
		waitGroup:      &waitGroup,
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
		chunk:          "invalidJsonThatContainsexample.com",
		domains:        &[]string{"example.com"},
		exclusions:     &[]string{},
		types:          &[]string{},
		jsonParserPool: &fastjson.ParserPool{},
		resultsChan:    resultsChan,
		errorsChan:     errorsChan,
		waitGroup:      &waitGroup,
	})

	waitGroup.Wait()

	assert.Len(t, resultsChan, 0)
	assert.Len(t, errorsChan, 1)
}

func BenchmarkFilterForPositiveMatch(b *testing.B) {
	domains := []string{"example.com"}
	exclusions := []string{}
	types := []string{"a"}
	jsonParserPool := fastjson.ParserPool{}

	for i := 0; i < b.N; i++ {
		filter(`{"name":"acme.example.com","value":"1.1.1.1","type":"a"}`, &types, &domains, &exclusions, &jsonParserPool)
	}
}

func BenchmarkFilterForDefinitiveMismatch(b *testing.B) {
	domains := []string{"somethingelse.com"}
	exclusions := []string{}
	types := []string{}
	jsonParserPool := fastjson.ParserPool{}

	for i := 0; i < b.N; i++ {
		filter(`{"name":"acme.example.com","value":"1.1.1.1","type":"a"}`, &types, &domains, &exclusions, &jsonParserPool)
	}
}

func BenchmarkFilterForDomainMismatch(b *testing.B) {
	domains := []string{"example.com"}
	exclusions := []string{}
	types := []string{"cname"}
	jsonParserPool := fastjson.ParserPool{}

	for i := 0; i < b.N; i++ {
		filter(`{"name":"acme.com","value":"other.example.com","type":"cname"}`, &types, &domains, &exclusions, &jsonParserPool)
	}
}

func BenchmarkFilterForTypeMismatch(b *testing.B) {
	domains := []string{"example.com"}
	exclusions := []string{}
	types := []string{"cname"}
	jsonParserPool := fastjson.ParserPool{}

	for i := 0; i < b.N; i++ {
		filter(`{"name":"acme.example.com","value":"1.1.1.1","type":"a"}`, &types, &domains, &exclusions, &jsonParserPool)
	}
}

func BenchmarkFilterForExclusionMatch(b *testing.B) {
	domains := []string{"example.com"}
	exclusions := []string{"acme.example.com"}
	types := []string{"a"}
	jsonParserPool := fastjson.ParserPool{}

	for i := 0; i < b.N; i++ {
		filter(`{"name":"acme.example.com","value":"1.1.1.1","type":"a"}`, &types, &domains, &exclusions, &jsonParserPool)
	}
}
