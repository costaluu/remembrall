package kvstore

import (
	"encoding/json"
	"os"
	"sync"
)

// Store representa nossa KV store temporária
type Store struct {
	filePath string
	mu       sync.RWMutex
	Data     map[string]string `json:"data"`
}

// NewStore cria uma nova instância e carrega dados do disco se existirem
func NewStore(filename string) *Store {
	s := &Store{
		filePath: filename,
		Data:     make(map[string]string),
	}
	s.load()
	return s
}

// Save persiste o estado atual no arquivo JSON
func (s *Store) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bytes, err := json.MarshalIndent(s.Data, "", "  ")
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
	bytes, _ := os.ReadFile(s.filePath)
	json.Unmarshal(bytes, &s.Data)
}

// Set adiciona ou atualiza um registro
func (s *Store) Set(key, value string) {
	s.mu.Lock()
	s.Data[key] = value
	s.mu.Unlock()
	s.save()
}

// Get recupera um valor
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
	s.mu.Unlock()

	// Sobrescreve o arquivo com um objeto vazio ou remove
	return os.WriteFile(s.filePath, []byte("{}"), 0644)
}
