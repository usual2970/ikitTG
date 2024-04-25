package audio

import (
	"context"
	"fmt"
	xhttp "ikit-api/internal/util/http"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

var xml = `
<speak version='1.0' xml:lang='en-US'><voice xml:lang='en-US' xml:gender='Male'
name='en-US-AndrewMultilingualNeural'>
	%s
</voice></speak>
`

var cacheAzure *expirable.LRU[string, string]

var cacheOnceAzure sync.Once

func newAzureTokenCache() *expirable.LRU[string, string] {

	cacheOnceAzure.Do(func() {
		cacheAzure = expirable.NewLRU[string, string](5, nil, time.Minute*10)
	})

	return cacheAzure
}

func Azure(ctx context.Context, text string) ([]byte, error) {
	token, err := getToken(ctx)
	if err != nil {
		return nil, err
	}

	body := fmt.Sprintf(xml, text)

	resp, err := xhttp.Req("https://eastus.tts.speech.microsoft.com/cognitiveservices/v1", http.MethodPost, strings.NewReader(body), map[string]string{
		"Authorization":            "Bearer " + token,
		"Content-Type":             "application/ssml+xml",
		"X-Microsoft-OutputFormat": "audio-24khz-48kbitrate-mono-mp3",
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func getToken(_ context.Context) (string, error) {

	cache := newAzureTokenCache()

	if rs, ok := cache.Get("token"); ok {
		return rs, nil
	}
	speechKey := os.Getenv("AZURE_SPEECH_KEY")
	resp, err := xhttp.Req("https://eastus.api.cognitive.microsoft.com/sts/v1.0/issueToken", http.MethodPost, nil, map[string]string{
		"Ocp-Apim-Subscription-Key": speechKey,
	})

	if err != nil {
		return "", err
	}

	rs := string(resp)
	cache.Add("token", rs)

	return rs, nil
}
