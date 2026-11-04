package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/adexcell/crm-ai/internal/config"
	"github.com/adexcell/crm-ai/internal/infra/crypto"
	"github.com/adexcell/crm-ai/internal/infra/jwt"
	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/adexcell/crm-ai/internal/modules/auth/service"
	"github.com/google/uuid"
)

// mockTxManager эмулирует атомарную транзакцию
type mockTxManager struct {
	mu sync.Mutex
}

func (m *mockTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fn(ctx)
}

// threadSafeRepo эмулирует таблицу refresh_tokens с поддержкой конкуретного чтения/записи
type threadSafeRepo struct {
	mu            sync.RWMutex
	tokens        map[string]*domain.RefreshToken
	user          *domain.User
	member        *domain.OrganizationMember
	killswitchHit bool
}

func newThreadSafeRepo(user *domain.User, member *domain.OrganizationMember) *threadSafeRepo {
	return &threadSafeRepo{
		tokens: make(map[string]*domain.RefreshToken),
		user:   user,
		member: member,
	}
}

func (r *threadSafeRepo) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token.TokenHash] = token
	return nil
}

func (r *threadSafeRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tokens[hash]
	if !ok {
		return nil, nil
	}
	tokenCopy := *t
	return &tokenCopy, nil
}

func (r *threadSafeRepo) RotateRefreshToken(ctx context.Context, oldID, newID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.tokens {
		if t.ID == oldID {
			now := time.Now()
			t.IsRevoked = true
			t.RevokedAt = &now
			t.ReplacedBy = &newID
			return nil
		}
	}
	return nil
}

func (r *threadSafeRepo) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.killswitchHit = true
	for _, t := range r.tokens {
		if t.UserID == userID {
			t.IsRevoked = true
		}
	}
	return nil
}

func (r *threadSafeRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return r.user, nil
}

func (r *threadSafeRepo) GetMember(ctx context.Context, userID uuid.UUID) (*domain.OrganizationMember, error) {
	return r.member, nil
}

func (r *threadSafeRepo) CreateUser(ctx context.Context, user *domain.User) error { return nil }
func (r *threadSafeRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}

func (r *threadSafeRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return nil
}

func (r *threadSafeRepo) UpdateUserProfile(ctx context.Context, userID uuid.UUID, firstName, lastName string) error {
	return nil
}

func (r *threadSafeRepo) GetUserMemberships(ctx context.Context, userID uuid.UUID) ([]domain.UserMembership, error) {
	return nil, nil
}

func (r *threadSafeRepo) CreateOrganization(ctx context.Context, org *domain.Organization) error {
	return nil
}

func (r *threadSafeRepo) GetOrganizationBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	return nil, nil
}
func (r *threadSafeRepo) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error { return nil }
func (r *threadSafeRepo) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]domain.RefreshToken, error) {
	return nil, nil
}

func (r *threadSafeRepo) RevokeUserSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	return nil
}

func (r *threadSafeRepo) AddMember(ctx context.Context, member *domain.OrganizationMember) error {
	return nil
}

func (r *threadSafeRepo) GetMemberByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationMember, error) {
	return nil, nil
}

func TestRefreshToken_ConcurrentGracePeriod(t *testing.T) {
	cfg := &config.Config{
		JWT: jwt.Config{
			SecretKey:              "super-secret-key-32-bytes-minimum!",
			AccessTokenExpireMins:  15,
			RefreshTokenExpireDays: 30,
		},
	}

	jwtManager := jwt.New(cfg.JWT)
	txManager := &mockTxManager{}

	userID := uuid.New()
	orgID := uuid.New()

	testUser := &domain.User{
		ID:       userID,
		Email:    "concurrency_test@acme.com",
		IsActive: true,
	}
	testMember := &domain.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Role:           domain.RoleOwner,
	}

	repo := newThreadSafeRepo(testUser, testMember)
	authService := service.New(repo, nil, txManager, jwtManager, cfg)
	// Создаем исходный валидный токен, который сейчас находится в куках всех 10 вкладок
	rawToken := "test-raw-refresh-token-for-10-tabs"
	tokenHash := crypto.HashToken(rawToken)
	initialToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IsRevoked: false,
		CreatedAt: time.Now(),
	}
	_ = repo.CreateRefreshToken(context.Background(), initialToken)
	// Количество параллельных запросов
	concurrentTabs := 10
	var wg sync.WaitGroup
	startBarrier := make(chan struct{})
	// Буферизированный канал для безопасного сбора ошибок из всех горутин
	errChan := make(chan error, concurrentTabs)
	successCount := 0
	var countMu sync.Mutex // Защищает счетчик успешных ответов от состояния гонки
	for i := 0; i < concurrentTabs; i++ {
		wg.Add(1)
		go func(tabID int) {
			defer wg.Done()
			<-startBarrier
			_, _, err := authService.RefreshToken(
				context.Background(),
				rawToken,
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0",
				"127.0.0.1",
			)
			if err != nil {
				errChan <- err
				return
			}
			countMu.Lock()
			successCount++
			countMu.Unlock()
		}(i)
	}
	close(startBarrier)
	wg.Wait()
	close(errChan)
	// -------------------------------------------------------------------------
	// 4. ПРОВЕРКА ИНВАРИАНТОВ БЕЗОПАСНОСТИ (ASSERTIONS)
	// -------------------------------------------------------------------------
	// Проверяем: не было ли среди ответов ошибки срабатывания Killswitch
	for err := range errChan {
		if errors.Is(err, domain.ErrTokenReplayDetected) {
			t.Fatalf("❌ КАТАСТРОФА: Сработал ложный Replay Attack Killswitch для легитимных вкладок! err: %v", err)
		}
		t.Fatalf("❌ Непредвиденная ошибка при конкурентном рефреше: %v", err)
	}
	// УТВЕРЖДЕНИЕ 1: Все 10 вкладок обязаны получить валидный ответ (200 OK)
	// 1-я вкладка сделала штатную ротацию, а остальные 9 попали в Grace Period (15 сек)
	if successCount != concurrentTabs {
		t.Fatalf("Ожидалось %d успешных ответов, получено: %d", concurrentTabs, successCount)
	}
	// УТВЕРЖДЕНИЕ 2: Killswitch НЕ должен был быть вызван
	repo.mu.RLock()
	if repo.killswitchHit {
		t.Fatalf("❌ Killswitch был активирован, хотя запросы легитимно укладывались в Grace Period!")
	}
	repo.mu.RUnlock()
	t.Logf("✅ УСПЕХ: Все %d вкладок успешно получили токены за Grace Period, 0 гонок данных!", concurrentTabs)
}
