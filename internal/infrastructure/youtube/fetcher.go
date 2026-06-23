package youtube

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/deliseev/video-notifier/internal/domain"
)

type YoutubeFetcher struct {
	client *http.Client
}

func NewYoutubeFetcher() *YoutubeFetcher {
	return &YoutubeFetcher{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// FetchLatestVideos реализует интерфейс VideoFetcher из пакета usecase
func (f *YoutubeFetcher) doFetch(ctx context.Context, playlistID string) ([]domain.Video, error) {
	url := fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?playlist_id=%s", playlistID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	// Тут будет магия XML парсинга
	videos, err := f.parseRSS(resp.Body)
	if err != nil {
		return nil, err
	}
	// Добавить PlaylistID
	for i := range videos {
		videos[i].PlaylistID = playlistID
	}
	return videos, nil
}

func (f *YoutubeFetcher) FetchLatestVideos(ctx context.Context, playlistID string) ([]domain.Video, error) {
	var lastErr error
	for i := 0; i < 3; i++ { // Попробуем 3 раза
		videos, err := f.doFetch(ctx, playlistID)
		if err == nil {
			return videos, nil
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1)) // Экспоненциальная задержка
	}
	return nil, lastErr
}

type rssFeed struct {
	Entries []rssEntry `xml:"entry"`
}

type rssEntry struct {
	ID    string `xml:"id"`
	Title string `xml:"title"`
	Link  struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Published time.Time `xml:"published"`
}

func (f *YoutubeFetcher) parseRSS(r io.Reader) ([]domain.Video, error) {
	var feed rssFeed
	if err := xml.NewDecoder(r).Decode(&feed); err != nil {
		return nil, err
	}

	var videos []domain.Video
	for _, entry := range feed.Entries {
		videos = append(videos, domain.Video{
			ID:        entry.ID, // Нужно будет отрезать "yt:video:"
			Title:     entry.Title,
			URL:       entry.Link.Href,
			Published: entry.Published,
		})
	}
	return videos, nil
}
