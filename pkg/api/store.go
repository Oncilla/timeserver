// Copyright 2020 oncilla
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/sdomino/scribble"
)

// Info is the API info of a user.
type Info struct {
	ID   string `json:"id"`
	User string `json:"user"`
	Role Role   `json:"role"`
}

// KeyStore persists the API keys.
type KeyStore struct {
	scribble *scribble.Driver
}

// NewKeyStore creates a new API key store.
func NewKeyStore(path string) (*KeyStore, error) {
	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, err
	}
	s, err := scribble.New(path, nil)
	if err != nil {
		return nil, err
	}
	return &KeyStore{
		scribble: s,
	}, nil

}

// Add adds an API key to the store.
func (s *KeyStore) Add(apiKey string, info Info) error {
	resource := fmt.Sprintf("%x", apiKey)
	return s.scribble.Write("api-keys", resource, info)
}

// Get reads the info associated with the user. If the API key does not exist,
// the returned info is nil.
func (s *KeyStore) Get(apiKey string) (*Info, error) {
	var info Info
	resource := fmt.Sprintf("%x", apiKey)
	err := s.scribble.Read("api-keys", resource, &info)
	switch {
	case err == nil:
		return &info, nil
	case errors.Is(err, os.ErrNotExist):
		return nil, nil
	default:
		return nil, err
	}
}

// All returns all API keys in the store.
func (s *KeyStore) All() ([]Info, error) {
	rawKeys, err := s.scribble.ReadAll("api-keys")
	if err != nil {
		return nil, err
	}
	keys := make([]Info, 0, len(rawKeys))
	for _, rawKey := range rawKeys {
		var info Info
		if err := json.Unmarshal(rawKey, &info); err != nil {
			return nil, err
		}
		keys = append(keys, info)
	}
	return keys, nil
}

// Delete deletes an API key. The first return value indicates if the key was deleted.
func (s *KeyStore) Delete(apiKey string) (bool, error) {
	resource := fmt.Sprintf("%x", apiKey)
	info, _ := s.Get(apiKey)
	if info == nil {
		return false, nil
	}
	if err := s.scribble.Delete("api-keys", resource); err != nil {
		return false, err
	}
	return true, nil
}
