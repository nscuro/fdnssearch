package interop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAmassConfig(t *testing.T) {
	cfg, err := ParseAmassConfig("./testdata/amass.ini")
	assert.NoError(t, err)

	assert.Len(t, cfg.Domains, 3)
	assert.Contains(t, cfg.Domains, "example.com")
	assert.Contains(t, cfg.Domains, "example.de")
	assert.Contains(t, cfg.Domains, "example.fr")

	assert.Len(t, cfg.Blacklisted, 1)
	assert.Contains(t, cfg.Blacklisted, "acme.example.com")
}
