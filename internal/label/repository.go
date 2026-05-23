package label

import (
	"context"
	"sync"
)

type MemoryLabelRepository struct {
	mu     sync.RWMutex
	labels map[string]*Label
}

// NewMemoryLabelRepository instantiates a isolated in-memory database for Labels.
func NewMemoryLabelRepository() *MemoryLabelRepository {
	return &MemoryLabelRepository{
		labels: make(map[string]*Label),
	}
}

func (r *MemoryLabelRepository) Create(ctx context.Context, label *Label) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already exists by ID
	if _, exists := r.labels[label.ID]; exists {
		return ErrLabelAlreadyCancelled // or generic already exists
	}

	// Check if already exists by tracking number
	for _, l := range r.labels {
		if l.TrackingNumber == label.TrackingNumber {
			return ErrLabelAlreadyCancelled
		}
	}

	r.labels[label.ID] = label
	return nil
}

func (r *MemoryLabelRepository) GetByID(ctx context.Context, id string) (*Label, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	label, exists := r.labels[id]
	if !exists {
		return nil, ErrLabelNotFound
	}

	cp := *label
	return &cp, nil
}

func (r *MemoryLabelRepository) GetByTracking(ctx context.Context, trackingNum string) (*Label, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, l := range r.labels {
		if l.TrackingNumber == trackingNum {
			cp := *l
			return &cp, nil
		}
	}

	return nil, ErrLabelNotFound
}

func (r *MemoryLabelRepository) Update(ctx context.Context, label *Label) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.labels[label.ID]; !exists {
		return ErrLabelNotFound
	}

	r.labels[label.ID] = label
	return nil
}

func (r *MemoryLabelRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.labels[id]; !exists {
		return ErrLabelNotFound
	}

	delete(r.labels, id)
	return nil
}
