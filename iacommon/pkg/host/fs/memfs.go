package fs

import (
	"context"
)

type MemFSProvider struct{}

func (p *MemFSProvider) Open(ctx context.Context, path string, opts OpenOptions) (FileHandle, error) {
	_, _, _, _ = ctx, path, opts, p
	return nil, nil
}

func (p *MemFSProvider) ReadFile(ctx context.Context, path string) ([]byte, error) {
	_, _, _ = ctx, path, p
	return nil, nil
}

func (p *MemFSProvider) WriteFile(ctx context.Context, path string, data []byte, opts WriteOptions) error {
	_, _, _, _, _ = ctx, path, data, opts, p
	return nil
}

func (p *MemFSProvider) AppendFile(ctx context.Context, path string, data []byte) error {
	_, _, _, _ = ctx, path, data, p
	return nil
}

func (p *MemFSProvider) ReadDir(ctx context.Context, path string) ([]DirEntry, error) {
	_, _, _ = ctx, path, p
	return nil, nil
}

func (p *MemFSProvider) Stat(ctx context.Context, path string) (FileInfo, error) {
	_, _, _ = ctx, path, p
	return FileInfo{}, nil
}

func (p *MemFSProvider) Mkdir(ctx context.Context, path string, opts MkdirOptions) error {
	_, _, _, _ = ctx, path, opts, p
	return nil
}

func (p *MemFSProvider) Remove(ctx context.Context, path string, opts RemoveOptions) error {
	_, _, _, _ = ctx, path, opts, p
	return nil
}

func (p *MemFSProvider) Rename(ctx context.Context, oldPath, newPath string) error {
	_, _, _, _ = ctx, oldPath, newPath, p
	return nil
}
