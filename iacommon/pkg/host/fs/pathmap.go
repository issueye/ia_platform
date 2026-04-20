package fs

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

var (
	ErrInvalidVirtualPath = errors.New("invalid virtual path")
	ErrPathNotMapped      = errors.New("path is not covered by any preopen")
	ErrReadOnlyPreopen    = errors.New("path is in a read-only preopen")
)

type Preopen struct {
	VirtualPath string
	RealPath    string
	ReadOnly    bool
}

type PathMapper interface {
	Resolve(virtualPath string) (realPath string, matchedPreopen Preopen, err error)
}

type PreopenPathMapper struct {
	preopens []Preopen
}

func NewPreopenPathMapper(preopens []Preopen) (*PreopenPathMapper, error) {
	normalized := make([]Preopen, 0, len(preopens))
	for _, preopen := range preopens {
		virtualPath, err := cleanVirtualPath(preopen.VirtualPath)
		if err != nil {
			return nil, fmt.Errorf("normalize preopen %q: %w", preopen.VirtualPath, err)
		}
		if preopen.RealPath == "" {
			return nil, fmt.Errorf("normalize preopen %q: empty real path", virtualPath)
		}

		realPath, err := filepath.Abs(preopen.RealPath)
		if err != nil {
			return nil, fmt.Errorf("normalize preopen %q real path: %w", virtualPath, err)
		}

		normalized = append(normalized, Preopen{
			VirtualPath: virtualPath,
			RealPath:    filepath.Clean(realPath),
			ReadOnly:    preopen.ReadOnly,
		})
	}

	sort.Slice(normalized, func(i, j int) bool {
		if len(normalized[i].VirtualPath) == len(normalized[j].VirtualPath) {
			return normalized[i].VirtualPath < normalized[j].VirtualPath
		}
		return len(normalized[i].VirtualPath) > len(normalized[j].VirtualPath)
	})

	return &PreopenPathMapper{preopens: normalized}, nil
}

func (m *PreopenPathMapper) Resolve(virtualPath string) (string, Preopen, error) {
	if m == nil {
		return "", Preopen{}, ErrPathNotMapped
	}

	cleanedPath, err := cleanVirtualPath(virtualPath)
	if err != nil {
		return "", Preopen{}, err
	}

	for _, preopen := range m.preopens {
		if !matchesPreopen(preopen.VirtualPath, cleanedPath) {
			continue
		}

		realPath, err := resolveWithinPreopen(preopen, cleanedPath)
		if err != nil {
			return "", Preopen{}, err
		}
		return realPath, preopen, nil
	}

	return "", Preopen{}, fmt.Errorf("%w: %s", ErrPathNotMapped, cleanedPath)
}

func cleanVirtualPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "/", nil
	}

	value = strings.ReplaceAll(value, "\\", "/")
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}

	cleaned := path.Clean(value)
	if cleaned == "." {
		cleaned = "/"
	}
	if !strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("%w: %s", ErrInvalidVirtualPath, value)
	}
	return cleaned, nil
}

func matchesPreopen(preopenRoot, target string) bool {
	if preopenRoot == "/" {
		return true
	}
	return target == preopenRoot || strings.HasPrefix(target, preopenRoot+"/")
}

func resolveWithinPreopen(preopen Preopen, virtualPath string) (string, error) {
	rel := strings.TrimPrefix(virtualPath, preopen.VirtualPath)
	rel = strings.TrimPrefix(rel, "/")

	realPath := preopen.RealPath
	if rel != "" {
		realPath = filepath.Join(preopen.RealPath, filepath.FromSlash(rel))
	}
	realPath = filepath.Clean(realPath)

	if err := ensureWithinPreopen(preopen.RealPath, realPath); err != nil {
		return "", err
	}

	return realPath, nil
}

func ensureWithinPreopen(root, target string) error {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return fmt.Errorf("resolve preopen path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%w: %s", ErrPathNotMapped, target)
	}
	return nil
}
