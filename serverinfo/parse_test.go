package serverinfo

import (
	"os"
	"testing"
)

func TestParseServerInfo(t *testing.T) {
	inputFiles := []string{
		"negative-space.xml",
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
