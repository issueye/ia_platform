package builtin

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

func newHMACModule() Object {
	sha1Fn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("hmac.sha1 expects 2 args: key, data")
		}
		key, err := asStringValue("hmac.sha1 arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		data, err := asStringValue("hmac.sha1 arg[1]", args[1])
		if err != nil {
			return nil, err
		}
		mac := hmac.New(sha1.New, []byte(key))
		_, _ = mac.Write([]byte(data))
		return hex.EncodeToString(mac.Sum(nil)), nil
	})

	sha256Fn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("hmac.sha256 expects 2 args: key, data")
		}
		key, err := asStringValue("hmac.sha256 arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		data, err := asStringValue("hmac.sha256 arg[1]", args[1])
		if err != nil {
			return nil, err
		}
		mac := hmac.New(sha256.New, []byte(key))
		_, _ = mac.Write([]byte(data))
		return hex.EncodeToString(mac.Sum(nil)), nil
	})

	sha512Fn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("hmac.sha512 expects 2 args: key, data")
		}
		key, err := asStringValue("hmac.sha512 arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		data, err := asStringValue("hmac.sha512 arg[1]", args[1])
		if err != nil {
			return nil, err
		}
		mac := hmac.New(sha512.New, []byte(key))
		_, _ = mac.Write([]byte(data))
		return hex.EncodeToString(mac.Sum(nil)), nil
	})

	verifySha256Fn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("hmac.verifySha256 expects 3 args: key, data, signatureHex")
		}
		key, err := asStringValue("hmac.verifySha256 arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		data, err := asStringValue("hmac.verifySha256 arg[1]", args[1])
		if err != nil {
			return nil, err
		}
		sigHex, err := asStringValue("hmac.verifySha256 arg[2]", args[2])
		if err != nil {
			return nil, err
		}
		expected, err := hex.DecodeString(sigHex)
		if err != nil {
			return false, nil
		}
		mac := hmac.New(sha256.New, []byte(key))
		_, _ = mac.Write([]byte(data))
		actual := mac.Sum(nil)
		return subtle.ConstantTimeCompare(actual, expected) == 1, nil
	})

	namespace := Object{
		"sha1":         sha1Fn,
		"sha256":       sha256Fn,
		"sha512":       sha512Fn,
		"verifySha256": verifySha256Fn,
	}
	module := cloneObject(namespace)
	module["hmac"] = namespace
	return module
}
