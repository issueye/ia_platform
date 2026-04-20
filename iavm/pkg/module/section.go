package module

type SectionID uint8

const (
	SectionType SectionID = iota + 1
	SectionImport
	SectionFunction
	SectionGlobal
	SectionExport
	SectionCode
	SectionData
	SectionCapability
	SectionCustom
)

type SectionHeader struct {
	ID     SectionID
	Length uint32
}
