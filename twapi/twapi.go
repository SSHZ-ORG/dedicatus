package twapi

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"

	"github.com/ChimeraCoder/anaconda"
	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

func getClient(ctx context.Context) *anaconda.TwitterApi {
	if config.TwitterAccessToken == "" {
		return nil
	}
	return anaconda.NewTwitterApiWithCredentials(config.TwitterAccessToken, config.TwitterAccessTokenSecret, config.TwitterAPIKey, config.TwitterAPIKeySecret)
}

const fileSizeLimit = 15 * 1024 * 1024
const chunkSizeLimit = 3 * 1024 * 1024 // Twitter allows 5MB after base64 encoding, technically we can do 5 * 3 / 4 MB.

func uploadInventory(ctx context.Context, api *anaconda.TwitterApi, i *models.Inventory) (string, error) {
	_, fileBytes, err := tgapi.FetchFileInfo(ctx, i.FileID)
	if err != nil {
		return "", err
	}

	var chunk []byte
	chunks := make([][]byte, 0, len(fileBytes)/chunkSizeLimit+1)
	for len(fileBytes) >= chunkSizeLimit {
		chunk, fileBytes = fileBytes[:chunkSizeLimit], fileBytes[chunkSizeLimit:]
		chunks = append(chunks, chunk)
	}
	if len(fileBytes) > 0 {
		chunks = append(chunks, fileBytes[:len(fileBytes)])
	}

	media, err := api.UploadVideoInit(i.FileSize, "video/mp4") // Currently we only support this.
	if err != nil {
		return "", err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(chunks))
	errs := make([]error, len(chunks))
	for i := range chunks {
		go func(i int) {
			defer wg.Done()
			errs[i] = api.UploadVideoAppend(media.MediaIDString, i, base64.StdEncoding.EncodeToString(chunks[i]))
		}(i)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return "", err
		}
	}
	m, err := api.UploadVideoFinalize(media.MediaIDString)
	if err != nil {
		return "", err
	}

	log.Debugf(ctx, "Uploaded %s to Twitter as media_id %s", i.FileUniqueID, m.MediaIDString)
	return m.MediaIDString, nil
}

func postTweet(ctx context.Context, api *anaconda.TwitterApi, i *models.Inventory) (string, error) {
	if i.TwitterMediaID == "" {
		return "", errors.New("why are we sending not-uploaded inventory?")
	}

	pns, err := i.PersonalityNames(ctx)
	if err != nil {
		return "", err
	}

	for i, pn := range pns {
		pns[i] = "#" + pn
	}
	t, err := api.PostTweet(strings.Join(pns, "\n"), url.Values{"media_ids": []string{i.TwitterMediaID}})
	if err != nil {
		return "", err
	}

	log.Debugf(ctx, "Sent tweet %s", t.IdStr)

	_ = models.UpdateLastTweetInfo(ctx, i.FileUniqueID, t.IdStr)

	return fmt.Sprintf("https://twitter.com/%s/status/%s", t.User.ScreenName, t.IdStr), nil
}

const (
	leastRecentProb        = 0.05
	leastRecentOffsetRange = 50
	standardPoolLimit      = 5
)

func pickRandomInventory(ctx context.Context) (*models.Inventory, error) {
	if rand.Float32() < leastRecentProb {
		log.Infof(ctx, "Won the lottery! Choosing something that was not tweeted recently...")
		return models.PickLeastRecentTweetedInventory(ctx, rand.Intn(leastRecentOffsetRange))
	}

	is, err := models.RandomInventories(ctx, standardPoolLimit)
	if err != nil {
		return nil, err
	}

	if len(is) == 0 {
		// Likely it's Cache Miss. Return error to let cron try again.
		return nil, errors.New("received 0 random Inventories back")
	}

	// Choose the first Inventory that was never posted before.
	// If all were posted before, choose the one that was least recently posted.
	var li = is[0]
	for _, i := range is {
		if i.LastTweetTime.IsZero() {
			return i, nil
		}
		if i.LastTweetTime.Before(li.LastTweetTime) {
			li = i
		}
	}
	return li, nil
}

func SendInventoryToTwitter(ctx context.Context, manualFileUniqueId string) (string, error) {
	api := getClient(ctx)
	if api == nil {
		// Twitter bot not enabled, return.
		return "", nil
	}

	var i *models.Inventory
	var err error

	if manualFileUniqueId == "" {
		i, err = pickRandomInventory(ctx)
	} else {
		i, err = models.GetInventory(ctx, manualFileUniqueId)
	}
	if err != nil {
		return "", err
	}

	log.Debugf(ctx, "Sending %s to Twitter.", i.FileUniqueID)

	if i.FileSize > fileSizeLimit {
		log.Debugf(ctx, "%s is too large for Twitter.", i.FileUniqueID)
		if manualFileUniqueId != "" {
			return "", errors.New("file too large")
		}
		return "", nil
	}

	if i.TwitterMediaID == "" {
		log.Debugf(ctx, "We never upload this to Twitter. Uploading now.")

		mediaID, err := uploadInventory(ctx, api, i)
		if err != nil {
			return "", err
		}

		i, err = models.SetTwitterMediaID(ctx, i.FileUniqueID, mediaID)
		if err != nil {
			return "", err
		}
	}

	return postTweet(ctx, api, i)
}

func FollowUser(ctx context.Context, screenName string) (userID string, err error) {
	api := getClient(ctx)

	u, err := api.FollowUser(screenName)
	if err != nil {
		return "", err
	}

	return u.IdStr, nil
}
