package mutex

import (
	"hash/fnv"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis"
)

const (
	minNodeNum     = 0
	maxNodeNum     = 1023
	defaultRedisDB = 0
)

const (
	// unlockScript is redis lua script to release a lock.
	// it guarantees that getting key and deleting key is in a transaction
	unlockScript = `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
	`
)

// Client is used to produce mutex
type Client struct {
	redis     *redis.Client
	redisOpts *redis.Options
	// nodeNum is a uniq number of current machine in cluster, which used to generate uuid
	nodeNum int64
	// node is used to generate uuid
	node *snowflake.Node
}

type ClientOption func(c *Client)

func SetRedisDB(db int) ClientOption {
	return func(c *Client) {
		c.redisOpts.DB = db
	}
}

func SetRedisPool(size int) ClientOption {
	return func(c *Client) {
		c.redisOpts.PoolSize = size
	}
}

func SetRedisPass(pass string) ClientOption {
	return func(c *Client) {
		c.redisOpts.Password = pass
	}
}

// SetNodeNum sets the node number by user but no the defaultNodeNum, which is recommended
func SetNodeNum(node int64) ClientOption {
	return func(c *Client) {
		if node >= minNodeNum && node <= maxNodeNum {
			c.nodeNum = node
		}
	}
}

func NewClient(addr string, opts ...ClientOption) *Client {
	cli := &Client{
		redisOpts: &redis.Options{
			Addr: addr,
			DB:   defaultRedisDB,
		},
		nodeNum: defaultNodeNum(),
	}
	for _, opt := range opts {
		opt(cli)
	}

	node, err := snowflake.NewNode(cli.nodeNum)
	if err != nil {
		log.Fatal(err)
	}
	cli.node = node

	redisCli := redis.NewClient(cli.redisOpts)
	if _, err := redisCli.Ping().Result(); err != nil {
		log.Fatal(err)
	}
	cli.redis = redisCli

	return cli
}

// NewMutex creates a mutex with locking key and expire time
func (c *Client) NewMutex(key string, expiration time.Duration) *mutex {
	return &mutex{
		redis:      c.redis,
		key:        key,
		value:      c.node.Generate().Int64(),
		expiration: expiration,
	}
}

// defaultNodeNum generates a node num between [0-1023] by host name
// NOTE: it cannot guarantee that the node num is unique in cluster，
// which may cause generated uuid not to be unique too, but the probability is extremely low.
func defaultNodeNum() int64 {
	name, _ := os.Hostname()
	h := fnv.New32a()
	h.Write([]byte(name))
	nodeNum := h.Sum32() % (maxNodeNum + 1)

	return int64(nodeNum)
}

type mutex struct {
	redis      *redis.Client
	key        string
	value      int64
	expiration time.Duration
}

func (m *mutex) Lock() bool {
	return m.redis.SetNX(m.key, m.value, m.expiration).Val()
}

func (m *mutex) Unlock() {
	m.redis.Eval(unlockScript, []string{m.key}, m.value)
}
