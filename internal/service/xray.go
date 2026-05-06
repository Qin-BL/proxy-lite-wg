package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"

	"proxy-lite-wg/internal/config"
	"proxy-lite-wg/internal/domain"
	"proxy-lite-wg/internal/store"
)

type Repository interface {
	CreateUser(ctx context.Context, user domain.User) error
	ListUsers(ctx context.Context) ([]domain.User, error)
	GetUserByID(ctx context.Context, id string) (domain.User, error)
	DeleteUser(ctx context.Context, id string) error
	CountClientsByUser(ctx context.Context, userID string) (int, error)
	SaveClient(ctx context.Context, client domain.Client) error
	ListClients(ctx context.Context) ([]domain.Client, error)
	ListActiveClients(ctx context.Context) ([]domain.Client, error)
	GetClientByID(ctx context.Context, id string) (domain.Client, error)
	DeleteClient(ctx context.Context, id string) error
}

type XrayService struct {
	cfg        config.Config
	repo       Repository
	docker     *client.Client
	restartXray bool
}

func NewXrayService(cfg config.Config, repo Repository) (*XrayService, error) {
	svc := &XrayService{
		cfg:         cfg,
		repo:        repo,
		restartXray: strings.TrimSpace(cfg.XrayContainerName) != "",
	}

	if svc.restartXray {
		host := cfg.DockerSocketPath
		if !strings.Contains(host, "://") {
			host = "unix://" + host
		}
		dockerClient, err := client.NewClientWithOpts(
			client.WithHost(host),
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			return nil, fmt.Errorf("create docker client: %w", err)
		}
		svc.docker = dockerClient
	}

	return svc, nil
}

func (s *XrayService) Close() error {
	if s.docker != nil {
		return s.docker.Close()
	}
	return nil
}

func (s *XrayService) Initialize(ctx context.Context) error {
	return s.syncRuntime(ctx)
}

func (s *XrayService) CreateUser(ctx context.Context, name, email string) (domain.User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if name == "" {
		return domain.User{}, errors.New("name is required")
	}
	if email == "" {
		return domain.User{}, errors.New("email is required")
	}

	now := time.Now().UTC()
	user := domain.User{
		ID:        uuid.NewString(),
		Name:      name,
		Email:     email,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *XrayService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.ListUsers(ctx)
}

func (s *XrayService) DeleteUser(ctx context.Context, id string) error {
	if _, err := s.repo.GetUserByID(ctx, strings.TrimSpace(id)); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	count, err := s.repo.CountClientsByUser(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("delete this user's clients first")
	}

	if err := s.repo.DeleteUser(ctx, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return errors.New("user not found")
		}
		return err
	}
	return nil
}

func (s *XrayService) CreateClient(ctx context.Context, userID, label string) (domain.Client, error) {
	if _, err := s.repo.GetUserByID(ctx, strings.TrimSpace(userID)); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return domain.Client{}, errors.New("user not found")
		}
		return domain.Client{}, err
	}

	label = strings.TrimSpace(label)
	if label == "" {
		return domain.Client{}, errors.New("label is required")
	}

	now := time.Now().UTC()
	clientRecord := domain.Client{
		ID:         uuid.NewString(),
		UserID:     strings.TrimSpace(userID),
		Label:      label,
		State:      "active",
		ClientUUID: uuid.NewString(),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.SaveClient(ctx, clientRecord); err != nil {
		return domain.Client{}, err
	}
	if err := s.syncRuntime(ctx); err != nil {
		_ = s.repo.DeleteClient(context.Background(), clientRecord.ID)
		return domain.Client{}, err
	}
	return s.decorateClient(clientRecord), nil
}

func (s *XrayService) ListClients(ctx context.Context) ([]domain.Client, error) {
	clients, err := s.repo.ListClients(ctx)
	if err != nil {
		return nil, err
	}
	for i := range clients {
		clients[i] = s.decorateClient(clients[i])
	}
	return clients, nil
}

func (s *XrayService) GetClient(ctx context.Context, id string) (domain.Client, error) {
	clientRecord, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return domain.Client{}, errors.New("client not found")
		}
		return domain.Client{}, err
	}
	return s.decorateClient(clientRecord), nil
}

func (s *XrayService) DisableClient(ctx context.Context, id string) error {
	current, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return errors.New("client not found")
		}
		return err
	}
	if current.State == "disabled" {
		return nil
	}

	now := time.Now().UTC()
	current.State = "disabled"
	current.DisabledAt = &now
	current.UpdatedAt = now
	if err := s.repo.SaveClient(ctx, current); err != nil {
		return err
	}
	if err := s.syncRuntime(ctx); err != nil {
		current.State = "active"
		current.DisabledAt = nil
		current.UpdatedAt = time.Now().UTC()
		_ = s.repo.SaveClient(context.Background(), current)
		return err
	}
	return nil
}

func (s *XrayService) EnableClient(ctx context.Context, id string) error {
	current, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return errors.New("client not found")
		}
		return err
	}
	if current.State == "active" {
		return nil
	}

	current.State = "active"
	current.DisabledAt = nil
	current.UpdatedAt = time.Now().UTC()
	if err := s.repo.SaveClient(ctx, current); err != nil {
		return err
	}
	if err := s.syncRuntime(ctx); err != nil {
		now := time.Now().UTC()
		current.State = "disabled"
		current.DisabledAt = &now
		current.UpdatedAt = now
		_ = s.repo.SaveClient(context.Background(), current)
		return err
	}
	return nil
}

func (s *XrayService) DeleteClient(ctx context.Context, id string) error {
	current, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return errors.New("client not found")
		}
		return err
	}

	if err := s.repo.DeleteClient(ctx, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return errors.New("client not found")
		}
		return err
	}
	if err := s.syncRuntime(ctx); err != nil {
		_ = s.repo.SaveClient(context.Background(), current)
		return err
	}
	return nil
}

func (s *XrayService) RenderShareLink(ctx context.Context, id string) (string, error) {
	clientRecord, err := s.GetClient(ctx, id)
	if err != nil {
		return "", err
	}
	return clientRecord.ShareLink, nil
}

func (s *XrayService) decorateClient(clientRecord domain.Client) domain.Client {
	clientRecord.ShareLink = s.renderShareLink(clientRecord)
	return clientRecord
}

func (s *XrayService) renderShareLink(clientRecord domain.Client) string {
	query := url.Values{}
	query.Set("encryption", "none")
	query.Set("security", "tls")
	query.Set("sni", s.cfg.PublicHost)
	query.Set("type", "ws")
	query.Set("host", s.cfg.PublicHost)
	query.Set("path", s.cfg.VLESSWSPath)

	name := sanitizeNodeName(clientRecord.Label)
	return fmt.Sprintf(
		"vless://%s@%s:%d?%s#%s",
		clientRecord.ClientUUID,
		s.cfg.PublicHost,
		s.cfg.PublicPort,
		query.Encode(),
		url.QueryEscape(name),
	)
}

func sanitizeNodeName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "proxy-lite-client"
	}
	return value
}

func (s *XrayService) syncRuntime(ctx context.Context) error {
	clients, err := s.repo.ListActiveClients(ctx)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(s.cfg.XrayConfigPath), 0o755); err != nil {
		return fmt.Errorf("create xray config dir: %w", err)
	}

	configBytes, err := json.MarshalIndent(s.buildXrayConfig(clients), "", "  ")
	if err != nil {
		return fmt.Errorf("marshal xray config: %w", err)
	}

	if err := writeFileAtomic(s.cfg.XrayConfigPath, configBytes, 0o644); err != nil {
		return fmt.Errorf("write xray config: %w", err)
	}

	if s.restartXray {
		timeout := 10
		if err := s.docker.ContainerRestart(ctx, s.cfg.XrayContainerName, container.StopOptions{Timeout: &timeout}); err != nil {
			return fmt.Errorf("restart xray container: %w", err)
		}
	}

	return nil
}

func (s *XrayService) buildXrayConfig(clients []domain.Client) map[string]any {
	xrayClients := make([]map[string]any, 0, len(clients))
	for _, clientRecord := range clients {
		xrayClients = append(xrayClients, map[string]any{
			"id":    clientRecord.ClientUUID,
			"email": fmt.Sprintf("%s:%s", clientRecord.ID, clientRecord.Label),
		})
	}

	return map[string]any{
		"log": map[string]any{
			"loglevel": s.cfg.XrayLogLevel,
		},
		"inbounds": []map[string]any{
			{
				"listen":   "0.0.0.0",
				"port":     s.cfg.XrayInboundPort,
				"protocol": "vless",
				"settings": map[string]any{
					"clients":    xrayClients,
					"decryption": "none",
				},
				"streamSettings": map[string]any{
					"network":  "ws",
					"security": "none",
					"wsSettings": map[string]any{
						"path": s.cfg.VLESSWSPath,
					},
				},
				"sniffing": map[string]any{
					"enabled": true,
					"destOverride": []string{
						"http",
						"tls",
						"quic",
					},
				},
			},
		},
		"outbounds": []map[string]any{
			{
				"protocol": "freedom",
				"tag":      "direct",
			},
		},
	}
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, append(data, '\n'), perm); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
