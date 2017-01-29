// all code here only serve to serve content from cacm to "give sens" to the qquery
package main

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
)

// cacmDoc is content of a document from cacm in a vaguely structured form
type cacmDoc struct {
	B string
	T string
	W string
	A string
	K string
}

// getCACMDoc returns a cacmDoc from parsing the cacm.all file
func getCACMDoc(index int) (cacmDoc, error) {
	index++
	file, err := os.Open(cacmFile)
	if err != nil {
		return cacmDoc{}, err
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			return cacmDoc{}, nil
		}
		if line[:3] != ".I " {
			continue
		}

		i, err := strconv.Atoi(line[3 : len(line)-1])
		if err != nil {
			return cacmDoc{}, err
		}
		// We just found the document
		if i == index {
			return ParseCACMDoc(buf)
		}
	}
}

// parseCACMDoc parses the content of one document in cacm.all
func ParseCACMDoc(buf *bufio.Reader) (cacmDoc, error) {
	state := id
	doc := cacmDoc{}
	var tmp bytes.Buffer
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			// end of file, we are returning the last doc
			return doc, nil
		}
		if line[0] == '.' {
			// add to previous state
			switch state {
			case title:
				doc.T = tmp.String()
			case summary:
				doc.W = tmp.String()
			case keyWords:
				doc.K = tmp.String()
			case publication:
				doc.B = tmp.String()
			case authors:
				doc.A = tmp.String()
			}
			// get new state
			state = identToField(line[:2])
			tmp.Reset()
			if state == id {
				return doc, nil
			}
		} else {
			tmp.WriteString(line)
		}
	}
}
