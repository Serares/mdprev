package main

import (
	"bytes"
	"os"
	"testing"
)

const (
	inputFile  = "./testdata/test1.md"
	resultFile = "test1.md.html"
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
	if err := run(inputFile); err != nil {
		t.Fatal(err)
	}

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
