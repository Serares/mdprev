package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const (
	header = `<!DOCTYPE html>
	<html>
	<head>
	<meta http-equiv="content-type" content="text/html; charset=utf-8">
	<title>Markdown Preview Tool</title>
	</head>
	<body>
	`
	footer = `
	</body>
	</html>
	`
)

func main() {
	filename := flag.String("file", "", "Name of the markdown file to preview")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}
	if err := run(*filename, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// the second argument is the interesting one
// we use the io.Writer interface
// because we are creating the out file as a temp one taht can be stored anywhere with a random name
// we have to capture it's name from the output of the function
// is the case of the programm the writer will be the STDOUT of the terminal so that the user can see the name of the outfile
// but in case of tests the outfile will be captured by a buffer slice see line 37 main_test.go
func run(filename string, out io.Writer) error {
	// Read all the data from the input file and check for errors
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	htmlData := parseContent(input)
	tempFile, err := os.CreateTemp("", "mdprev.*.html")
	if err != nil {
		return err
	}

	if err := tempFile.Close(); err != nil {
		return err
	}

	fmt.Fprint(out, tempFile.Name())
	return saveHTML(tempFile.Name(), htmlData)
}

func parseContent(input []byte) []byte {
	// Parse the markdown file through blackfriday and bluemonday
	// to generate a valid and safe HTML
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	var combinedHtmlBytes bytes.Buffer

	combinedHtmlBytes.WriteString(header)
	combinedHtmlBytes.Write(body)
	combinedHtmlBytes.WriteString(footer)

	return combinedHtmlBytes.Bytes()
}

func saveHTML(outFName string, data []byte) error {
	return os.WriteFile(outFName, data, 0644)
}
