package main

import (
	"bufio"
	"fmt"
	"github.com/antchfx/xmlquery"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var validWriter *bufio.Writer
var invalidWriter *bufio.Writer
var deadWriter *bufio.Writer
var protocol = regexp.MustCompile("^http")

func main() {

	validFile, _ := os.Create("valid-urls.tsv")
	defer validFile.Close()
	validWriter = bufio.NewWriter(validFile)

	invalidFile, _ := os.Create("invalid-urls.tsv")
	defer invalidFile.Close()
	invalidWriter = bufio.NewWriter(invalidFile)

	deadLinkFile, _ := os.Create("dead-links.tsv")
	defer deadLinkFile.Close()
	deadWriter = bufio.NewWriter(deadLinkFile)

	eadRoot := "/home/menneric/work/dlts-finding-aids-ead-sample-set-3/ead-files"

	err := filepath.Walk(eadRoot, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() != true {
			getURLs(path, info.Name())
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}


func getURLs(path string, filename string) {

	bytes, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}
	fileString := (string(bytes))
	doc, err := xmlquery.Parse(strings.NewReader(fileString))
	if err != nil {
		panic(err)
	}

	for _, i := range xmlquery.Find(doc, "//@xlink:href") {
		link := strings.TrimSpace(i.InnerText())
		log.Println("- INFO - testing", link)

		if protocol.MatchString(link) == false {
			invalidWriter.WriteString(fmt.Sprintf("%s\t%s\n", filename, link))
			invalidWriter.Flush()
			continue
		}

		_, err := url.Parse(link)
		if err != nil {
			invalidWriter.WriteString(fmt.Sprintf("%s\t%s\n", filename, link))
			invalidWriter.Flush()
			continue
		}


		response, err := http.Get(link)
		if err != nil {
			deadWriter.WriteString(fmt.Sprintf("%s\t%s\n", filename, link))
			deadWriter.Flush()
			continue
		}

		if response.StatusCode != 200 {
			deadWriter.WriteString(fmt.Sprintf("%s\t%d\t%s\n",filename, response.StatusCode, link))
			deadWriter.Flush()
			continue
		}


		validWriter.WriteString(fmt.Sprintf("%s\t%s\n", filename, link))
		validWriter.Flush()

	}
}