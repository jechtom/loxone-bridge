package flatten

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON_SimpleObject(t *testing.T) {
	input := `{"name": "device-1", "value": 42}`
	result, err := JSON([]byte(input))
	require.NoError(t, err)
	assert.Contains(t, result, "name=device-1\n")
	assert.Contains(t, result, "value=42\n")
}

func TestJSON_NestedObject(t *testing.T) {
	input := `{"data": {"volume": 124, "error": false}}`
	result, err := JSON([]byte(input))
	require.NoError(t, err)
	assert.Contains(t, result, "data.volume=124\n")
	assert.Contains(t, result, "data.error=false\n")
}

func TestJSON_Array(t *testing.T) {
	input := `{"versions": ["a", "b", "c"]}`
	result, err := JSON([]byte(input))
	require.NoError(t, err)
	assert.Contains(t, result, "versions[0]=a\n")
	assert.Contains(t, result, "versions[1]=b\n")
	assert.Contains(t, result, "versions[2]=c\n")
}

func TestJSON_ComplexExample(t *testing.T) {
	input := `{
		"data": {
			"volume": 124,
			"error": false
		},
		"name": "device-1",
		"versions": ["a", "b", "c"]
	}`
	result, err := JSON([]byte(input))
	require.NoError(t, err)

	expected := `data.error=false
data.volume=124
name=device-1
versions[0]=a
versions[1]=b
versions[2]=c
`
	assert.Equal(t, expected, result)
}

func TestJSON_NullValue(t *testing.T) {
	input := `{"key": null}`
	result, err := JSON([]byte(input))
	require.NoError(t, err)
	assert.Contains(t, result, "key=null\n")
}

func TestJSON_FloatValue(t *testing.T) {
	input := `{"temp": 23.5}`
	result, err := JSON([]byte(input))
	require.NoError(t, err)
	assert.Contains(t, result, "temp=23.5\n")
}

func TestJSON_EmptyObject(t *testing.T) {
	input := `{}`
	result, err := JSON([]byte(input))
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestJSON_DeeplyNested(t *testing.T) {
	input := `{"a": {"b": {"c": "deep"}}}`
	result, err := JSON([]byte(input))
	require.NoError(t, err)
	assert.Contains(t, result, "a.b.c=deep\n")
}

func TestJSON_InvalidJSON(t *testing.T) {
	_, err := JSON([]byte("not json"))
	assert.Error(t, err)
}

func TestJSON_NestedArray(t *testing.T) {
	input := `{"matrix": [[1, 2], [3, 4]]}`
	result, err := JSON([]byte(input))
	require.NoError(t, err)
	assert.Contains(t, result, "matrix[0][0]=1\n")
	assert.Contains(t, result, "matrix[0][1]=2\n")
	assert.Contains(t, result, "matrix[1][0]=3\n")
	assert.Contains(t, result, "matrix[1][1]=4\n")
}

func TestJSON_BooleanValues(t *testing.T) {
	input := `{"enabled": true, "disabled": false}`
	result, err := JSON([]byte(input))
	require.NoError(t, err)
	assert.Contains(t, result, "enabled=true\n")
	assert.Contains(t, result, "disabled=false\n")
}
