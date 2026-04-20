package builtin

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
)

func newCompressModule() Object {
	gzipCompressFn := NativeFunction(func(args []Value) (Value, error) {
		return compressStringToBase64("compress.gzipCompress", args,
			func(w io.Writer, level int) (io.WriteCloser, error) { return gzip.NewWriterLevel(w, level) })
	})
	gzipDecompressFn := NativeFunction(func(args []Value) (Value, error) {
		return decompressBase64ToString("compress.gzipDecompress", args,
			func(r io.Reader) (io.ReadCloser, error) { return gzip.NewReader(r) })
	})
	gzipCompressBytesFn := NativeFunction(func(args []Value) (Value, error) {
		return compressBytesToBase64("compress.gzipCompressBytes", args,
			func(w io.Writer, level int) (io.WriteCloser, error) { return gzip.NewWriterLevel(w, level) })
	})
	gzipDecompressBytesFn := NativeFunction(func(args []Value) (Value, error) {
		return decompressBase64ToBytes("compress.gzipDecompressBytes", args,
			func(r io.Reader) (io.ReadCloser, error) { return gzip.NewReader(r) })
	})

	zlibCompressFn := NativeFunction(func(args []Value) (Value, error) {
		return compressStringToBase64("compress.zlibCompress", args,
			func(w io.Writer, level int) (io.WriteCloser, error) { return zlib.NewWriterLevel(w, level) })
	})
	zlibDecompressFn := NativeFunction(func(args []Value) (Value, error) {
		return decompressBase64ToString("compress.zlibDecompress", args,
			func(r io.Reader) (io.ReadCloser, error) { return zlib.NewReader(r) })
	})
	zlibCompressBytesFn := NativeFunction(func(args []Value) (Value, error) {
		return compressBytesToBase64("compress.zlibCompressBytes", args,
			func(w io.Writer, level int) (io.WriteCloser, error) { return zlib.NewWriterLevel(w, level) })
	})
	zlibDecompressBytesFn := NativeFunction(func(args []Value) (Value, error) {
		return decompressBase64ToBytes("compress.zlibDecompressBytes", args,
			func(r io.Reader) (io.ReadCloser, error) { return zlib.NewReader(r) })
	})

	namespace := Object{
		"defaultCompression":  float64(gzip.DefaultCompression),
		"noCompression":       float64(gzip.NoCompression),
		"bestSpeed":           float64(gzip.BestSpeed),
		"bestCompression":     float64(gzip.BestCompression),
		"huffmanOnly":         float64(gzip.HuffmanOnly),
		"gzip":                gzipCompressFn,
		"gunzip":              gzipDecompressFn,
		"deflate":             zlibCompressFn,
		"inflate":             zlibDecompressFn,
		"gzipCompress":        gzipCompressFn,
		"gzipDecompress":      gzipDecompressFn,
		"gzipCompressBytes":   gzipCompressBytesFn,
		"gzipDecompressBytes": gzipDecompressBytesFn,
		"zlibCompress":        zlibCompressFn,
		"zlibDecompress":      zlibDecompressFn,
		"zlibCompressBytes":   zlibCompressBytesFn,
		"zlibDecompressBytes": zlibDecompressBytesFn,
	}
	module := cloneObject(namespace)
	module["compress"] = namespace
	return module
}

func compressStringToBase64(fn string, args []Value, newWriter func(io.Writer, int) (io.WriteCloser, error)) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("%s expects 1-2 args, got %d", fn, len(args))
	}
	text, err := asStringArg(fn, args, 0)
	if err != nil {
		return nil, err
	}
	level, err := compressionLevelArg(fn, args)
	if err != nil {
		return nil, err
	}
	return compressPayloadToBase64([]byte(text), level, newWriter)
}

func compressBytesToBase64(fn string, args []Value, newWriter func(io.Writer, int) (io.WriteCloser, error)) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("%s expects 1-2 args, got %d", fn, len(args))
	}
	buf, err := asByteArray(fn+" arg[0]", args[0])
	if err != nil {
		return nil, err
	}
	level, err := compressionLevelArg(fn, args)
	if err != nil {
		return nil, err
	}
	return compressPayloadToBase64(buf, level, newWriter)
}

func compressPayloadToBase64(payload []byte, level int, newWriter func(io.Writer, int) (io.WriteCloser, error)) (Value, error) {
	var b bytes.Buffer
	zw, err := newWriter(&b, level)
	if err != nil {
		return nil, err
	}
	if _, err := zw.Write(payload); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func decompressBase64ToString(fn string, args []Value, newReader func(io.Reader) (io.ReadCloser, error)) (Value, error) {
	out, err := decompressBase64Payload(fn, args, newReader)
	if err != nil {
		return nil, err
	}
	return string(out), nil
}

func decompressBase64ToBytes(fn string, args []Value, newReader func(io.Reader) (io.ReadCloser, error)) (Value, error) {
	out, err := decompressBase64Payload(fn, args, newReader)
	if err != nil {
		return nil, err
	}
	return byteSliceToArray(out), nil
}

func decompressBase64Payload(fn string, args []Value, newReader func(io.Reader) (io.ReadCloser, error)) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s expects 1 arg, got %d", fn, len(args))
	}
	encoded, err := asStringArg(fn, args, 0)
	if err != nil {
		return nil, err
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	zr, err := newReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}

func compressionLevelArg(fn string, args []Value) (int, error) {
	if len(args) == 1 {
		return gzip.DefaultCompression, nil
	}
	level, err := asIntArg(fn, args, 1)
	if err != nil {
		return 0, err
	}
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		return 0, fmt.Errorf("%s arg[1] expects compression level in [-2, 9], got %d", fn, level)
	}
	return level, nil
}
