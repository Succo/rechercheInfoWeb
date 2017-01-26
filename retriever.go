// All code in retriever should permit getting the raw content of a document
package main

import (
	"bufio"
	"bytes"
	"errors"
	"os"
)

type retriever interface {
	// getDoc returns the content of a document
	retrieve(string, int) (string, error)
}

type cacmRetriever struct {
	// array that point to the document index in cacm.all
	idMap []int64
}

func (r *cacmRetriever) retrieve(title string, id int) (string, error) {
	if id > len(r.idMap)-1 {
		return "", errors.New("Undefined document")
	}
	file, err := os.Open(cacmFile)
	if err != nil {
		return "", err
	}
	defer file.Close()
	file.Seek(r.idMap[id], 0)

	var buf bytes.Buffer
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := buf.Bytes()
		if line[0] == '.' {
			if line[1] == 'I' {
				break
			}
			continue
		}

		buf.Write(line)
	}
	return buf.String(), nil
}

type cs276Retriever struct {
}

func (r *cs276Retriever) retrieve(title string, id int) (string, error) {
	file, err := os.Open(cs276File + "/" + title)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// write file content to a buffer
	var buf bytes.Buffer
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		buf.Write(scanner.Bytes())
	}
	// return a string with the content
	return buf.String(), nil
}
