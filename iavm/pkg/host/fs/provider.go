package fs

import "context"

type Provider interface {
	Open(ctx context.Context, path string, opts OpenOptions) (FileHandle, error)
	ReadFile(ctx context.Context, path string) ([]byte, error)
	WriteFile(ctx context.Context, path string, data []byte, opts WriteOptions) error
	AppendFile(ctx context.Context, path string, data []byte) error
	ReadDir(ctx context.Context, path string) ([]DirEntry, error)
	Stat(ctx context.Context, path string) (FileInfo, error)
	Mkdir(ctx context.Context, path string, opts MkdirOptions) error
	Remove(ctx context.Context, path string, opts RemoveOptions) error
	Rename(ctx context.Context, oldPath, newPath string) error
}

type OpenOptions struct {
	Read   bool
	Write  bool
	Create bool
	Trunc  bool
	Append bool
}

type WriteOptions struct {
	Create bool
	Trunc  bool
}

type MkdirOptions struct {
	Recursive bool
}

type RemoveOptions struct {
	Recursive bool
}

type FileHandle interface {
	Read(ctx context.Context, p []byte) (int, error)
	Write(ctx context.Context, p []byte) (int, error)
	Seek(ctx context.Context, offset int64, whence int) (int64, error)
	Close(ctx context.Context) error
}

type FileInfo struct {
	Name    string
	Size    int64
	Mode    string
	IsDir   bool
	ModUnix int64
}

type DirEntry struct {
	Name  string
	IsDir bool
}
