package builtin

import (
	"fmt"
	"net/http"
	"path/filepath"

	goMime "mime"
)

func newMIMEModule() Object {
	typeByExtFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("mime.typeByExt expects 1 arg, got %d", len(args))
		}
		ext, err := asStringArg("mime.typeByExt", args, 0)
		if err != nil {
			return nil, err
		}
		if ext != "" && ext[0] != '.' {
			ext = "." + ext
		}
		t := goMime.TypeByExtension(ext)
		if t == "" {
			return nil, nil
		}
		return t, nil
	})

	extByTypeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("mime.extByType expects 1 arg, got %d", len(args))
		}
		mimeType, err := asStringArg("mime.extByType", args, 0)
		if err != nil {
			return nil, err
		}
		exts, err := goMime.ExtensionsByType(mimeType)
		if err != nil {
			return nil, err
		}
		out := make(Array, 0, len(exts))
		for _, ext := range exts {
			out = append(out, ext)
		}
		return out, nil
	})

	detectTypeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("mime.detectType expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("mime.detectType", args, 0)
		if err != nil {
			return nil, err
		}
		return http.DetectContentType([]byte(text)), nil
	})

	detectByPathFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("mime.detectByPath expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("mime.detectByPath", args, 0)
		if err != nil {
			return nil, err
		}
		ext := filepath.Ext(path)
		if ext == "" {
			return nil, nil
		}
		t := goMime.TypeByExtension(ext)
		if t == "" {
			return nil, nil
		}
		return t, nil
	})

	namespace := Object{
		"typeByExt":    typeByExtFn,
		"extByType":    extByTypeFn,
		"detectType":   detectTypeFn,
		"detectByPath": detectByPathFn,
	}
	module := cloneObject(namespace)
	module["mime"] = namespace
	return module
}
