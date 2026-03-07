package gitscope_test

import (
	"encoding/json"
	"os"
	"testing"
)

func TestScoopManifestIsValidJSON(t *testing.T) {
	data, err := os.ReadFile("git-scope.json")
	if err != nil {
		t.Fatalf("failed to read git-scope.json: %v", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("git-scope.json is not valid JSON: %v", err)
	}

	requiredFields := []string{"version", "description", "homepage", "license", "architecture", "bin", "checkver", "autoupdate"}
	for _, field := range requiredFields {
		if _, ok := manifest[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}

	arch, ok := manifest["architecture"].(map[string]interface{})
	if !ok {
		t.Fatal("architecture field is not an object")
	}
	for _, key := range []string{"64bit", "arm64"} {
		entry, ok := arch[key].(map[string]interface{})
		if !ok {
			t.Errorf("architecture.%s is not an object", key)
			continue
		}
		if _, ok := entry["url"]; !ok {
			t.Errorf("architecture.%s missing url", key)
		}
	}
}
