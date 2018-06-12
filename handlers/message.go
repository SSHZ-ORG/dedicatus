package handlers

import (
	"fmt"
	"strings"

	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"gopkg.in/telegram-bot-api.v4"
)

func HandleMessage(ctx context.Context, update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	message := update.Message
	if message.Document != nil {
		replyMessage := ""

		fileID := message.Document.FileID

		i, err := models.GetInventory(ctx, fileID)
		if err == nil {
			s, err := i.ToString(ctx)
			if err != nil {
				return err
			}

			replyMessage = fmt.Sprintf("I know this GIF %s", s)
		} else {
			replyMessage = fmt.Sprintf("New GIF %s", fileID)
		}

		if replyMessage != "" {
			reply := tgbotapi.NewMessage(message.Chat.ID, replyMessage)
			reply.ReplyToMessageID = message.MessageID
			_, err := bot.Send(reply)
			return err
		}
	}

	if strings.HasPrefix(message.Text, "/") {
		replyMessage := ""
		args := strings.Split(message.Text, " ")
		userID := message.From.ID

		config := models.GetConfig(ctx)
		admins := utils.NewIntSetFromSlice(config.Admins)
		isAdmin := admins.Contains(userID)

		switch args[0] {
		case "/me":
			replyMessage = fmt.Sprintf("User %d", userID)
		case "/n":
			if isAdmin {
				replyMessage = commandCreatePersonality(ctx, args)
			} else {
				replyMessage = "Unauthorized"
			}
		case "/s":
			replyMessage = commandFindPersonality(ctx, args)
		case "/u":
			if isAdmin {
				replyMessage = commandUpdatePersonalityNickname(ctx, args)
			} else {
				replyMessage = "Unauthorized"
			}
		case "/g":
			if isAdmin {
				replyMessage = commandRegisterInventory(ctx, args, userID)
			} else {
				replyMessage = "Unauthorized"
			}
		}

		if replyMessage != "" {
			reply := tgbotapi.NewMessage(message.Chat.ID, replyMessage)
			reply.ReplyToMessageID = message.MessageID
			_, err := bot.Send(reply)
			return err
		}
	}

	return nil
}

func commandCreatePersonality(ctx context.Context, args []string) string {
	if len(args) != 3 {
		return "Usage: /n <CanonicalName> <KnowledgeGraphID>\nExample: /n 井口裕香 /m/064m4km"
	}

	name := args[1]
	KGID := args[2]
	if _, p, _ := models.GetPersonalityByName(ctx, name); p != nil {
		return fmt.Sprintf("Personality known as %s", p.ToString())
	}
	if p, _ := models.GetPersonalityByKGID(ctx, KGID); p != nil {
		return fmt.Sprintf("Personality known as %s", p.ToString())
	}

	p, err := models.CreatePersonality(ctx, KGID, name)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Created Personality %s", p.ToString())
}

func commandFindPersonality(ctx context.Context, args []string) string {
	if len(args) != 2 {
		return "Usage: /s <Query>\nExample: /s 井口裕香"
	}

	query := args[1]

	key, err := models.TryFindPersonalityWithKG(ctx, query)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		return "Error"
	}
	if key == nil {
		return "Not found"
	}

	p, err := models.GetPersonality(ctx, key)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		return "Error"
	}
	return fmt.Sprintf("Found Personality %s", p.ToString())
}

func commandUpdatePersonalityNickname(ctx context.Context, args []string) string {
	if len(args) != 4 || (args[1] != "add" && args[1] != "delete") {
		return "Usage: /u add|delete <CanonicalName> <Nickname>\nExample: /u add 井口裕香 知性"
	}

	name := args[2]
	nickname := args[3]

	key, p, err := models.GetPersonalityByName(ctx, name)
	if err != nil {
		return err.Error()
	}
	if key == nil {
		return "Personality not found"
	}

	nicknames := utils.NewStringSetFromSlice(p.Nickname)
	switch args[1] {
	case "add":
		maybeConflictKey, err := models.TryFindPersonality(ctx, nickname)
		if err != nil {
			return err.Error()
		}
		if maybeConflictKey != nil {
			conflictP, err := models.GetPersonality(ctx, maybeConflictKey)
			if err != nil {
				return err.Error()
			}
			return fmt.Sprintf("Will conflict with Personality %s", conflictP.ToString())
		}
		nicknames.Add(nickname)
	case "delete":
		nicknames.Remove(nickname)
	}

	p.Nickname = nicknames.ToSlice()
	_, err = nds.Put(ctx, key, p)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Updated Personality %s", p.ToString())
}

func commandRegisterInventory(ctx context.Context, args []string, userID int) string {
	if len(args) < 3 {
		return "Usage: /g <FileID> <Nickname...>\nExample: /g ABCDEFGH 井口裕香 佐藤利奈"
	}

	fileID := args[1]
	nicknames := args[2:]

	var keys []*datastore.Key
	for _, nickname := range nicknames {
		key, err := models.TryFindPersonality(ctx, nickname)
		if err != nil {
			return err.Error()
		}

		if key == nil {
			return fmt.Sprintf("Unknown Personality %s", nickname)
		}
		keys = append(keys, key)
	}

	i, err := models.CreateInventory(ctx, fileID, keys, userID)
	if err != nil {
		return err.Error()
	}

	s, err := i.ToString(ctx)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("Updated Inventory %s", s)
}
