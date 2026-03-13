// Package flatten provides JSON flattening into a key=value text format.
package flatten

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// JSON takes a JSON byte slice and returns a flattened string representation
// where nested keys are separated by dots and array indices are shown in brackets.
//
// Example input:
//
//	{"data": {"volume": 124, "error": false}, "name": "device-1", "versions": ["a","b","c"]}
//
// Example output:
//
//	data.error=false
//	data.volume=124
//	name=device-1
//	versions[0]=a
//	versions[1]=b
//	versions[2]=c
func JSON(data []byte) (string, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", fmt.Errorf("parsing JSON: %w", err)
	}

	result := make(map[string]string)
	flattenValue("", obj, result)

	// Sort keys for deterministic output
	keys := make([]string, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(result[k])
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func flattenValue(prefix string, value interface{}, result map[string]string) {
	switch v := value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			newPrefix := key
			if prefix != "" {
				newPrefix = prefix + "." + key
			}
			flattenValue(newPrefix, val, result)
		}
	case []interface{}:
		for i, val := range v {
			newPrefix := fmt.Sprintf("%s[%d]", prefix, i)
			flattenValue(newPrefix, val, result)
		}
	case float64:
		// Format integers without decimal point
		if v == float64(int64(v)) {
			result[prefix] = fmt.Sprintf("%d", int64(v))
		} else {
			result[prefix] = fmt.Sprintf("%g", v)
		}
	case bool:
		result[prefix] = fmt.Sprintf("%t", v)
	case string:
		result[prefix] = v
	case nil:
		result[prefix] = "null"
	default:
		result[prefix] = fmt.Sprintf("%v", v)
	}
}
