package utils

const (
	FileTypeMPEG4GIF = "mpeg4_gif"
)

type InlineQueryResultCachedMPEG4GIF struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	FileID string `json:"mpeg4_file_id"`
}

func makeInlineQueryResultCachedMPEG4GIF(uniqueID, fileID string) InlineQueryResultCachedMPEG4GIF {
	return InlineQueryResultCachedMPEG4GIF{
		Type:   "mpeg4_gif",
		ID:     uniqueID,
		FileID: fileID,
	}
}

func MakeInlineQueryResult(uniqueID, fileID, fileType string) interface{} {
	switch fileType {
	case FileTypeMPEG4GIF:
		return makeInlineQueryResultCachedMPEG4GIF(uniqueID, fileID)
	}
	return nil
}
