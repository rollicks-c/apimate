package client

import (
	"embed"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:embed res/fixture.json
var resources embed.FS

func TestJsonArrayParsing(t *testing.T) {

	rawList, err := resources.ReadFile("res/fixture.json")
	assert.NoError(t, err)

	rawArray, err := ParseArrayList(rawList)
	assert.NoError(t, err)

	type item struct {
		K1 string `json:"k1"`
		K2 string `json:"k2"`
	}
	var data []item
	err = json.Unmarshal(rawArray, &data)
	assert.NoError(t, err)

	assert.Equal(t, 4, len(data))
	assert.Equal(t, "v1", data[0].K1)
	assert.Equal(t, "v2", data[0].K2)
	assert.Equal(t, "v3", data[1].K1)
	assert.Equal(t, "v4", data[1].K2)
	assert.Equal(t, "v5", data[2].K1)
	assert.Equal(t, "v6", data[2].K2)
	assert.Equal(t, "v7", data[3].K1)
	assert.Equal(t, "v8", data[3].K2)

}
