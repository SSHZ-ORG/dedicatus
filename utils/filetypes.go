package utils

import "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	FileTypeMPEG4GIF = "mpeg4_gif"
)

func MakeInlineQueryResult(uniqueID, fileID, fileType string) interface{} {
	switch fileType {
	case FileTypeMPEG4GIF:
		return tgbotapi.NewInlineQueryResultCachedMPEG4GIF(uniqueID, fileID)
	}
	return nil
}
