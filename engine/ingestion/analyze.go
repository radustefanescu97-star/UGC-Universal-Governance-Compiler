package ingestion

import (
	"fmt"
	"os"
)

func AnalyzeProject() error {
	fmt.Println("Analyzing project framework indicators...")
	indicators := []string{
		"package.json",
		"go.mod",
		"requirements.txt",
		"pyproject.toml",
	}

	found := []string{}
	for _, ind := range indicators {
		if _, err := os.Stat(ind); err == nil {
			found = append(found, ind)
		}
	}

	if len(found) > 0 {
		fmt.Printf("Found framework indicators: %v\n", found)
	} else {
		fmt.Println("No common framework indicators found.")
	}

	return nil
}
