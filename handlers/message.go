package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

const (
	errorMessageNotAdmin       = "Only admins can do this, sorry."
	errorMessageNotContributor = "Only contributors can do this, sorry."
)

var commandMap = map[string]func(ctx context.Context, args []string, userID int) (string, error){
	"/start": commandStart,
	"/me":    commandUserInfo,
	"/n":     commandCreatePersonality,
	"/s":     commandFindPersonality,
	"/u":     commandUpdatePersonalityNickname,
	"/a":     commandEditAlias,
	"/g":     commandRegisterInventory,
	"/r":     commandReplaceInventoryFileID,
	"/c":     commandManageContributors,
	"/kg":    commandQueryKG,
	"/stats": commandStats,
}

func HandleMessage(ctx context.Context, update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	message := update.Message
	if message.Document != nil {

		fileID := message.Document.FileID
		fileUniqueID := message.Document.FileUniqueID

		replyMessages := []string{"Received FileUniqueID: " + fileUniqueID, "FileID: " + fileID}

		i, err := models.GetInventory(ctx, fileUniqueID)
		if err != nil {
			if err != datastore.ErrNoSuchEntity {
				return err
			}
			i, err = models.GetInventoryByFile(ctx, fileID, message.Document.FileSize)
			if err != nil {
				return err
			}
		}

		if i != nil {
			s, err := i.ToString(ctx)
			if err != nil {
				return err
			}

			replyMessages = append(replyMessages, s)
		} else {
			replyMessages = append(replyMessages, "No matching Inventory found.")
		}

		if len(replyMessages) > 0 {
			reply := tgbotapi.NewMessage(message.Chat.ID, strings.Join(replyMessages, "\n"))
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

func commandStart(ctx context.Context, args []string, userID int) (string, error) {
	return fmt.Sprintf("Dedicatus %s", appengine.VersionID(ctx)), nil
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

	key, p, err := models.GetPersonalityByName(ctx, name)
	if err != nil {
		return "", err
	}
	if p != nil {
		s, err := p.ToString(ctx, key)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Personality known as %s", s), nil
	}

	key, p, err = models.GetPersonalityByKGID(ctx, KGID)
	if err != nil {
		return "", err
	}
	if p != nil {
		s, err := p.ToString(ctx, key)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Personality known as %s", s), nil
	}

	key, p, err = models.CreatePersonality(ctx, KGID, name)
	if err != nil {
		return "", err
	}

	s, err := p.ToString(ctx, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Created Personality %s", s), nil
}

func commandFindPersonality(ctx context.Context, args []string, userID int) (string, error) {
	if len(args) != 2 {
		return "Usage:\n/s <Query>\nExample: /s 井口裕香", nil
	}

	query := args[1]

	keys, err := models.TryFindPersonalitiesWithKG(ctx, query)
	if err != nil {
		return "", err
	}
	if len(keys) == 0 {
		return "Not found", nil
	}

	ps, err := models.GetPersonalities(ctx, keys)
	if err != nil {
		return "", err
	}

	var ss []string
	for i, p := range ps {
		s, err := p.ToString(ctx, keys[i])
		if err != nil {
			return "", err
		}
		ss = append(ss, s)
	}

	return fmt.Sprintf("Found Personality:\n%s", strings.Join(ss, "\n")), nil
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
	alias := strings.ToLower(args[3])

	key, _, err := models.GetPersonalityByName(ctx, name)
	if err != nil {
		return "", err
	}
	if key == nil {
		return "Personality not found", nil
	}

	var a *models.Alias
	switch args[1] {
	case "add":
		a, err = models.AddAlias(ctx, alias, key)
	case "delete":
		a, err = models.DeleteAlias(ctx, alias, key)
	}

	if err != nil {
		return "", err
	}

	if a != nil {
		s, err := a.ToString(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Updated Alias %s", s), nil
	} else {
		return fmt.Sprintf("Alias %s no longer exists", alias), nil
	}
}

func commandEditAlias(ctx context.Context, args []string, userID int) (string, error) {
	c := models.GetConfig(ctx)
	if !c.IsAdmin(userID) {
		return errorMessageNotAdmin, nil
	}

	if len(args) != 4 || (args[1] != "cp" && args[1] != "mv") {
		return "Usage:\n/a cp|mv <Alias> <NewAlias>\nExample: /a cp あやてる てるあや", nil
	}

	alias := args[2]
	newAlias := args[3]

	var (
		a   *models.Alias
		err error
	)
	switch args[1] {
	case "cp":
		a, err = models.CopyAlias(ctx, alias, newAlias)
	case "mv":
		a, err = models.RenameAlias(ctx, alias, newAlias)
	}

	if err != nil {
		return "", err
	}

	s, err := a.ToString(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Updated Alias %s", s), nil
}

func commandRegisterInventory(ctx context.Context, args []string, userID int) (string, error) {
	return "Under migration", nil
	// c := models.GetConfig(ctx)
	// if !c.IsContributor(userID) {
	// 	return errorMessageNotContributor, nil
	// }
	//
	// if len(args) < 3 {
	// 	return "Usage:\n/g <FileID> <Nickname...>\nExample: /g ABCDEFGH 井口裕香 佐藤利奈", nil
	// }
	//
	// fileID := args[1]
	// nicknames := args[2:]
	//
	// var keys []*datastore.Key
	// for _, nickname := range nicknames {
	// 	key, err := models.TryFindOnlyPersonality(ctx, nickname)
	// 	if err != nil {
	// 		return "", err
	// 	}
	//
	// 	if key == nil {
	// 		return fmt.Sprintf("Unknown Personality %s", nickname), nil
	// 	}
	// 	keys = append(keys, key)
	// }
	//
	// i, err := models.CreateInventory(ctx, fileID, keys, userID, c)
	// if err != nil {
	// 	if err == models.ErrorOnlyAdminCanUpdateInventory {
	// 		return "This GIF is already known. Only admins or its creator can modify it now.", nil
	// 	}
	// 	return "", err
	// }
	//
	// s, err := i.ToString(ctx)
	// if err != nil {
	// 	return "", err
	// }
	// return fmt.Sprintf("Updated Inventory %s", s), nil
}

func commandReplaceInventoryFileID(ctx context.Context, args []string, userID int) (string, error) {
	return "Under migration", nil
	// c := models.GetConfig(ctx)
	// if !c.IsAdmin(userID) {
	// 	return errorMessageNotAdmin, nil
	// }
	//
	// if len(args) != 3 {
	// 	return "Usage:\n/r <OldFileID> <NewFileID>", nil
	// }
	//
	// oldFileID := args[1]
	// newFileID := args[2]
	//
	// i, err := models.ReplaceFileID(ctx, oldFileID, newFileID)
	// if err != nil {
	// 	if err == datastore.ErrNoSuchEntity {
	// 		return "Inventory not found", nil
	// 	}
	// 	return "", err
	// }
	//
	// s, err := i.ToString(ctx)
	// if err != nil {
	// 	return "", err
	// }
	// return fmt.Sprintf("Updated Inventory %s", s), nil
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

func commandStats(ctx context.Context, args []string, userID int) (string, error) {
	c := models.GetConfig(ctx)
	if !c.IsAdmin(userID) {
		return errorMessageNotAdmin, nil
	}

	keys, err := models.ListAllPersonalities(ctx)
	if err != nil {
		return "", err
	}

	ps, err := models.GetPersonalities(ctx, keys)
	if err != nil {
		return "", err
	}

	rs := make([]string, len(ps))
	errs := make([]error, len(ps))
	wg := sync.WaitGroup{}
	wg.Add(len(keys))

	for i := range keys {
		go func(i int) {
			defer wg.Done()
			count, err := models.CountInventories(ctx, keys[i])
			if err == nil {
				rs[i] = fmt.Sprintf("kg:%s %s: %d", ps[i].KGID, ps[i].CanonicalName, count)
			}
			errs[i] = err
		}(i)
	}

	wg.Wait()
	return strings.Join(rs, "\n"), nil
}
