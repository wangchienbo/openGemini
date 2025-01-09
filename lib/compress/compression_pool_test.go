// Copyright 2024 openGemini Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
)

func TestGzipWriterPool(t *testing.T) {
	var buf bytes.Buffer
	writer := GetGzipWriter(&buf)
	_, err := writer.Write([]byte("test data"))
	assert.NoError(t, err)
	PutGzipWriter(writer)

	reader, err := gzip.NewReader(&buf)
	assert.NoError(t, err)
	defer reader.Close()

	result := new(bytes.Buffer)
	_, err = io.Copy(result, reader)
	assert.NoError(t, err)
	assert.Equal(t, "test data", result.String())
}

func TestZstdWriterPool(t *testing.T) {
	var buf bytes.Buffer
	writer := GetZstdWriter(&buf)
	_, err := writer.Write([]byte("test data"))
	assert.NoError(t, err)
	PutZstdWriter(writer)

	reader, err := zstd.NewReader(&buf)
	assert.NoError(t, err)
	defer reader.Close()

	result := new(bytes.Buffer)
	_, err = io.Copy(result, reader)
	assert.NoError(t, err)
	assert.Equal(t, "test data", result.String())
}

func TestSnappyWriterPool(t *testing.T) {
	var buf bytes.Buffer
	writer := GetSnappyWriter(&buf)
	_, err := writer.Write([]byte("test data"))
	assert.NoError(t, err)
	PutSnappyWriter(writer)

	reader := snappy.NewReader(&buf)
	result := new(bytes.Buffer)
	_, err = io.Copy(result, reader)
	assert.NoError(t, err)
	assert.Equal(t, "test data", result.String())
}

func TestSnappyReaderPool(t *testing.T) {
	// Write data using Snappy writer
	var buf bytes.Buffer
	writer := snappy.NewBufferedWriter(&buf)
	_, err := writer.Write([]byte("test data"))
	assert.NoError(t, err)
	writer.Close()

	// Get Snappy reader from pool and read data
	reader := GetSnappyReader(&buf)
	result := new(bytes.Buffer)
	_, err = io.Copy(result, reader)
	assert.NoError(t, err)
	assert.Equal(t, "test data", result.String())

	// Put Snappy reader back to pool
	PutSnappyReader(reader)
}

func TestZstdReaderPool(t *testing.T) {
	// Write data using Zstd writer
	var buf bytes.Buffer
	writer, err := zstd.NewWriter(&buf)
	assert.NoError(t, err)
	_, err = writer.Write([]byte("test data"))
	assert.NoError(t, err)
	writer.Close()

	// Get Zstd reader from pool and read data
	reader := GetZstdReader(&buf)
	result := new(bytes.Buffer)
	_, err = io.Copy(result, reader)
	assert.NoError(t, err)
	assert.Equal(t, "test data", result.String())

	// Put Zstd reader back to pool
	PutZstdReader(reader)
}
