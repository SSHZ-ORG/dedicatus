package tgapi

import "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	FileTypeMPEG4GIF = "mpeg4_gif"
	FileTypeVideo    = "video"
)

type TGFile struct {
	FileID       string
	FileUniqueID string
	FileName     string
	MimeType     string
	FileSize     int

	FileType string
}

func TGFileFromChatAnimation(animation *tgbotapi.Animation) *TGFile {
	return &TGFile{
		FileID:       animation.FileID,
		FileUniqueID: animation.FileUniqueID,
		FileName:     animation.FileName,
		MimeType:     animation.MimeType,
		FileSize:     animation.FileSize,
		FileType:     FileTypeMPEG4GIF,
	}
}

func TGFileFromVideo(video *tgbotapi.Video) *TGFile {
	return &TGFile{
		FileID:       video.FileID,
		FileUniqueID: video.FileUniqueID,
		FileName:     "", // Video does not have FileName.
		MimeType:     video.MimeType,
		FileSize:     video.FileSize,
		FileType:     FileTypeVideo,
	}
}

func MakeInlineQueryResult(uniqueID, fileID, fileType string) interface{} {
	switch fileType {
	case FileTypeMPEG4GIF:
		return tgbotapi.NewInlineQueryResultCachedMPEG4GIF(uniqueID, fileID)
	case FileTypeVideo:
		return tgbotapi.NewInlineQueryResultCachedVideo(uniqueID, fileID, uniqueID)
	}
	return nil
}

func MakeFileable(chatID int64, fileID, fileType, caption string) tgbotapi.Fileable {
	switch fileType {
	case FileTypeMPEG4GIF:
		share := tgbotapi.NewAnimation(chatID, tgbotapi.FileID(fileID))
		share.Caption = caption
		return share
	case FileTypeVideo:
		share := tgbotapi.NewVideo(chatID, tgbotapi.FileID(fileID))
		share.Caption = caption
		return share
	}
	return nil
}
