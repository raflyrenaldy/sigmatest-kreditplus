package redis

import (
	"context"
	"fmt"
	"time"

	"user/sigmatech/app/constants"
	"user/sigmatech/app/service/logger"

	"github.com/redis/go-redis/v9"
)

type IRedisClient interface {
	Ping() error
	Close() error
	Set(key, value string, expiration time.Duration) error
	Get(key string) (string, error)
	Delete(keys ...string) (int64, error)
	Exists(key string) (bool, error)
	HMSet(key string, fieldsAndValues map[string]interface{}) error
	HGetAll(key string) (map[string]string, error)
	Publish(channel, message string) error
	LPush(key string, values ...interface{}) error
	RPush(key string, values ...interface{}) error
	LPop(key string) (string, error)
	RPop(key string) (string, error)
	SAdd(key string, members ...interface{}) (int64, error)
	SRem(key string, members ...interface{}) (int64, error)
	SMembers(key string) ([]string, error)
	SIsMember(key string, member interface{}) (bool, error)
	ZAdd(key string, members ...redis.Z) (int64, error)
	ZRange(key string, start, stop int64) ([]string, error)
	ZScore(key string, member string) (float64, error)
	HSet(key, field string, value interface{}) error
	HGet(key, field string) (string, error)
	HDel(key string, fields ...string) (int64, error)
	HKeys(key string) ([]string, error)
	Subscribe(channels ...string) (*redis.PubSub, error)
	Unsubscribe(pubsub *redis.PubSub, channels ...string) error
	PSubscribe(patterns ...string) (*redis.PubSub, error)
	PUnsubscribe(pubsub *redis.PubSub, patterns ...string) error
	Expire(key string, expiration time.Duration) (bool, error)
	Keys(pattern string) ([]string, error)
	ListenForKeyExpiration()
}

// RedisClient is a struct that holds the Redis client instance.
type RedisClient struct {
	client *redis.Client
}

// Init initializes a new Redis client and returns it.
func Init(ctx context.Context) (*RedisClient, error) {
	log := logger.Logger(ctx)

	// Create a new Redis client
	client := redis.NewClient(&redis.Options{
		Addr:            constants.Config.RedisConfig.REDIS_HOST + ":" + constants.Config.RedisConfig.REDIS_PORT, // Redis server address (e.g., "localhost:6379")
		Password:        constants.Config.RedisConfig.REDIS_PASSWORD,                                             // Password (leave empty if not required)
		DB:              constants.Config.RedisConfig.REDIS_DB,                                                   // Database number
		MaxRetries:      constants.Config.RedisConfig.REDIS_MAX_RETRIES,                                          // Maximum number of retries before giving up
		DialTimeout:     time.Duration(constants.Config.RedisConfig.REDIS_DIAl_TIMEOUT) * time.Second,            // Timeout for establishing new connections
		MaxActiveConns:  constants.Config.RedisConfig.REDIS_MAX_OPEN_CONNECTION,                                  // Maximum number of connections in the pool
		MaxIdleConns:    constants.Config.RedisConfig.REDIS_MAX_IDLE_CONNECTION,                                  // Maximum number of idle connections in the pool
		ConnMaxLifetime: time.Duration(constants.Config.RedisConfig.REDIS_CONNECTION_MAX_LIFETIME) * time.Second, // Maximum amount of time a connection may be reused
	})

	// Ping the Redis server to check if it's reachable
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Errorf("could not connect to Redis: %v", err)
		return nil, fmt.Errorf("could not connect to Redis: %v", err)
	}

	return &RedisClient{client: client}, nil
}

// GetRedisClient returns an existing Redis client instance.
func GetRedisClient(redisClient *RedisClient) *redis.Client {
	return redisClient.client
}

// Ping pings the Redis server to check if it's reachable.
func (rc *RedisClient) Ping() error {
	ctx := context.Background()
	_, err := rc.client.Ping(ctx).Result()
	if err != nil {
		return err
	}
	return nil
}

// Close closes the Redis client connection.
func (rc *RedisClient) Close() error {
	return rc.client.Close()
}

// Set sets a key-value pair in Redis with an optional expiration time.
func (rc *RedisClient) Set(key, value string, expiration time.Duration) error {
	ctx := context.Background()
	return rc.client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value from Redis based on the key.
func (rc *RedisClient) Get(key string) (string, error) {
	ctx := context.Background()
	result, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key '%s' not found", key)
	} else if err != nil {
		return "", err
	}
	return result, nil
}

// Delete deletes one or more keys in Redis.
func (rc *RedisClient) Delete(keys ...string) (int64, error) {
	ctx := context.Background()
	result, err := rc.client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// Exists checks if a key exists in Redis.
func (rc *RedisClient) Exists(key string) (bool, error) {
	ctx := context.Background()
	result, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

// HMSet sets multiple fields and values in a Redis hash.
func (rc *RedisClient) HMSet(key string, fieldsAndValues map[string]interface{}) error {
	ctx := context.Background()
	return rc.client.HMSet(ctx, key, fieldsAndValues).Err()
}

// HGetAll gets all fields and values from a Redis hash.
func (rc *RedisClient) HGetAll(key string) (map[string]string, error) {
	ctx := context.Background()
	result, err := rc.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Publish publishes a message to a Redis channel.
func (rc *RedisClient) Publish(channel, message string) error {
	ctx := context.Background()
	return rc.client.Publish(ctx, channel, message).Err()
}

// LPush prepends one or more values to a Redis list.
func (rc *RedisClient) LPush(key string, values ...interface{}) error {
	ctx := context.Background()
	return rc.client.LPush(ctx, key, values...).Err()
}

// RPush appends one or more values to a Redis list.
func (rc *RedisClient) RPush(key string, values ...interface{}) error {
	ctx := context.Background()
	return rc.client.RPush(ctx, key, values...).Err()
}

// LPop removes and returns the first element from a Redis list.
func (rc *RedisClient) LPop(key string) (string, error) {
	ctx := context.Background()
	result, err := rc.client.LPop(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("list '%s' is empty", key)
	} else if err != nil {
		return "", err
	}
	return result, nil
}

// RPop removes and returns the last element from a Redis list.
func (rc *RedisClient) RPop(key string) (string, error) {
	ctx := context.Background()
	result, err := rc.client.RPop(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("list '%s' is empty", key)
	} else if err != nil {
		return "", err
	}
	return result, nil
}

// SAdd adds one or more members to a Redis set.
func (rc *RedisClient) SAdd(key string, members ...interface{}) (int64, error) {
	ctx := context.Background()
	result, err := rc.client.SAdd(ctx, key, members...).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// SRem removes one or more members from a Redis set.
func (rc *RedisClient) SRem(key string, members ...interface{}) (int64, error) {
	ctx := context.Background()
	result, err := rc.client.SRem(ctx, key, members...).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// SMembers gets all members of a Redis set.
func (rc *RedisClient) SMembers(key string) ([]string, error) {
	ctx := context.Background()
	result, err := rc.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SIsMember checks if a member exists in a Redis set.
func (rc *RedisClient) SIsMember(key string, member interface{}) (bool, error) {
	ctx := context.Background()
	result, err := rc.client.SIsMember(ctx, key, member).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// ZAdd adds one or more members to a Redis sorted set.
func (rc *RedisClient) ZAdd(key string, members ...redis.Z) (int64, error) {
	ctx := context.Background()
	result, err := rc.client.ZAdd(ctx, key, members...).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// ZRange gets a range of members from a Redis sorted set.
func (rc *RedisClient) ZRange(key string, start, stop int64) ([]string, error) {
	ctx := context.Background()
	result, err := rc.client.ZRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ZScore gets the score of a member in a Redis sorted set.
func (rc *RedisClient) ZScore(key string, member string) (float64, error) {
	ctx := context.Background()
	result, err := rc.client.ZScore(ctx, key, member).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// HSet sets the string value of a hash field.
func (rc *RedisClient) HSet(key, field string, value interface{}) error {
	ctx := context.Background()
	return rc.client.HSet(ctx, key, field, value).Err()
}

// HGet gets the value of a hash field.
func (rc *RedisClient) HGet(key, field string) (string, error) {
	ctx := context.Background()
	result, err := rc.client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("field '%s' not found in hash '%s'", field, key)
	} else if err != nil {
		return "", err
	}
	return result, nil
}

// HDel deletes one or more hash fields.
func (rc *RedisClient) HDel(key string, fields ...string) (int64, error) {
	ctx := context.Background()
	result, err := rc.client.HDel(ctx, key, fields...).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// HKeys gets all the fields in a hash.
func (rc *RedisClient) HKeys(key string) ([]string, error) {
	ctx := context.Background()
	result, err := rc.client.HKeys(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Subscribe subscribes to one or more Redis channels.
func (rc *RedisClient) Subscribe(channels ...string) (*redis.PubSub, error) {
	ctx := context.Background()
	pubsub := rc.client.Subscribe(ctx, channels...)
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, err
	}
	return pubsub, nil
}

// Unsubscribe unsubscribes from one or more Redis channels.
func (rc *RedisClient) Unsubscribe(pubsub *redis.PubSub, channels ...string) error {
	ctx := context.Background()
	return pubsub.Unsubscribe(ctx, channels...)
}

// PSubscribe subscribes to one or more Redis patterns.
func (rc *RedisClient) PSubscribe(patterns ...string) (*redis.PubSub, error) {
	ctx := context.Background()
	pubsub := rc.client.PSubscribe(ctx, patterns...)
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, err
	}
	return pubsub, nil
}

// PUnsubscribe unsubscribes from one or more Redis patterns.
func (rc *RedisClient) PUnsubscribe(pubsub *redis.PubSub, patterns ...string) error {
	ctx := context.Background()
	return pubsub.PUnsubscribe(ctx, patterns...)
}

// Expire sets a timeout on a Redis key.
func (rc *RedisClient) Expire(key string, expiration time.Duration) (bool, error) {
	ctx := context.Background()
	result, err := rc.client.Expire(ctx, key, expiration).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

// Keys gets all keys matching the given pattern.
func (rc *RedisClient) Keys(pattern string) ([]string, error) {
	ctx := context.Background()
	result, err := rc.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListenForKeyExpiration listens for expired keys and deletes them.
func (rc *RedisClient) ListenForKeyExpiration() {
	ctx := context.Background()
	pubsub := rc.client.PSubscribe(ctx, "__keyevent@*__:expired")
	for {
		fmt.Println("listening for expired keys...")
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("expired key:", msg.Payload)
	}
}

// TODO Multi starts a Redis transaction.
// TODO Exec executes all queued commands in a Redis transaction.
// TODO Discard discards all commands in a Redis transaction.
