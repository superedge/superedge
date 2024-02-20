package cache

import "testing"

func BenchmarkGzip(t *testing.B) {
	data := "gzip input data test test tttttttt"

	// test encode
	encodeData, err := gzipEncode([]byte(data))
	if err != nil {
		t.Fatalf("encode data to gzip format failed, err: %v", err)
	}
	t.Logf("gzip data: %v", string(encodeData))

	// test decode
	decodeData, err := gzipDecode(encodeData)
	if err != nil {
		t.Fatalf("decode gzip data failed, err: %v", err)
	}

	if data != string(decodeData) {
		t.Fatalf("invalid decode data, expect: %s, decode: %s", data, string(decodeData))
	}

	t.Logf("decode data: %s", string(decodeData))
}
