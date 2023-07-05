package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const (
	defaultTemplate = `<!DOCTYPE html>
	<html>
	<head>
	<meta http-equiv="content-type" content="text/html; charset=utf-8">
	<title>{{ .Title }}</title>
	</head>
	<body>
			<h1>Preview of File: {{ .FileName }}</hi>
			<h3>Bellow is the content of the md file</h3>
			</br>
			{{ .Body }}
	</body>
	</html>
	`
)

type content struct {
	Title    string
	Body     template.HTML
	FileName string
}

func main() {
	filename := flag.String("file", "", "Name of the markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip the preview of the html file")
	tFname := flag.String("t", "", "Alternate template filename")
	flag.Parse()
	if err := run(*filename, os.Stdout, os.Stdin, *skipPreview, *tFname); err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(1)
	}
}

// the second argument is the interesting one
// we use the io.Writer interface
// because we are creating the out file as a temp one taht can be stored anywhere with a random name
// we have to capture it's name from the output of the function
// is the case of the programm the writer will be the STDOUT of the terminal so that the user can see the name of the outfile
// but in case of tests the outfile will be captured by a buffer slice see line 37 main_test.go
func run(filename string, out io.Writer, r io.Reader, skipPreview bool, templateFname string) error {
	input, err := getInput(filename, r)
	if err != nil {
		fmt.Printf("error extracting the input %v", err)
		return err
	}

	htmlData, err := parseContent(input, templateFname, filename)
	if err != nil {
		fmt.Printf("error parsing the content %v", err)
		return err
	}

	tempFile, err := os.CreateTemp("", "mdprev.*.html")
	if err != nil {
		fmt.Printf("error creating temp file %v", err)
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

func getInput(filename string, r io.Reader) ([]byte, error) {
	var input []byte
	if filename == "" {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Printf("Got an error scanning the stdin %v\n", scanner.Err().Error())
			}
			input = append(input, scanner.Bytes()...)
		}
	}
	if filename != "" {
		// Read all the data from the input file and check for errors
		extractedBytes, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		input = extractedBytes
	}

	return input, nil
}

func parseContent(input []byte, templateFname string, fileName string) ([]byte, error) {
	// Parse the markdown file through blackfriday and bluemonday
	// to generate a valid and safe HTML
	output := blackfriday.Run(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)

	t, err := template.New("mdp").Parse(defaultTemplate)
	if err != nil {
		return nil, err
	}

	if templateFname != "" {
		t, err = template.ParseFiles(templateFname)
		if err != nil {
			return nil, err
		}
	}

	if fileName == "" {
		fileName = "MD provided from the STDIN"
	}
	c := content{
		Title:    "Preview for the md file",
		Body:     template.HTML(body),
		FileName: fileName,
	}
	var buffContent bytes.Buffer
	if err := t.Execute(&buffContent, c); err != nil {
		return nil, err
	}
	return buffContent.Bytes(), nil
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
