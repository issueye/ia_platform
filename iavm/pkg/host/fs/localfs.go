package fs

import (
	"context"
	"errors"
	"os"
)

var ErrPathMapperNotConfigured = errors.New("path mapper is not configured")

type LocalFSProvider struct {
	Mapper PathMapper
}

type localFileHandle struct {
	file *os.File
}

func (p *LocalFSProvider) Open(ctx context.Context, path string, opts OpenOptions) (FileHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	realPath, preopen, err := p.resolve(path)
	if err != nil {
		return nil, err
	}
	if openNeedsWrite(opts) && preopen.ReadOnly {
		return nil, ErrReadOnlyPreopen
	}

	file, err := os.OpenFile(realPath, openFlags(opts), 0o644)
	if err != nil {
		return nil, err
	}
	return &localFileHandle{file: file}, nil
}

func (p *LocalFSProvider) ReadFile(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	realPath, _, err := p.resolve(path)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(realPath)
}

func (p *LocalFSProvider) WriteFile(ctx context.Context, path string, data []byte, opts WriteOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	realPath, preopen, err := p.resolve(path)
	if err != nil {
		return err
	}
	if preopen.ReadOnly {
		return ErrReadOnlyPreopen
	}

	flags := os.O_WRONLY
	if opts.Create {
		flags |= os.O_CREATE
	}
	if opts.Trunc {
		flags |= os.O_TRUNC
	}

	file, err := os.OpenFile(realPath, flags, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func (p *LocalFSProvider) AppendFile(ctx context.Context, path string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	realPath, preopen, err := p.resolve(path)
	if err != nil {
		return err
	}
	if preopen.ReadOnly {
		return ErrReadOnlyPreopen
	}

	file, err := os.OpenFile(realPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func (p *LocalFSProvider) ReadDir(ctx context.Context, path string) ([]DirEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	realPath, _, err := p.resolve(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(realPath)
	if err != nil {
		return nil, err
	}

	result := make([]DirEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, DirEntry{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		})
	}
	return result, nil
}

func (p *LocalFSProvider) Stat(ctx context.Context, path string) (FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return FileInfo{}, err
	}

	realPath, _, err := p.resolve(path)
	if err != nil {
		return FileInfo{}, err
	}

	info, err := os.Stat(realPath)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode().String(),
		IsDir:   info.IsDir(),
		ModUnix: info.ModTime().Unix(),
	}, nil
}

func (p *LocalFSProvider) Mkdir(ctx context.Context, path string, opts MkdirOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	realPath, preopen, err := p.resolve(path)
	if err != nil {
		return err
	}
	if preopen.ReadOnly {
		return ErrReadOnlyPreopen
	}

	if opts.Recursive {
		return os.MkdirAll(realPath, 0o755)
	}
	return os.Mkdir(realPath, 0o755)
}

func (p *LocalFSProvider) Remove(ctx context.Context, path string, opts RemoveOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	realPath, preopen, err := p.resolve(path)
	if err != nil {
		return err
	}
	if preopen.ReadOnly {
		return ErrReadOnlyPreopen
	}

	if opts.Recursive {
		return os.RemoveAll(realPath)
	}
	return os.Remove(realPath)
}

func (p *LocalFSProvider) Rename(ctx context.Context, oldPath, newPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	oldRealPath, oldPreopen, err := p.resolve(oldPath)
	if err != nil {
		return err
	}
	if oldPreopen.ReadOnly {
		return ErrReadOnlyPreopen
	}

	newRealPath, newPreopen, err := p.resolve(newPath)
	if err != nil {
		return err
	}
	if newPreopen.ReadOnly {
		return ErrReadOnlyPreopen
	}

	return os.Rename(oldRealPath, newRealPath)
}

func (p *LocalFSProvider) resolve(virtualPath string) (string, Preopen, error) {
	if p == nil || p.Mapper == nil {
		return "", Preopen{}, ErrPathMapperNotConfigured
	}
	return p.Mapper.Resolve(virtualPath)
}

func openNeedsWrite(opts OpenOptions) bool {
	return opts.Write || opts.Create || opts.Trunc || opts.Append
}

func openFlags(opts OpenOptions) int {
	flags := os.O_RDONLY
	if opts.Read && openNeedsWrite(opts) {
		flags = os.O_RDWR
	} else if openNeedsWrite(opts) {
		flags = os.O_WRONLY
	}
	if opts.Create {
		flags |= os.O_CREATE
	}
	if opts.Trunc {
		flags |= os.O_TRUNC
	}
	if opts.Append {
		flags |= os.O_APPEND
	}
	return flags
}

func (h *localFileHandle) Read(ctx context.Context, p []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return h.file.Read(p)
}

func (h *localFileHandle) Write(ctx context.Context, p []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return h.file.Write(p)
}

func (h *localFileHandle) Seek(ctx context.Context, offset int64, whence int) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return h.file.Seek(offset, whence)
}

func (h *localFileHandle) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return h.file.Close()
}
