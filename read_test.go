// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codesearch

import (
	"bytes"
	"log"
	"reflect"
	"testing"
)

var _ = log.Println

func TestReadWrite(t *testing.T) {
	b := NewIndexBuilder()
	b.AddFile("filename", []byte("abcde"))

	var buf bytes.Buffer
	b.Write(&buf)
	f := &memSeeker{buf.Bytes(), 0}

	r := reader{r: f}

	var toc indexTOC
	r.readTOC(&toc)

	if r.err != nil {
		t.Errorf("got read error %v", r.err)
	}
	if toc.fileContents.content.data.sz != 5 {
		t.Errorf("got contents size %d, want 5", toc.fileContents.content.data.sz)
	}

	data := r.readIndexData(&toc)
	if got := data.fileName(0); got != "filename" {
		t.Errorf("got filename %s, want %v", got, "filename")
	}

	if len(data.ngrams) != 3 {
		t.Fatalf("got ngrams %v, want 3 ngrams", data.ngrams)
	}

	if want := []uint32{5}; !reflect.DeepEqual(data.fileEnds, want) {
		t.Fatalf("got fileEnds %v, want %v", data.fileEnds, want)
	}

	if _, ok := data.ngrams[stringToNGram("bcq")]; ok {
		t.Errorf("found ngram bcd in %v", data.ngrams)
	}

	got := fromDeltas(r.readSectionBlob(data.ngrams[stringToNGram("bcd")]))
	if want := []uint32{1}; !reflect.DeepEqual(got, want) {
		t.Errorf("got posting data %v, want %v", got, want)
	}
}

func TestReadWriteNames(t *testing.T) {
	b := NewIndexBuilder()
	b.AddFile("abCd", []byte(""))

	var buf bytes.Buffer
	b.Write(&buf)
	f := &memSeeker{buf.Bytes(), 0}

	r := reader{r: f}

	var toc indexTOC
	r.readTOC(&toc)
	if r.err != nil {
		t.Errorf("got read error %v", r.err)
	}
	if toc.fileNames.content.data.sz != 4 {
		t.Errorf("got contents size %d, want 4", toc.fileNames.content.data.sz)
	}

	data := r.readIndexData(&toc)
	if !reflect.DeepEqual([]byte{0x4}, data.fileNameCaseBits) {
		t.Errorf("got case bits %v, want {0x4}", data.fileNameCaseBits)
	}
	if !reflect.DeepEqual([]uint32{0, 4}, data.fileNameIndex) {
		t.Errorf("got index %v, want {0,4}", data.fileNameIndex)
	}
	if got := data.fileNameNgrams[stringToNGram("bcd")]; !reflect.DeepEqual(got, []uint32{1}) {
		t.Errorf("got trigram bcd at bits %v, want sz 2", data.fileNameNgrams)
	}
}
