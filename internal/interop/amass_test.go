package interop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAmassConfig(t *testing.T) {
	domains, err := ParseAmassConfig("./testdata/amass.ini")
	assert.NoError(t, err)
	assert.Len(t, domains, 3)
	assert.Contains(t, domains, "example.com")
	assert.Contains(t, domains, "example.de")
	assert.Contains(t, domains, "example.fr")
}
