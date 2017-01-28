// All code in retriever should permit getting the raw content of a document
package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"os"
)

// Retrivier is an interface that can get content of a document by it's index or id
type Retriever interface {
	// getDoc returns the content of a document
	retrieve(string, int) (string, error)
	// Serialize, save to disk the content of the struct if needed
	Serialize(string)
}

type cacmRetriever struct {
	// array that point to the document index in cacm.all
	Ids []int64
}

func (r *cacmRetriever) retrieve(title string, id int) (string, error) {
	if id > len(r.Ids)-1 {
		return "", errors.New("Undefined document")
	}
	file, err := os.Open(cacmFile)
	if err != nil {
		return "", err
	}
	defer file.Close()
	file.Seek(r.Ids[id], 0)

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

func (r *cacmRetriever) Serialize(name string) {
	file, err := os.Create("indexes/" + name + ".retriever")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	en := gob.NewEncoder(file)
	err = en.Encode(r.Ids)
	if err != nil {
		panic(err)
	}
	file.Sync()
}

func UnserializeCacmRetriever(name string) *cacmRetriever {
	r := cacmRetriever{}
	file, err := os.Open("indexes/" + name + ".retriever")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	en := gob.NewDecoder(file)
	err = en.Decode(&r.Ids)
	if err != nil {
		panic(err)
	}
	return &r
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

// Dummy function as nothing is needed fot cs276Retriever
func (r *cs276Retriever) Serialize(name string) {}
