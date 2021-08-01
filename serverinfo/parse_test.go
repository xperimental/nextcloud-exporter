package serverinfo

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	inputFiles := []string{
		"negative-space.xml",
		"na-values.xml",
	}

	for _, inputFile := range inputFiles {
		inputFile := inputFile
		t.Run(inputFile, func(t *testing.T) {
			t.Parallel()

			reader, err := os.Open("testdata/" + inputFile)
			if err != nil {
				t.Fatalf("error opening test data: %s", err)
			}

			if _, err := Parse(reader); err != nil {
				t.Errorf("got error %q", err)
			}
		})
	}
}

func TestParseJSON(t *testing.T) {
	inputFiles := []string{
		"info.json",
	}

	for _, inputFile := range inputFiles {
		inputFile := inputFile
		t.Run(inputFile, func(t *testing.T) {
			t.Parallel()

			reader, err := os.Open("testdata/" + inputFile)
			if err != nil {
				t.Fatalf("error opening test data: %s", err)
			}

			if _, err := ParseJSON(reader); err != nil {
				t.Errorf("got error %q", err)
			}
		})
	}
}
