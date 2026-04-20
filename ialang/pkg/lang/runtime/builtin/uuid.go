package builtin

import (
	"crypto/rand"
	"fmt"
	goRegexp "regexp"
)

var uuidV4Pattern = goRegexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func newUUIDModule() Object {
	v4Fn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("uuid.v4 expects 0 args, got %d", len(args))
		}

		var b [16]byte
		if _, err := rand.Read(b[:]); err != nil {
			return nil, err
		}
		// RFC 4122 variant/version bits for v4 UUID.
		b[6] = (b[6] & 0x0f) | 0x40
		b[8] = (b[8] & 0x3f) | 0x80

		id := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]),
			uint16(b[4])<<8|uint16(b[5]),
			uint16(b[6])<<8|uint16(b[7]),
			uint16(b[8])<<8|uint16(b[9]),
			uint64(b[10])<<40|uint64(b[11])<<32|uint64(b[12])<<24|uint64(b[13])<<16|uint64(b[14])<<8|uint64(b[15]),
		)
		return id, nil
	})

	isValidFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("uuid.isValid expects 1 arg: text")
		}
		text, err := asStringArg("uuid.isValid", args, 0)
		if err != nil {
			return nil, err
		}
		return uuidV4Pattern.MatchString(text), nil
	})

	namespace := Object{
		"v4":      v4Fn,
		"isValid": isValidFn,
	}
	module := cloneObject(namespace)
	module["uuid"] = namespace
	return module
}
