// Package userprofile реализует минимальный клиент для эндпоинта
// /admin/realms/{realm}/users/profile, которого нет в gocloak v13.
// Структуры повторяют модель Keycloak (UPConfig/UPAttribute/...).
package userprofile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Permissions описывает, кто может читать/редактировать атрибут.
type Permissions struct {
	View []string `json:"view"`
	Edit []string `json:"edit"`
}

// Required описывает, для каких ролей атрибут обязателен.
type Required struct {
	Roles  []string `json:"roles,omitempty"`
	Scopes []string `json:"scopes,omitempty"`
}

// Attribute — описание одного атрибута пользовательского профиля.
type Attribute struct {
	Name        string         `json:"name"`
	DisplayName string         `json:"displayName,omitempty"`
	Permissions Permissions    `json:"permissions"`
	Required    *Required      `json:"required,omitempty"`
	Multivalued bool           `json:"multivalued,omitempty"`
	Validations map[string]any `json:"validations,omitempty"`
}

// Config — корневая конфигурация User Profile.
type Config struct {
	Attributes []Attribute `json:"attributes"`
}

// AttributeNames возвращает множество имён атрибутов текущей конфигурации.
func (c *Config) AttributeNames() map[string]struct{} {
	names := make(map[string]struct{}, len(c.Attributes))
	for _, a := range c.Attributes {
		names[a.Name] = struct{}{}
	}
	return names
}

// Client — тонкий HTTP-клиент эндпоинта User Profile.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient создаёт клиент. baseURL — адрес Keycloak без завершающего слэша.
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL, HTTPClient: http.DefaultClient}
}

func (c *Client) url(realm string) string {
	return fmt.Sprintf("%s/admin/realms/%s/users/profile", c.BaseURL, realm)
}

// Get возвращает текущую конфигурацию User Profile realm'а.
func (c *Client) Get(ctx context.Context, token, realm string) (*Config, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(realm), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить user profile: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user profile: неожиданный статус %d: %s", resp.StatusCode, string(body))
	}

	var cfg Config
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, fmt.Errorf("не удалось разобрать ответ user profile: %w", err)
	}
	return &cfg, nil
}

// Update полностью заменяет конфигурацию User Profile realm'а.
func (c *Client) Update(ctx context.Context, token, realm string, cfg *Config) error {
	payload, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url(realm), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("не удалось обновить user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update user profile: неожиданный статус %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
