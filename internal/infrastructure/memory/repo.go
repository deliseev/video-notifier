package memory

import (
	"context"
	"sync"

	"github.com/deliseev/video-notifier/internal/domain"
)

type InMemoryVideoRepo struct {
	knownIDs    []string
	savedVideos []domain.Video
	mu          sync.RWMutex
}

func NewInMemoryVideoRepo() *InMemoryVideoRepo {
	return &InMemoryVideoRepo{}
}

func (m *InMemoryVideoRepo) GetKnownVideoIDs(ctx context.Context, playlistID string) ([]string, error) {
	// return m.knownIDs, nil
	m.mu.RLock() // Читать могут многие одновременно
	defer m.mu.RUnlock()
	return m.knownIDs, nil
}
func (m *InMemoryVideoRepo) SaveVideos(ctx context.Context, videos []domain.Video) error {
	// m.savedVideos = append(m.savedVideos, videos...)
	// return nil
	m.mu.Lock() // Писать может только один
	defer m.mu.Unlock()

	m.savedVideos = append(m.savedVideos, videos...)
	return nil
}
