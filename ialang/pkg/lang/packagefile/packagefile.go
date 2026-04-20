package packagefile

import common "iacommon/pkg/ialang/packagefile"

type Package = common.Package

const (
	PackageFormatMagic   = common.PackageFormatMagic
	PackageFormatVersion = common.PackageFormatVersion
)

func Encode(pkg *Package) ([]byte, error) {
	return common.Encode(pkg)
}

func Decode(data []byte) (*Package, error) {
	return common.Decode(data)
}
