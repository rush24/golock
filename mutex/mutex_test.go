package mutex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasicLock(t *testing.T) {
	cli := NewClient(
		"127.0.0.1:6379",
		SetRedisPool(2),
		SetNodeNum(1))
	key := "test_key"
	m := cli.NewMutex(key, time.Second)
	m1 := cli.NewMutex(key, time.Second)

	assert.True(t, m.Lock())
	assert.False(t, m1.Lock())
	assert.NotEmpty(t, cli.redis.Get(key).Val())
	m.Unlock()
	assert.Empty(t, cli.redis.Get(key).Val())
	assert.True(t, m1.Lock())
	assert.NotEmpty(t, cli.redis.Get(key).Val())
	m1.Unlock()
	assert.Empty(t, cli.redis.Get(key).Val())
}

func TestSafeUnlock(t *testing.T) {
	cli := NewClient(
		"127.0.0.1:6379",
		SetRedisPool(2),
		SetNodeNum(1))
	key := "test_key"
	m := cli.NewMutex(key, time.Second)
	m.Lock()
	// simulate key expiration
	cli.redis.Del(key)
	m1 := cli.NewMutex(key, time.Second)
	assert.True(t, m1.Lock())
	m.Unlock()
	assert.NotEmpty(t, cli.redis.Get(key).Val())
	m1.Unlock()
	assert.Empty(t, cli.redis.Get(key).Val())
}
