package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	errorMessageNotAdmin       = "Only admins can do this, sorry."
	errorMessageNotContributor = "Only contributors can do this, sorry."
)

var commandMap = map[string]func(ctx context.Context, args []string, userID int) (string, error){
	"/me": commandUserInfo,
	"/n":  commandCreatePersonality,
	"/s":  commandFindPersonality,
	"/u":  commandUpdatePersonalityNickname,
	"/g":  commandRegisterInventory,
	"/c":  commandManageContributors,
	"/kg": commandQueryKG,
}

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
		var err error

		args := strings.Split(message.Text, " ")
		userID := message.From.ID

		if handler, ok := commandMap[args[0]]; ok {
			replyMessage, err = handler(ctx, args, userID)
		} else {
			replyMessage = "Command not recognized."
		}

		if err != nil {
			reply := tgbotapi.NewMessage(message.Chat.ID, "Your action triggered an internal server error.")
			reply.ReplyToMessageID = message.MessageID
			bot.Send(reply) // Fire and forget
			return err
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

func commandUserInfo(ctx context.Context, args []string, userID int) (string, error) {
	c := models.GetConfig(ctx)

	return fmt.Sprintf("User %d\nisAdmin: %v\nisContributor: %v", userID, c.IsAdmin(userID), c.IsContributor(userID)), nil
}

func commandCreatePersonality(ctx context.Context, args []string, userID int) (string, error) {
	c := models.GetConfig(ctx)
	if !c.IsAdmin(userID) {
		return errorMessageNotAdmin, nil
	}

	if len(args) != 3 {
		return "Usage:\n/n <CanonicalName> <KnowledgeGraphID>\nExample: /n 井口裕香 /m/064m4km", nil
	}

	name := args[1]
	KGID := args[2]

	_, p, err := models.GetPersonalityByName(ctx, name)
	if err != nil {
		return "", err
	}
	if p != nil {
		return fmt.Sprintf("Personality known as %s", p.ToString()), nil
	}

	p, err = models.GetPersonalityByKGID(ctx, KGID)
	if err != nil {
		return "", err
	}
	if p != nil {
		return fmt.Sprintf("Personality known as %s", p.ToString()), nil
	}

	p, err = models.CreatePersonality(ctx, KGID, name)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Created Personality %s", p.ToString()), nil
}

func commandFindPersonality(ctx context.Context, args []string, userID int) (string, error) {
	if len(args) != 2 {
		return "Usage:\n/s <Query>\nExample: /s 井口裕香", nil
	}

	query := args[1]

	key, err := models.TryFindPersonalityWithKG(ctx, query)
	if err != nil {
		return "", err
	}
	if key == nil {
		return "Not found", nil
	}

	p, err := models.GetPersonality(ctx, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Found Personality %s", p.ToString()), nil
}

func commandUpdatePersonalityNickname(ctx context.Context, args []string, userID int) (string, error) {
	c := models.GetConfig(ctx)
	if !c.IsAdmin(userID) {
		return errorMessageNotAdmin, nil
	}

	if len(args) != 4 || (args[1] != "add" && args[1] != "delete") {
		return "Usage:\n/u add|delete <CanonicalName> <Nickname>\nExample: /u add 井口裕香 知性", nil
	}

	name := args[2]
	nickname := strings.ToLower(args[3])

	key, p, err := models.GetPersonalityByName(ctx, name)
	if err != nil {
		return "", err
	}
	if key == nil {
		return "Personality not found", nil
	}

	nicknames := utils.NewStringSetFromSlice(p.Nickname)
	switch args[1] {
	case "add":
		maybeConflictKey, err := models.TryFindPersonality(ctx, nickname)
		if err != nil {
			return "", err
		}
		if maybeConflictKey != nil {
			conflictP, err := models.GetPersonality(ctx, maybeConflictKey)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Will conflict with Personality %s", conflictP.ToString()), nil
		}
		nicknames.Add(nickname)
	case "delete":
		nicknames.Remove(nickname)
	}

	p.Nickname = nicknames.ToSlice()
	_, err = nds.Put(ctx, key, p)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Updated Personality %s", p.ToString()), nil
}

func commandRegisterInventory(ctx context.Context, args []string, userID int) (string, error) {
	c := models.GetConfig(ctx)
	if !c.IsContributor(userID) {
		return errorMessageNotContributor, nil
	}

	if len(args) < 3 {
		return "Usage:\n/g <FileID> <Nickname...>\nExample: /g ABCDEFGH 井口裕香 佐藤利奈", nil
	}

	fileID := args[1]
	nicknames := args[2:]

	var keys []*datastore.Key
	for _, nickname := range nicknames {
		key, err := models.TryFindPersonality(ctx, nickname)
		if err != nil {
			return "", err
		}

		if key == nil {
			return fmt.Sprintf("Unknown Personality %s", nickname), nil
		}
		keys = append(keys, key)
	}

	i, err := models.CreateInventory(ctx, fileID, keys, userID, c)
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
	return fmt.Sprintf("Updated Inventory %s", s), nil
}

func commandManageContributors(ctx context.Context, args []string, userID int) (string, error) {
	c := models.GetConfig(ctx)
	if !c.IsAdmin(userID) {
		return errorMessageNotAdmin, nil
	}

	if len(args) != 3 || (args[1] != "add" && args[1] != "delete") {
		return "Usage:\n/c add|delete <UserId>\nExample: /c add 88888888", nil
	}

	newContributor, err := strconv.Atoi(args[2])
	if err != nil {
		return "Illegal UserID.", nil
	}

	contributors := utils.NewIntSetFromSlice(c.Contributors)
	switch args[1] {
	case "add":
		if !contributors.Add(newContributor) {
			return "Is already a contributor.", nil
		}
	case "delete":
		if !contributors.Remove(newContributor) {
			return "Is not a contributor.", nil
		}
	}

	c.Contributors = contributors.ToSlice()
	if err = models.SetConfig(ctx, c); err != nil {
		return "", err
	}

	return "Contributors updated", nil
}

func commandQueryKG(ctx context.Context, args []string, userID int) (string, error) {
	c := models.GetConfig(ctx)
	if !c.IsAdmin(userID) {
		return errorMessageNotAdmin, nil
	}

	if len(args) != 2 {
		return "Usage:\n/kg <Query>\nExample: /kg 井口裕香", nil
	}
	return utils.GetKGQueryResult(ctx, args[1])
}
