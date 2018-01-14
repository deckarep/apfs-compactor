/*
Open Source Initiative OSI - The MIT License (MIT):Licensing
The MIT License (MIT)
Copyright (c) 2018 Ralph Caraveo (deckarep@gmail.com)
Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"crypto/md5"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	mapset "github.com/deckarep/golang-set"
)

const (
	destFolder = "dest_files"
)

var (
	fileTypes      = mapset.NewSetFromSlice([]interface{}{"bin"})
	allFilesSeen   = mapset.NewSet()
	fileSizeCorpus = make(map[int64][]string)
	fileHashCorpus = make(map[string][]string)
)

func main() {
	flag.Parse()

	rootFolder := flag.Arg(0)
	// Walk the file-system for a given root folder.
	err := filepath.Walk(rootFolder, visit)
	if err != nil {
		log.Println("Walk failed with err: ", err)
	}

	// Identifies files first by byte size.
	// Then for matching files will hash them using sha1.
	fmt.Println("Identifying and compacting duplicates...")
	fmt.Println(strings.Repeat("-", 30))
	identifyDuplicates()

	// Anything in this list is a duplicate.
	for h, items := range fileHashCorpus {
		fmt.Printf("h: %s\n", h)
		for _, f := range items {
			fmt.Printf("\tRecreating file as clone: %s\n", f)
			err := copyFile(f, path.Join(destFolder, path.Base(f)))
			if err != nil {
				log.Fatalf("Couldn't copy file source: %s to dest: %s", f, path.Join(destFolder, path.Base(f)))
			}
		}
	}

	// Move over anything that wasn't a duplicate.
	fmt.Println("Moving over non-duplicates...")
	fmt.Println(strings.Repeat("-", 30))
	copyNonExistingFiles(rootFolder, destFolder)
}

func copyNonExistingFiles(srcFolder, dstFolder string) error {
	allFilesSeen.Each(func(item interface{}) bool {
		f := item.(string)
		baseFile := path.Base(f)
		destFile := path.Join(dstFolder, baseFile)
		if _, err := os.Stat(destFile); os.IsNotExist(err) {
			fmt.Println("Copying non-duplicate file", destFile)
			err := copyFile(f, destFile)
			if err != nil {
				log.Fatalf("Coudn't copy non-duplicate file: %s", f)
			}
		}
		return false
	})
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func identifyDuplicates() {
	for _, v := range fileSizeCorpus {
		if len(v) > 1 {
			for _, path := range v {
				h := hashFile(path)
				item, ok := fileHashCorpus[h]
				if ok {
					item = append(item, path)
					fileHashCorpus[h] = item
				} else {
					fileHashCorpus[h] = []string{path}
				}
			}
		}
	}
}

func visit(path string, f os.FileInfo, err error) error {
	if !f.IsDir() {
		pieces := strings.Split(path, ".")
		ext := pieces[len(pieces)-1]

		if fileTypes.Contains(ext) {
			allFilesSeen.Add(path)

			fi, err := os.Open(path)
			if err != nil {
				panic("Could open file!")
			}

			defer fi.Close()

			in, err := fi.Stat()
			if err != nil {
				panic("Could get file info!")
			}

			item, ok := fileSizeCorpus[in.Size()]
			if ok {
				item = append(item, path)
				fileSizeCorpus[in.Size()] = item
			} else {
				fileSizeCorpus[in.Size()] = []string{path}
			}
		}
	}
	return nil
}

func hashString(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func hashFile(path string) string {
	h := sha1.New()
	f, err := os.Open(path)
	if err != nil {
		log.Fatal("Couldn't open file: ", err)
	}
	defer f.Close()

	_, err = io.Copy(h, f)
	if err != nil {
		log.Fatal("Couldn't copy file bytes over to hash: ", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
