package zkprequest

import "sync"

// RequestStore persists pending requests until verified/expired.
type RequestStore interface {
	Save(req PresentationRequest) error
	Load(id string) (PresentationRequest, bool)
	Delete(id string)
}

// InMemoryStore is fine for demos; replace with Redis/DB in prod.
type InMemoryStore struct {
	m sync.Map // id -> PresentationRequest
}

func (s *InMemoryStore) Save(req PresentationRequest) error {
	s.m.Store(req.RequestID, req)
	return nil
}

func (s *InMemoryStore) Load(id string) (PresentationRequest, bool) {
	v, ok := s.m.Load(id)
	if !ok {
		return PresentationRequest{}, false
	}
	return v.(PresentationRequest), true
}

func (s *InMemoryStore) Delete(id string) { s.m.Delete(id) }
