package messages

import (
	"fmt"
	"strings"

	"github.com/SSHZ-ORG/dedicatus/dctx"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func handleAnimation(ctx context.Context, message *tgbotapi.Message) (tgbotapi.Chattable, error) {
	animation := message.Animation
	if message.ReplyToMessage != nil && message.ReplyToMessage.Animation != nil {
		animation = message.ReplyToMessage.Animation
	}

	tgFile := tgapi.TGFileFromChatAnimation(animation)
	return handleTGFile(ctx, message, tgFile)
}

func handleVideo(ctx context.Context, message *tgbotapi.Message) (tgbotapi.Chattable, error) {
	video := message.Video
	if message.ReplyToMessage != nil && message.ReplyToMessage.Video != nil {
		video = message.ReplyToMessage.Video
	}

	tgFile := tgapi.TGFileFromVideo(video)
	return handleTGFile(ctx, message, tgFile)
}

func handleTGFile(ctx context.Context, message *tgbotapi.Message, tgFile *tgapi.TGFile) (tgbotapi.Chattable, error) {
	caption := message.Caption
	if message.ReplyToMessage != nil {
		caption = message.Text
	}

	replyMessages := []string{"Received:\n" + tgFile.FileName + "\nUniqueID: " + tgFile.FileUniqueID}

	allowUpdate := true

	// Check MimeType. Both MPEG4_GIF and Video have to be video/mp4 for now.
	if tgFile.MimeType != "video/mp4" {
		replyMessages = append(replyMessages, fmt.Sprintf("Unexpected file type %s, disallowing update.", tgFile.MimeType))
		allowUpdate = false
	}

	// Match Inventory
	i, err := models.TryGetInventoryByTgFile(ctx, tgFile)
	if err != nil {
		if err == models.ErrorHashConflict {
			replyMessages = append(replyMessages, "Hash conflict!")
			allowUpdate = false
		} else if err != datastore.ErrNoSuchEntity {
			return nil, err
		}
	}

	if i != nil {
		// It matched to some existing Inventory, pretend that we received the existing one.
		tgFile.FileUniqueID = i.FileUniqueID
		tgFile.FileID = i.FileID

		s, err := i.ToString(ctx)
		if err != nil {
			return nil, err
		}
		replyMessages = append(replyMessages, "\nMatched to:\n"+s)
	} else {
		replyMessages = append(replyMessages, "\nNo matching Inventory found.")
	}

	// OK, we either didn't know this Inventory before, or it matches something without hash conflict.
	if allowUpdate {
		if caption != "" {
			// Update Inventory, if instructed
			resp, err := handleTGFileCaption(ctx, tgFile, caption)
			if err != nil {
				return nil, err
			}
			replyMessages = append(replyMessages, "\n"+resp)
		} else {
			if i != nil && i.FileName == "" && tgFile.FileName != "" {
				// Existing Inventory but we don't know FileName, and we know FileName now. Backfill the field.
				if err := models.OverrideFileName(ctx, i.FileUniqueID, tgFile.FileName); err != nil {
					return nil, err
				}
			}
		}
	}

	if len(replyMessages) > 0 {
		return makeReplyMessage(message, strings.Join(replyMessages, "\n")), nil
	}
	return nil, nil
}

func handleTGFileCaption(ctx context.Context, tgFile *tgapi.TGFile, caption string) (string, error) {
	if !dctx.IsContributor(ctx) {
		return errorMessageNotContributor, nil
	}

	args := strings.Fields(caption)

	if args[0] == "/r" {
		// Use the received document to replace some existing Inventory.
		if !dctx.IsAdmin(ctx) {
			return errorMessageNotAdmin, nil
		}

		if len(args) != 2 {
			return "Usage:\n/r <OldFileUniqueID>", nil
		}

		oldFileUniqueID := args[1]
		i, err := models.ReplaceFileID(ctx, oldFileUniqueID, tgFile)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				return "Old Inventory not found", nil
			}
			return "", err
		}

		s, err := i.ToString(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Replaced Inventory %s with:\n%s", oldFileUniqueID, s), nil
	}

	var personalityKeys []*datastore.Key
	for _, nickname := range args {
		key, err := models.TryFindOnlyPersonality(ctx, nickname)
		if err != nil {
			return "", err
		}

		if key == nil {
			return fmt.Sprintf("Unknown Personality %s", nickname), nil
		}
		personalityKeys = append(personalityKeys, key)
	}

	i, err := models.CreateOrUpdateInventory(ctx, tgFile, personalityKeys)
	if err != nil {
		if err == models.ErrorOnlyAdminCanUpdateInventory {
			return "This GIF is already known. Only admins or its creator can modify it now.", nil
		}
		return "", err
	}

	s, err := i.ToString(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Updated to:\n%s", s), nil
}
