package api

import hostfs "iacommon/pkg/host/fs"

type FSOpenRequest struct {
	Path string
	Opts hostfs.OpenOptions
}

type FSOpenResponse struct {
	Handle uint64
}

type FSReadRequest struct {
	Handle uint64
	Size   int64
}

type FSReadResponse struct {
	Data []byte
	N    int64
	EOF  bool
}

type FSWriteRequest struct {
	Handle uint64
	Data   []byte
}

type FSWriteResponse struct {
	N int64
}

type FSSeekRequest struct {
	Handle uint64
	Offset int64
	Whence int64
}

type FSSeekResponse struct {
	Offset int64
}

type FSCloseRequest struct {
	Handle uint64
}

func decodeFSOpenRequest(args map[string]any) (FSOpenRequest, error) {
	path, err := readString(args, "path")
	if err != nil {
		return FSOpenRequest{}, err
	}
	return FSOpenRequest{
		Path: path,
		Opts: hostfs.OpenOptions{
			Read:   readBool(args, "read"),
			Write:  readBool(args, "write"),
			Create: readBool(args, "create"),
			Trunc:  readBool(args, "trunc"),
			Append: readBool(args, "append"),
		},
	}, nil
}

func encodeFSOpenResponse(handle uint64) CallResult {
	return CallResult{Value: map[string]any{"handle": handle}}
}

func decodeFSReadRequest(args map[string]any) (FSReadRequest, error) {
	handle, err := readRequiredUint64(args, "handle")
	if err != nil {
		return FSReadRequest{}, err
	}
	size, err := readOptionalInt64Any(args, "size")
	if err != nil {
		return FSReadRequest{}, err
	}
	if size <= 0 {
		size = 4096
	}
	return FSReadRequest{Handle: handle, Size: size}, nil
}

func encodeFSReadResponse(data []byte, n int64, eof bool) CallResult {
	return CallResult{Value: map[string]any{
		"data": data,
		"n":    n,
		"eof":  eof,
	}}
}

func decodeFSWriteRequest(args map[string]any) (FSWriteRequest, error) {
	handle, err := readRequiredUint64(args, "handle")
	if err != nil {
		return FSWriteRequest{}, err
	}
	data, err := readBytes(args, "data")
	if err != nil {
		return FSWriteRequest{}, err
	}
	return FSWriteRequest{Handle: handle, Data: data}, nil
}

func encodeFSWriteResponse(n int64) CallResult {
	return CallResult{Value: map[string]any{"n": n}}
}

func decodeFSSeekRequest(args map[string]any) (FSSeekRequest, error) {
	handle, err := readRequiredUint64(args, "handle")
	if err != nil {
		return FSSeekRequest{}, err
	}
	offset, err := readOptionalInt64Any(args, "offset")
	if err != nil {
		return FSSeekRequest{}, err
	}
	whence, err := readOptionalInt64Any(args, "whence")
	if err != nil {
		return FSSeekRequest{}, err
	}
	return FSSeekRequest{Handle: handle, Offset: offset, Whence: whence}, nil
}

func encodeFSSeekResponse(offset int64) CallResult {
	return CallResult{Value: map[string]any{"offset": offset}}
}

func decodeFSCloseRequest(args map[string]any) (FSCloseRequest, error) {
	handle, err := readRequiredUint64(args, "handle")
	if err != nil {
		return FSCloseRequest{}, err
	}
	return FSCloseRequest{Handle: handle}, nil
}
