package api

import hostfs "iacommon/pkg/host/fs"

type FSReadFileRequest struct {
	Path string
}

type FSReadFileResponse struct {
	Data []byte
}

type FSWriteFileRequest struct {
	Path string
	Data []byte
	Opts hostfs.WriteOptions
}

type FSAppendFileRequest struct {
	Path string
	Data []byte
}

type FSReadDirRequest struct {
	Path string
}

type FSReadDirResponse struct {
	Entries []hostfs.DirEntry
}

type FSStatRequest struct {
	Path string
}

type FSStatResponse struct {
	Info hostfs.FileInfo
}

type FSMkdirRequest struct {
	Path string
	Opts hostfs.MkdirOptions
}

type FSRemoveRequest struct {
	Path string
	Opts hostfs.RemoveOptions
}

type FSRenameRequest struct {
	OldPath string
	NewPath string
}

func decodeFSReadFileRequest(args map[string]any) (FSReadFileRequest, error) {
	path, err := readString(args, "path")
	if err != nil {
		return FSReadFileRequest{}, err
	}
	return FSReadFileRequest{Path: path}, nil
}

func encodeFSReadFileResponse(data []byte) CallResult {
	return CallResult{Value: map[string]any{"data": data}}
}

func decodeFSWriteFileRequest(args map[string]any) (FSWriteFileRequest, error) {
	path, err := readString(args, "path")
	if err != nil {
		return FSWriteFileRequest{}, err
	}
	data, err := readBytes(args, "data")
	if err != nil {
		return FSWriteFileRequest{}, err
	}
	return FSWriteFileRequest{
		Path: path,
		Data: data,
		Opts: hostfs.WriteOptions{
			Create: readBool(args, "create"),
			Trunc:  readBool(args, "trunc"),
		},
	}, nil
}

func decodeFSAppendFileRequest(args map[string]any) (FSAppendFileRequest, error) {
	path, err := readString(args, "path")
	if err != nil {
		return FSAppendFileRequest{}, err
	}
	data, err := readBytes(args, "data")
	if err != nil {
		return FSAppendFileRequest{}, err
	}
	return FSAppendFileRequest{Path: path, Data: data}, nil
}

func decodeFSReadDirRequest(args map[string]any) (FSReadDirRequest, error) {
	path, err := readString(args, "path")
	if err != nil {
		return FSReadDirRequest{}, err
	}
	return FSReadDirRequest{Path: path}, nil
}

func encodeFSReadDirResponse(entries []hostfs.DirEntry) CallResult {
	return CallResult{Value: map[string]any{"entries": entries}}
}

func decodeFSStatRequest(args map[string]any) (FSStatRequest, error) {
	path, err := readString(args, "path")
	if err != nil {
		return FSStatRequest{}, err
	}
	return FSStatRequest{Path: path}, nil
}

func encodeFSStatResponse(info hostfs.FileInfo) CallResult {
	return CallResult{Value: map[string]any{"info": info}}
}

func decodeFSMkdirRequest(args map[string]any) (FSMkdirRequest, error) {
	path, err := readString(args, "path")
	if err != nil {
		return FSMkdirRequest{}, err
	}
	return FSMkdirRequest{
		Path: path,
		Opts: hostfs.MkdirOptions{Recursive: readBool(args, "recursive")},
	}, nil
}

func decodeFSRemoveRequest(args map[string]any) (FSRemoveRequest, error) {
	path, err := readString(args, "path")
	if err != nil {
		return FSRemoveRequest{}, err
	}
	return FSRemoveRequest{
		Path: path,
		Opts: hostfs.RemoveOptions{Recursive: readBool(args, "recursive")},
	}, nil
}

func decodeFSRenameRequest(args map[string]any) (FSRenameRequest, error) {
	oldPath, err := readStringAny(args, "old_path", "oldPath")
	if err != nil {
		return FSRenameRequest{}, err
	}
	newPath, err := readStringAny(args, "new_path", "newPath")
	if err != nil {
		return FSRenameRequest{}, err
	}
	return FSRenameRequest{OldPath: oldPath, NewPath: newPath}, nil
}

func emptyCallResult() CallResult {
	return CallResult{Value: map[string]any{}}
}
