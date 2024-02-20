package cache

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
)

var gzipEncodePool = &sync.Pool{
	New: func() interface{} {
		gw, err := gzip.NewWriterLevel(nil, defaultGzipContentEncodingLevel)
		if err != nil {
			panic(err)
		}
		return gw
	},
}

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

const (
	// defaultGzipContentEncodingLevel is set to 1 which uses least CPU compared to higher levels, yet offers
	// similar compression ratios (off by at most 1.5x, but typically within 1.1x-1.3x). For further details see -
	// https://github.com/kubernetes/kubernetes/issues/112296
	defaultGzipContentEncodingLevel = 1
	// defaultGzipThresholdBytes is compared to the size of the first write from the stream
	// (usually the entire object), and if the size is smaller no gzipping will be performed
	// if the client requests it.
	defaultGzipThresholdBytes = 128 * 1024
)

// gzipEncode encode data to gzip
func gzipEncode(in []byte) (out []byte, err error) {
	writer := gzipEncodePool.Get().(*gzip.Writer)
	defer func() {
		writer.Reset(nil)
		gzipEncodePool.Put(writer)
	}()
	buffer := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		bufferPool.Put(buffer)
	}()

	// reset writer
	writer.Reset(buffer)
	// write data
	_, err = writer.Write(in)
	if err != nil {
		return out, err
	}
	writer.Flush()

	err = writer.Close()
	if err != nil {
		return out, err
	}
	out = []byte(buffer.String())
	return out, nil
}

// gzipDecode decode gzip data
func gzipDecode(in []byte) (out []byte, err error) {
	reader, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return out, err
	}
	defer reader.Close()

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer buffer.Reset()
	defer bufferPool.Put(buffer)

	_, err = io.Copy(buffer, reader)
	if err != nil {
		return
	}

	out = []byte(buffer.String())
	return out, nil
}
