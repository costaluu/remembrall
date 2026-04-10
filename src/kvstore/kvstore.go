package kvstore

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var (
	instance *Store
	once     sync.Once
)

func GetInstance(filename string) *Store {
	once.Do(func() {
		instance = NewStore(filename)
	})
	return instance
}

// storeFile é a estrutura persistida no disco
type storeFile struct {
	NextID int               `json:"next_id"`
	Data   map[string]string `json:"data"`
}

// Store representa nossa KV store com chaves auto incrementais
type Store struct {
	filePath string
	mu       sync.RWMutex
	nextID   int
	Data     map[string]string
}

// NewStore cria uma nova instância e carrega dados do disco se existirem
func NewStore(filename string) *Store {
	s := &Store{
		filePath: filename,
		nextID:   1,
		Data:     make(map[string]string),
	}
	s.load()
	return s
}

// save persiste o estado atual no arquivo JSON
func (s *Store) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	payload := storeFile{
		NextID: s.nextID,
		Data:   s.Data,
	}
	bytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, bytes, 0644)
}

// load carrega os dados do arquivo para a memória
func (s *Store) load() {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return
	}
	var payload storeFile
	bytes, _ := os.ReadFile(s.filePath)
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return
	}
	if payload.NextID > 0 {
		s.nextID = payload.NextID
	}
	if payload.Data != nil {
		s.Data = payload.Data
	}
}

// Set adiciona um novo registro com chave auto incremental e retorna a chave gerada
func (s *Store) Set(value string) string {
	s.mu.Lock()
	key := fmt.Sprintf("%d", s.nextID)
	s.Data[key] = value
	s.nextID++
	s.mu.Unlock()

	s.save()
	return key
}

// Get recupera um valor pela chave numérica
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.Data[key]
	return val, ok
}

// Reset limpa tudo do zero, tanto na memória quanto no disco
func (s *Store) Reset() error {
	s.mu.Lock()
	s.Data = make(map[string]string)
	s.nextID = 1
	s.mu.Unlock()

	payload := storeFile{NextID: 1, Data: map[string]string{}}
	bytes, _ := json.MarshalIndent(payload, "", "  ")
	return os.WriteFile(s.filePath, bytes, 0644)
}
