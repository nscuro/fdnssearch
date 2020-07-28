package search

import (
	"testing"

	"github.com/valyala/fastjson"
)

func BenchmarkFilterForPositiveMatch(b *testing.B) {
	domains := []string{"example.com"}
	exclusions := []string{}
	types := []string{"a"}
	jsonParser := fastjson.Parser{}

	for i := 0; i < b.N; i++ {
		filter(`{"name":"acme.example.com","value":"1.1.1.1","type":"a"}`, &types, &domains, &exclusions, &jsonParser)
	}
}

func BenchmarkFilterForDefinitiveMismatch(b *testing.B) {
	domains := []string{"somethingelse.com"}
	exclusions := []string{}
	types := []string{}
	jsonParser := fastjson.Parser{}

	for i := 0; i < b.N; i++ {
		filter(`{"name":"acme.example.com","value":"1.1.1.1","type":"a"}`, &types, &domains, &exclusions, &jsonParser)
	}
}

func BenchmarkFilterForDomainMismatch(b *testing.B) {
	domains := []string{"example.com"}
	exclusions := []string{}
	types := []string{"cname"}
	jsonParser := fastjson.Parser{}

	for i := 0; i < b.N; i++ {
		filter(`{"name":"acme.com","value":"other.example.com","type":"cname"}`, &types, &domains, &exclusions, &jsonParser)
	}
}

func BenchmarkFilterForTypeMismatch(b *testing.B) {
	domains := []string{"example.com"}
	exclusions := []string{}
	types := []string{"cname"}
	jsonParser := fastjson.Parser{}

	for i := 0; i < b.N; i++ {
		filter(`{"name":"acme.example.com","value":"1.1.1.1","type":"a"}`, &types, &domains, &exclusions, &jsonParser)
	}
}

func BenchmarkFilterForExclusionMatch(b *testing.B) {
	domains := []string{"example.com"}
	exclusions := []string{"acme.example.com"}
	types := []string{"a"}
	jsonParser := fastjson.Parser{}

	for i := 0; i < b.N; i++ {
		filter(`{"name":"acme.example.com","value":"1.1.1.1","type":"a"}`, &types, &domains, &exclusions, &jsonParser)
	}
}
