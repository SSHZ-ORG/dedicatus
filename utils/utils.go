package utils

import (
	"crypto/md5"
	"fmt"
)

func TgWebhookPath(token string) string {
	return fmt.Sprintf("/webhook/%x", md5.Sum([]byte(token)))
}
