package vcstore

import (
	"sync"
	"time"
)

type VerifiableCredential struct {
	ID        string    `json:"id"`
	Format    string    `json:"format"` // np. "jwt_vc_json"
	Raw       string    `json:"raw"`    // cały JWT albo JSON-LD
	Issuer    string    `json:"issuer"`
	Subject   string    `json:"subject"`
	Types     []string  `json:"types"`
	CreatedAt time.Time `json:"created_at"`
}

type VCStore interface {
	Save(vc VerifiableCredential) error
	Get(id string) (VerifiableCredential, bool)
	List() ([]VerifiableCredential, error)
	Delete(id string) error
	DeleteAll() error
}

// ----- In-memory (MVP / demo) -----

type InMemoryVCStore struct {
	mu sync.RWMutex
	m  map[string]VerifiableCredential
}

func NewInMemoryVCStore() *InMemoryVCStore {
	return &InMemoryVCStore{m: make(map[string]VerifiableCredential)}
}

func (s *InMemoryVCStore) Save(vc VerifiableCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[vc.ID] = vc
	return nil
}

func (s *InMemoryVCStore) Get(id string) (VerifiableCredential, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vc, ok := s.m[id]
	return vc, ok
}

func (s *InMemoryVCStore) List() ([]VerifiableCredential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]VerifiableCredential, 0, len(s.m))
	for _, v := range s.m {
		out = append(out, v)
	}
	return out, nil
}

func (s *InMemoryVCStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, id)
	return nil
}

func (s *InMemoryVCStore) DeleteAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m = make(map[string]VerifiableCredential)
	return nil
}
