package metadatamode

type MetadataMode string

const (
	Default      MetadataMode = ""
	ReadFileName MetadataMode = "READ_FILE_NAME"
)

func (m MetadataMode) ToString() string {
	return string(m)
}

func FromString(s string) MetadataMode {
	return MetadataMode(s)
}
