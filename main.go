package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

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
	skipPreview := flag.Bool("s", false, "Skip the preview of the html file")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}
	if err := run(*filename, os.Stdout, *skipPreview); err != nil {
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
func run(filename string, out io.Writer, skipPreview bool) error {
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
	outFileName := tempFile.Name()
	fmt.Fprint(out, outFileName)
	if err := saveHTML(outFileName, htmlData); err != nil {
		return err
	}

	if skipPreview {
		return nil
	}
	// this is going to introduce a race condition
	// where the browser might not get the chance to open the file
	// before the preview returns
	// so we will add a delay in the preview function to give the browser time to open the file
	defer os.Remove(outFileName)
	return preview(outFileName)
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

func preview(fname string) error {
	cName := ""
	cParams := []string{}
	// Define executable based on OS
	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "cmd.exe"
		cParams = []string{"/C", "start"}
	case "darwin":
		cName = "open"
	default:
		return fmt.Errorf("OS not supported")
	}
	// Append filename to parameters slice
	cParams = append(cParams, fname)
	// Locate executable in PATH
	cPath, err := exec.LookPath(cName)

	if err != nil {
		return err
	}
	// Open the file using default program
	err = exec.Command(cPath, cParams...).Run()
	// TODO this is not the recommended solution
	// use signals to cleanup files
	time.Sleep(2 * time.Second)

	return err
}
