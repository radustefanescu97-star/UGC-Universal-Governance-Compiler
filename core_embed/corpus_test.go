package core_embed

import (
	"io/fs"
	"strings"
	"testing"
)

func TestStandardCorpusIsUGCNative(t *testing.T) {
	forbidden := []string{
		"PromSpace",
		"Tycho",
		"Firebase",
		"Cloud Run",
		"Nominatim",
		"Stripe",
		"Apple",
		"GCP",
	}
	allowed := map[string]map[string]string{
		"standard_corpus/SOPs/UGC_REPOSITORY_EXPLAINABILITY_SOP.md": {
			"firebase": "`firebase.json` is listed as a generic deployment-config filename example, not as project-specific guidance.",
		},
	}

	err := fs.WalkDir(StandardCorpus, "standard_corpus", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, err := fs.ReadFile(StandardCorpus, path)
		if err != nil {
			return err
		}
		content := strings.ToLower(string(data))
		for _, term := range forbidden {
			term = strings.ToLower(term)
			if strings.Contains(content, term) && !allowedCorpusOccurrence(allowed, path, term) {
				t.Errorf("%s contains project-specific term %q", path, term)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk standard corpus: %v", err)
	}
}

func allowedCorpusOccurrence(allowed map[string]map[string]string, path, term string) bool {
	terms, ok := allowed[path]
	if !ok {
		return false
	}
	_, ok = terms[term]
	return ok
}
