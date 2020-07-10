package dataset

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchDatasets(t *testing.T) {
	datasets, err := FetchDatasets(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, datasets)
}

func TestGetDatasetTypes(t *testing.T) {
	types, err := getDatasetTypes("2020-06-02-1591078479-fdns_txt_mx_mta-sts.json.gz")
	assert.NoError(t, err)
	assert.Equal(t, types, []string{"txt", "mx", "mta-sts"})
}
