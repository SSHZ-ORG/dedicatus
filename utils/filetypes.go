package utils

const (
	FileTypeMPEG4GIF = "mpeg4_gif"
)

type InlineQueryResultCachedMPEG4GIF struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	FileID string `json:"mpeg4_file_id"`
}

func MakeInlineQueryResultCachedMPEG4GIF(fileID string) InlineQueryResultCachedMPEG4GIF {
	return InlineQueryResultCachedMPEG4GIF{
		Type:   "mpeg4_gif",
		ID:     fileID,
		FileID: fileID,
	}
}

func MakeInlineQueryResult(fileID, fileType string) interface{} {
	switch fileType {
	case FileTypeMPEG4GIF:
		return MakeInlineQueryResultCachedMPEG4GIF(fileID)
	}
	return nil
}
