package builtin

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"hash/fnv"
)

func computeHash(algo string, text string) (Value, error) {
	switch algo {
	case "sha1":
		sum := sha1.Sum([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	case "sha256":
		sum := sha256.Sum256([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	case "sha512":
		sum := sha512.Sum512([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	case "md5":
		sum := md5.Sum([]byte(text))
		return hex.EncodeToString(sum[:]), nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", algo)
	}
}

func newHashModule() Object {
	namespace := Object{
		"sha1":   newHashFn("hash.sha1", "sha1"),
		"sha256": newHashFn("hash.sha256", "sha256"),
		"sha512": newHashFn("hash.sha512", "sha512"),
		"crc32": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("hash.crc32 expects 1 arg, got %d", len(args))
			}
			text, err := asStringArg("hash.crc32", args, 0)
			if err != nil {
				return nil, err
			}
			return float64(crc32.ChecksumIEEE([]byte(text))), nil
		}),
		"fnv32a": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("hash.fnv32a expects 1 arg, got %d", len(args))
			}
			text, err := asStringArg("hash.fnv32a", args, 0)
			if err != nil {
				return nil, err
			}
			h := fnv.New32a()
			_, _ = h.Write([]byte(text))
			return float64(h.Sum32()), nil
		}),
		"fnv64a": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("hash.fnv64a expects 1 arg, got %d", len(args))
			}
			text, err := asStringArg("hash.fnv64a", args, 0)
			if err != nil {
				return nil, err
			}
			h := fnv.New64a()
			_, _ = h.Write([]byte(text))
			return float64(h.Sum64()), nil
		}),
	}
	module := cloneObject(namespace)
	module["hash"] = namespace
	return module
}
