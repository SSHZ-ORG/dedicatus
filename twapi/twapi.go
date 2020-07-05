package twapi

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"sync"

	"github.com/ChimeraCoder/anaconda"
	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/models/reservoir"
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

func sendInventoryToTwitter(ctx context.Context, api *anaconda.TwitterApi, i *models.Inventory) error {
	if i.TwitterMediaID == "" {
		return errors.New("why are we sending not-uploaded inventory?")
	}

	pns, err := i.PersonalityNames(ctx)
	if err != nil {
		return err
	}

	for i, pn := range pns {
		pns[i] = "#" + pn
	}
	t, err := api.PostTweet(strings.Join(pns, "\n"), url.Values{"media_ids": []string{i.TwitterMediaID}})
	if err != nil {
		return err
	}

	log.Debugf(ctx, "Sent tweet %s", t.IdStr)
	return nil
}

func SendInventoryToTwitter(ctx context.Context, fileUniqueID string) error {
	api := getClient(ctx)
	if api == nil {
		// Twitter bot not enabled, return.
		return nil
	}

	if fileUniqueID == "" {
		keys, _ := reservoir.ReadReservoir(ctx, 1)
		if len(keys) != 1 {
			return nil
		}
		fileUniqueID = keys[0].StringID()
	}

	log.Debugf(ctx, "Sending %s to Twitter.", fileUniqueID)

	i, err := models.GetInventory(ctx, fileUniqueID)
	if err != nil {
		return err
	}

	if i.FileSize > fileSizeLimit {
		log.Debugf(ctx, "%s is too large for Twitter.", fileUniqueID)
		return nil
	}

	if i.TwitterMediaID == "" {
		log.Debugf(ctx, "We never upload this to Twitter. Uploading now.")

		mediaID, err := uploadInventory(ctx, api, i)
		if err != nil {
			return err
		}

		err = models.SetTwitterMediaID(ctx, fileUniqueID, mediaID)
		if err != nil {
			return err
		}

		i, err = models.GetInventory(ctx, fileUniqueID)
		if err != nil {
			return err
		}
	}

	return sendInventoryToTwitter(ctx, api, i)
}
