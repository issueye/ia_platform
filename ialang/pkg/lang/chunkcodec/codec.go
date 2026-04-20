package chunkcodec

import (
	common "iacommon/pkg/ialang/chunkcodec"
	bc "ialang/pkg/lang/bytecode"
)

const (
	FormatMagic   = common.FormatMagic
	FormatVersion = common.FormatVersion
)

func Serialize(chunk *bc.Chunk) ([]byte, error) {
	return common.Serialize(chunk)
}

func Deserialize(data []byte) (*bc.Chunk, error) {
	return common.Deserialize(data)
}
