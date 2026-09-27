package detection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// StateTracker defines interface for tracking failed login counters and alert states
type StateTracker interface {
	// IncrementFailedLogin increments failed attempts for (tenantID, key) and returns new count.
	// Sets TTL on the key if it's the first attempt.
	IncrementFailedLogin(ctx context.Context, tenantID, key string, window time.Duration) (int64, error)

	// HasAlertBeenSent checks if an alert has already been generated for this key in the active window.
	HasAlertBeenSent(ctx context.Context, tenantID, key string) (bool, error)

	// SetAlertSent marks that an alert has been generated for this key for the duration of the window.
	SetAlertSent(ctx context.Context, tenantID, key string, ttl time.Duration) error

	// Reset resets the counter and alert state for key
	Reset(ctx context.Context, tenantID, key string) error
}

// RedisStateTracker implements StateTracker using Redis
type RedisStateTracker struct {
	client *redis.Client
}

func NewRedisStateTracker(redisURL string) (*RedisStateTracker, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		opts = &redis.Options{
			Addr: redisURL,
		}
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisStateTracker{client: client}, nil
}

func (r *RedisStateTracker) IncrementFailedLogin(ctx context.Context, tenantID, key string, window time.Duration) (int64, error) {
	redisKey := fmt.Sprintf("brute_force:count:%s:%s", tenantID, key)
	count, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		r.client.Expire(ctx, redisKey, window)
	}

	return count, nil
}

func (r *RedisStateTracker) HasAlertBeenSent(ctx context.Context, tenantID, key string) (bool, error) {
	alertKey := fmt.Sprintf("brute_force:alerted:%s:%s", tenantID, key)
	exists, err := r.client.Exists(ctx, alertKey).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *RedisStateTracker) SetAlertSent(ctx context.Context, tenantID, key string, ttl time.Duration) error {
	alertKey := fmt.Sprintf("brute_force:alerted:%s:%s", tenantID, key)
	return r.client.Set(ctx, alertKey, "1", ttl).Err()
}

func (r *RedisStateTracker) Reset(ctx context.Context, tenantID, key string) error {
	redisKey := fmt.Sprintf("brute_force:count:%s:%s", tenantID, key)
	alertKey := fmt.Sprintf("brute_force:alerted:%s:%s", tenantID, key)
	return r.client.Del(ctx, redisKey, alertKey).Err()
}

// InMemoryStateTracker implements StateTracker in memory for tests or when Redis is absent
type InMemoryStateTracker struct {
	mu       sync.RWMutex
	counters map[string]*counterEntry
	alerted  map[string]time.Time
}

type counterEntry struct {
	count     int64
	expiresAt time.Time
}

func NewInMemoryStateTracker() *InMemoryStateTracker {
	return &InMemoryStateTracker{
		counters: make(map[string]*counterEntry),
		alerted:  make(map[string]time.Time),
	}
}

func (m *InMemoryStateTracker) IncrementFailedLogin(ctx context.Context, tenantID, key string, window time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fullKey := fmt.Sprintf("%s:%s", tenantID, key)
	now := time.Now()

	entry, exists := m.counters[fullKey]
	if !exists || now.After(entry.expiresAt) {
		m.counters[fullKey] = &counterEntry{
			count:     1,
			expiresAt: now.Add(window),
		}
		return 1, nil
	}

	entry.count++
	return entry.count, nil
}

func (m *InMemoryStateTracker) HasAlertBeenSent(ctx context.Context, tenantID, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fullKey := fmt.Sprintf("%s:%s", tenantID, key)
	exp, exists := m.alerted[fullKey]
	if !exists {
		return false, nil
	}

	if time.Now().After(exp) {
		return false, nil
	}

	return true, nil
}

func (m *InMemoryStateTracker) SetAlertSent(ctx context.Context, tenantID, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fullKey := fmt.Sprintf("%s:%s", tenantID, key)
	m.alerted[fullKey] = time.Now().Add(ttl)
	return nil
}

func (m *InMemoryStateTracker) Reset(ctx context.Context, tenantID, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fullKey := fmt.Sprintf("%s:%s", tenantID, key)
	delete(m.counters, fullKey)
	delete(m.alerted, fullKey)
	return nil
}
