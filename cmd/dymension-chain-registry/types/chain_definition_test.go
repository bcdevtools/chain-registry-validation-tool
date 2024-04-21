package types

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_GetXUrls(t *testing.T) {
	tests := []struct {
		name       string
		urlsGetter func(ChainDefinition) ([]string, error)
	}{
		{
			name: "rpc",
			urlsGetter: func(cd ChainDefinition) ([]string, error) {
				return cd.GetRpcUrls()
			},
		},
		{
			name: "rest",
			urlsGetter: func(cd ChainDefinition) ([]string, error) {
				return cd.GetRestUrls()
			},
		},
	}
	for _, tt := range tests {
		name := tt.name
		f := tt.urlsGetter
		t.Run("string "+name, func(t *testing.T) {
			jsonContent := `{"` + name + `": "123"}`

			var chainDefinition ChainDefinition
			err := json.Unmarshal([]byte(jsonContent), &chainDefinition)
			require.NoError(t, err)

			urls, err := f(chainDefinition)
			require.NoError(t, err)
			require.Equal(t, []string{"123"}, urls)
		})

		t.Run("multiple "+name, func(t *testing.T) {
			jsonContent := `{"` + name + `": ["123", "456"]}`

			var chainDefinition ChainDefinition
			err := json.Unmarshal([]byte(jsonContent), &chainDefinition)
			require.NoError(t, err)

			urls, err := f(chainDefinition)
			require.NoError(t, err)
			require.Equal(t, []string{"123", "456"}, urls)
		})

		t.Run("no "+name, func(t *testing.T) {
			jsonContent := `{}`

			var chainDefinition ChainDefinition
			err := json.Unmarshal([]byte(jsonContent), &chainDefinition)
			require.NoError(t, err)

			urls, err := f(chainDefinition)
			require.NoError(t, err)
			require.Empty(t, urls)
		})

		t.Run("null "+name, func(t *testing.T) {
			jsonContent := `{"` + name + `": null}`

			var chainDefinition ChainDefinition
			err := json.Unmarshal([]byte(jsonContent), &chainDefinition)
			require.NoError(t, err)

			urls, err := f(chainDefinition)
			require.NoError(t, err)
			require.Empty(t, urls)
		})
	}
}
