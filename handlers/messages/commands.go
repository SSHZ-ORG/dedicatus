package messages

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/SSHZ-ORG/dedicatus/dctx"
	"github.com/SSHZ-ORG/dedicatus/dctx/protoconf"
	"github.com/SSHZ-ORG/dedicatus/dctx/protoconf/pb"
	"github.com/SSHZ-ORG/dedicatus/kgapi"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/twapi"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/protobuf/encoding/prototext"
)

var commandMap = map[string]func(ctx context.Context, args []string) (string, error){
	"/start": commandStart,
	"/me":    commandUserInfo,
	"/n":     commandCreatePersonality,
	"/s":     commandFindPersonality,
	"/u":     commandUpdatePersonalityNickname,
	"/a":     commandEditAlias,
	"/stats": commandStats,
	"/tweet": commandTweet,
	"/fo":    commandUpdatePersonalityTwitterUserID,
	"/ukt":   commandUnknownTwitterPersonalities,
	"/c":     commandConfig,
}

var complexCommandMap = map[string]func(ctx context.Context, args []string, message *tgbotapi.Message) (tgbotapi.Chattable, error){
	"/kg":     commandQueryKG,
	"/sendme": commandSendMe,
}

func commandStart(ctx context.Context, args []string) (string, error) {
	return fmt.Sprintf("Dedicatus %s", appengine.VersionID(ctx)), nil
}

func commandUserInfo(ctx context.Context, args []string) (string, error) {
	return fmt.Sprintf("User %d\nType: %v", dctx.UserFromContext(ctx).ID, dctx.UserTypeFromContext(ctx)), nil
}

func commandCreatePersonality(ctx context.Context, args []string) (string, error) {
	if !dctx.IsAdmin(ctx) {
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
	if !dctx.IsAdmin(ctx) {
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
	if !dctx.IsAdmin(ctx) {
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

func commandQueryKG(ctx context.Context, args []string, message *tgbotapi.Message) (tgbotapi.Chattable, error) {
	if !dctx.IsAdmin(ctx) {
		return makeReplyMessage(message, errorMessageNotAdmin), nil
	}

	if len(args) != 2 {
		return makeReplyMessage(message, "Usage:\n/kg <Query>\nExample: /kg 井口裕香"), nil
	}

	inputName := args[1]
	encoded, id, name, err := kgapi.GetKGQueryResult(ctx, inputName)
	if err != nil {
		return nil, err
	}

	reply := makeReplyMessage(message, "```json\n"+encoded+"\n```")
	reply.ParseMode = tgbotapi.ModeMarkdownV2
	if inputName == name {
		keyboard := tgbotapi.NewOneTimeReplyKeyboard([]tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(fmt.Sprintf("/n %s %s", name, id))})
		keyboard.Selective = true
		reply.ReplyMarkup = keyboard
	}
	return reply, nil
}

func commandStats(ctx context.Context, args []string) (string, error) {
	if !dctx.IsAdmin(ctx) {
		return errorMessageNotAdmin, nil
	}

	var keys []*datastore.Key
	var err error

	if len(args) > 1 {
		keys, err = models.ListAllPersonalities(ctx)
		if err != nil {
			return "", err
		}
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

func commandSendMe(ctx context.Context, args []string, message *tgbotapi.Message) (tgbotapi.Chattable, error) {
	if len(args) != 2 {
		return makeReplyMessage(message, "Usage:\n/sendme <FileUniqueID>"), nil
	}

	i, err := models.GetInventory(ctx, args[1])
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return makeReplyMessage(message, "Unknown inventory"), nil
		}
		return nil, err
	}

	return i.SendToChat(ctx, message.Chat.ID)
}

func commandTweet(ctx context.Context, args []string) (string, error) {
	if !dctx.IsAdmin(ctx) {
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
	if !dctx.IsAdmin(ctx) {
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
	if !dctx.IsAdmin(ctx) {
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

func commandConfig(ctx context.Context, args []string) (string, error) {
	if !dctx.IsAdmin(ctx) {
		return errorMessageNotAdmin, nil
	}

	usage := "Usage:\n/c get|set|auth"

	var c *pb.Protoconf
	var err error

	if len(args) < 2 {
		return usage, nil
	}

	switch args[1] {
	case "get":
		c = dctx.ProtoconfFromContext(ctx)
	case "set":
		if len(args) != 4 {
			return "Usage:\n/c set <Key> <Value>", nil
		}
		c, err = protoconf.EditConf(ctx, args[2], args[3])
	case "auth":
		if len(args) != 4 {
			return "Usage:\n/c auth <UserID> <UserType>", nil
		}
		c, err = protoconf.SetUserType(ctx, args[2], args[3])
	default:
		return usage, nil
	}

	if err != nil {
		return err.Error(), nil
	}
	m := prototext.Format(c)
	if m == "" {
		return "(empty)", nil
	}
	return m, nil
}
