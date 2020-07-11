package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/SSHZ-ORG/dedicatus/kgapi"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"github.com/SSHZ-ORG/dedicatus/twapi"
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

var commandMap = map[string]func(ctx context.Context, args []string) (string, error){
	"/start": commandStart,
	"/me":    commandUserInfo,
	"/n":     commandCreatePersonality,
	"/s":     commandFindPersonality,
	"/u":     commandUpdatePersonalityNickname,
	"/a":     commandEditAlias,
	"/c":     commandManageContributors,
	"/stats": commandStats,
	"/tweet": commandTweet,
	"/fo":    commandUpdatePersonalityTwitterUserID,
	"/ukt":   commandUnknownTwitterPersonalities,
}

var complexCommandMap = map[string]func(ctx context.Context, args []string, message *tgbotapi.Message) error{
	"/kg":     commandQueryKG,
	"/sendme": commandSendMe,
}

func makeReplyMessage(message *tgbotapi.Message, reply string) *tgbotapi.MessageConfig {
	c := tgbotapi.NewMessage(message.Chat.ID, reply)
	c.ReplyToMessageID = message.MessageID
	c.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	return &c
}

func HandleMessage(ctx context.Context, update tgbotapi.Update) error {
	bot := tgapi.BotFromContext(ctx)

	message := update.Message
	if message.Animation != nil || (message.ReplyToMessage != nil && message.ReplyToMessage.Animation != nil) {
		return handleAnimation(ctx, message)
	}
	if message.Video != nil || (message.ReplyToMessage != nil && message.ReplyToMessage.Video != nil) {
		return handleVideo(ctx, message)
	}

	if strings.HasPrefix(message.Text, "/") {
		replyMessage := ""
		var err error

		args := strings.Fields(message.Text)

		if handler, ok := commandMap[args[0]]; ok {
			replyMessage, err = handler(ctx, args)
		} else if handler, ok := complexCommandMap[args[0]]; ok {
			err = handler(ctx, args, message)
		} else {
			replyMessage = "Command not recognized."
		}

		if err != nil {
			reply := makeReplyMessage(message, "Your action triggered an internal server error.")
			_, _ = bot.Send(reply) // Fire and forget
			return err
		}

		if replyMessage != "" {
			_, err := bot.Send(makeReplyMessage(message, replyMessage))
			return err
		}
	}

	return nil
}

func handleAnimation(ctx context.Context, message *tgbotapi.Message) error {
	animation := message.Animation
	if message.ReplyToMessage != nil && message.ReplyToMessage.Animation != nil {
		animation = message.ReplyToMessage.Animation
	}

	tgFile := tgapi.TGFileFromChatAnimation(animation)
	return handleTGFile(ctx, message, tgFile)
}

func handleVideo(ctx context.Context, message *tgbotapi.Message) error {
	video := message.Video
	if message.ReplyToMessage != nil && message.ReplyToMessage.Video != nil {
		video = message.ReplyToMessage.Video
	}

	tgFile := tgapi.TGFileFromVideo(video)
	return handleTGFile(ctx, message, tgFile)
}

func handleTGFile(ctx context.Context, message *tgbotapi.Message, tgFile *tgapi.TGFile) error {
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
			return err
		}
	}

	if i != nil {
		// It matched to some existing Inventory, pretend that we received the existing one.
		tgFile.FileUniqueID = i.FileUniqueID
		tgFile.FileID = i.FileID

		s, err := i.ToString(ctx)
		if err != nil {
			return err
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
				return err
			}
			replyMessages = append(replyMessages, "\n"+resp)
		} else {
			if i != nil && i.FileName == "" && tgFile.FileName != "" {
				// Existing Inventory but we don't know FileName, and we know FileName now. Backfill the field.
				if err := models.OverrideFileName(ctx, i.FileUniqueID, tgFile.FileName); err != nil {
					return err
				}
			}
		}
	}

	if len(replyMessages) > 0 {
		_, err := tgapi.BotFromContext(ctx).Send(makeReplyMessage(message, strings.Join(replyMessages, "\n")))
		return err
	}
	return nil
}

func handleTGFileCaption(ctx context.Context, tgFile *tgapi.TGFile, caption string) (string, error) {
	if !tgapi.IsContributor(ctx) {
		return errorMessageNotContributor, nil
	}

	args := strings.Fields(caption)

	if args[0] == "/r" {
		// Use the received document to replace some existing Inventory.
		if !tgapi.IsAdmin(ctx) {
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

func commandStart(ctx context.Context, args []string) (string, error) {
	return fmt.Sprintf("Dedicatus %s", appengine.VersionID(ctx)), nil
}

func commandUserInfo(ctx context.Context, args []string) (string, error) {
	return fmt.Sprintf("User %d\nisAdmin: %v\nisContributor: %v", tgapi.UserFromContext(ctx).ID, tgapi.IsAdmin(ctx), tgapi.IsContributor(ctx)), nil
}

func commandCreatePersonality(ctx context.Context, args []string) (string, error) {
	if !tgapi.IsAdmin(ctx) {
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

func commandFindPersonality(ctx context.Context, args []string) (string, error) {
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

func commandUpdatePersonalityNickname(ctx context.Context, args []string) (string, error) {
	if !tgapi.IsAdmin(ctx) {
		return errorMessageNotAdmin, nil
	}

	if len(args) != 4 || (args[1] != "add" && args[1] != "delete") {
		return "Usage:\n/u add|delete <CanonicalName> <Nickname>\nExample: /u add 井口裕香 知性", nil
	}

	name := args[2]
	alias := utils.NormalizeAlias(args[3])

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

func commandEditAlias(ctx context.Context, args []string) (string, error) {
	if !tgapi.IsAdmin(ctx) {
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

func commandManageContributors(ctx context.Context, args []string) (string, error) {
	if !tgapi.IsAdmin(ctx) {
		return errorMessageNotAdmin, nil
	}

	if len(args) != 3 || (args[1] != "add" && args[1] != "delete") {
		return "Usage:\n/c add|delete <UserId>\nExample: /c add 88888888", nil
	}

	c := tgapi.GetConfig(ctx)

	newContributor, err := strconv.Atoi(args[2])
	if err != nil {
		return "Illegal UserID.", nil
	}

	switch args[1] {
	case "add":
		if c.IsContributor(newContributor) {
			return "Is already a contributor.", nil
		}
		c.Contributors = append(c.Contributors, newContributor)
	case "delete":
		if !c.IsContributor(newContributor) {
			return "Is not a contributor.", nil
		}
		c.Contributors = utils.Remove(c.Contributors, newContributor)
	}

	if err = tgapi.SetConfig(ctx, c); err != nil {
		return "", err
	}

	return "Contributors updated", nil
}

func commandQueryKG(ctx context.Context, args []string, message *tgbotapi.Message) error {
	bot := tgapi.BotFromContext(ctx)

	if !tgapi.IsAdmin(ctx) {
		_, err := bot.Send(makeReplyMessage(message, errorMessageNotAdmin))
		return err
	}

	if len(args) != 2 {
		_, err := bot.Send(makeReplyMessage(message, "Usage:\n/kg <Query>\nExample: /kg 井口裕香"))
		return err
	}

	inputName := args[1]
	encoded, id, name, err := kgapi.GetKGQueryResult(ctx, inputName)
	if err != nil {
		return err
	}

	reply := makeReplyMessage(message, "```json\n"+encoded+"\n```")
	reply.ParseMode = tgbotapi.ModeMarkdownV2
	reply.DisableWebPagePreview = true
	if inputName == name {
		keyboard := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(fmt.Sprintf("/n %s %s", name, id))})
		keyboard.OneTimeKeyboard = true
		keyboard.Selective = true
		reply.ReplyMarkup = keyboard
	}
	_, err = bot.Send(reply)
	return err
}

func commandStats(ctx context.Context, args []string) (string, error) {
	if !tgapi.IsAdmin(ctx) {
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
			count, err := models.CountInventories(ctx, keys[i], false)
			if err == nil && count > 0 {
				rs[i] = fmt.Sprintf("kg:%s %s: %d", ps[i].KGID, ps[i].CanonicalName, count)
			}
			errs[i] = err
		}(i)
	}

	ct, err := models.CountInventories(ctx, nil, false)
	if err != nil {
		return err.Error(), nil
	}
	cut, err := models.CountInventories(ctx, nil, true)
	if err != nil {
		return err.Error(), nil
	}

	wg.Wait()

	out := strings.Builder{}
	for i, r := range rs {
		if errs[i] != nil {
			return errs[i].Error(), nil
		}
		if r != "" {
			out.WriteString(r)
			out.WriteRune('\n')
		}
	}

	out.WriteString(fmt.Sprintf("\nTotal: %d\nUntweeted: %d", ct, cut))

	return out.String(), nil
}

func commandSendMe(ctx context.Context, args []string, message *tgbotapi.Message) error {
	bot := tgapi.BotFromContext(ctx)

	if len(args) != 2 {
		_, err := bot.Send(makeReplyMessage(message, "Usage:\n/sendme <FileUniqueID>"))
		return err
	}

	i, err := models.GetInventory(ctx, args[1])
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			_, err := bot.Send(makeReplyMessage(message, "Unknown inventory"))
			return err
		}
		return err
	}

	return i.SendToChat(ctx, message.Chat.ID)
}

func commandTweet(ctx context.Context, args []string) (string, error) {
	if !tgapi.IsAdmin(ctx) {
		return errorMessageNotAdmin, nil
	}

	if len(args) != 2 {
		return "Usage:\n/tweet <FileUniqueID>", nil
	}

	id, err := twapi.SendInventoryToTwitter(ctx, args[1])
	if err != nil {
		return err.Error(), nil
	}
	return "Posted " + id, nil
}

func commandUpdatePersonalityTwitterUserID(ctx context.Context, args []string) (string, error) {
	if !tgapi.IsAdmin(ctx) {
		return errorMessageNotAdmin, nil
	}

	if len(args) != 3 {
		return "Usage:\n/fo <CanonicalName> <TwitterScreenName>|-", nil
	}

	key, _, err := models.GetPersonalityByName(ctx, args[1])
	if err != nil {
		return "", err
	}
	if key == nil {
		return fmt.Sprintf("Unknown Personality %s", args[1]), nil
	}

	screenName := args[2]
	twitterUserID := ""

	if screenName == models.PersonalityNoTwitterAccountPlaceholder {
		twitterUserID = models.PersonalityNoTwitterAccountPlaceholder
	} else {
		twitterUserID, err = twapi.FollowUser(ctx, screenName)
		if err != nil {
			return err.Error(), nil
		}
	}

	err = models.UpdateTwitterUserID(ctx, key, twitterUserID)
	if err != nil {
		return err.Error(), nil
	}

	return "Set TwitterUserID to " + twitterUserID, nil
}

func commandUnknownTwitterPersonalities(ctx context.Context, args []string) (string, error) {
	if !tgapi.IsAdmin(ctx) {
		return errorMessageNotAdmin, nil
	}

	if len(args) != 2 {
		return "Usage:\n/ukt <Limit>", nil
	}

	limit, err := strconv.Atoi(args[1])
	if err != nil {
		return err.Error(), nil
	}

	is, err := models.LastTweetedInventories(ctx, limit)
	if err != nil {
		return err.Error(), nil
	}

	var pks []*datastore.Key
	for _, i := range is {
		for _, p := range i.Personality {
			if utils.FindKeyIndex(pks, p) == -1 {
				pks = append(pks, p)
			}
		}
	}

	ps, err := models.GetPersonalities(ctx, pks)
	if err != nil {
		return err.Error(), nil
	}

	var os []string
	for _, p := range ps {
		if p.TwitterUserID == "" {
			os = append(os, p.CanonicalName)
		}
	}

	if len(os) == 0 {
		return "(empty)", nil
	}
	return strings.Join(os, ", "), nil
}
