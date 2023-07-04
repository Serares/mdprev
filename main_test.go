package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

const (
	inputFile  = "./testdata/test1.md"
	goldenFile = "./testdata/test1.md.html"
)

func TestParseContent(t *testing.T) {
	input, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatal(err)
	}

	result := parseContent(input)

	expect, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, expect) {
		t.Logf("golden:\n%s\n", expect)
		t.Logf("result:\n%s\n", result)
		t.Error("result content does not match the expected file")
	}
}

func TestRun(t *testing.T) {
	var tempOutFile bytes.Buffer
	// important
	// we pass the bytes.Buffer as a refference because
	// the bytes.Buffer implements the io.Write interface using a pointer receiver
	if err := run(inputFile, &tempOutFile); err != nil {
		t.Fatal(err)
	}
	resultFile := strings.TrimSpace(tempOutFile.String())
	result, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatal(err)
	}
	expect, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, expect) {
		t.Logf("golden:\n%s\n", expect)
		t.Logf("result:\n%s\n", result)
		t.Error("result content does not match the expected file")
	}

	os.Remove(resultFile)
}
