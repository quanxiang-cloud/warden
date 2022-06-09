package store

import (
	"context"
	"fmt"
	"github.com/quanxiang-cloud/warden/pkg/jwts"
	"github.com/quanxiang-cloud/warden/pkg/jwts/models"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"time"
)

var (
	_             jwts.TokenStore = &RedisTokenStore{}
	jsonMarshal                   = jsoniter.Marshal
	jsonUnmarshal                 = jsoniter.Unmarshal
)

const (
	// JWTRedis 当前服务前缀
	JWTRedis = "jwt:"
	// JWTRedisUsers 当前服务用户相关
	JWTRedisUsers = "jwt:users:"
)

// NewRedisStore create an instance of a redis store
func NewRedisStore(opts *redis.Options, keyNamespace ...string) *RedisTokenStore {
	if opts == nil {
		panic("options cannot be nil")
	}
	return NewRedisStoreWithCli(redis.NewClient(opts), keyNamespace...)
}

// NewRedisStoreWithCli create an instance of a redis store
func NewRedisStoreWithCli(cli *redis.Client, keyNamespace ...string) *RedisTokenStore {
	store := &RedisTokenStore{
		cli: cli,
	}

	if len(keyNamespace) > 0 {
		store.ns = keyNamespace[0]
	}
	return store
}

// NewRedisClusterStore create an instance of a redis cluster store
func NewRedisClusterStore(opts *redis.ClusterOptions, keyNamespace ...string) *RedisTokenStore {
	if opts == nil {
		panic("options cannot be nil")
	}
	return NewRedisClusterStoreWithCli(redis.NewClusterClient(opts), keyNamespace...)
}

// NewRedisClusterStoreWithCli create an instance of a redis cluster store
func NewRedisClusterStoreWithCli(cli *redis.ClusterClient, keyNamespace ...string) *RedisTokenStore {
	store := &RedisTokenStore{
		cli: cli,
	}

	if len(keyNamespace) > 0 {
		store.ns = keyNamespace[0]
	}
	return store
}

type clienter interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Exists(ctx context.Context, key ...string) *redis.IntCmd
	TxPipeline() redis.Pipeliner
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	HDel(ctx context.Context, key string, values ...string) *redis.IntCmd
	HKeys(ctx context.Context, key string) *redis.StringSliceCmd
	HGet(ctx context.Context, key, field string) *redis.StringCmd
	Close() error
}

// RedisTokenStore redis token store
type RedisTokenStore struct {
	cli clienter
	ns  string
}

// Close close the store
func (s *RedisTokenStore) Close() error {
	return s.cli.Close()
}

func (s *RedisTokenStore) wrapperKey(key string) string {
	return fmt.Sprintf("%s%s", s.ns, key)
}

func (s *RedisTokenStore) checkError(result redis.Cmder) (bool, error) {
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// remove
func (s *RedisTokenStore) remove(ctx context.Context, key string) error {
	result := s.cli.Del(ctx, s.wrapperKey(JWTRedis+key))
	_, err := s.checkError(result)
	return err
}

// remove
func (s *RedisTokenStore) hRemove(ctx context.Context, key string, values ...string) error {
	result := s.cli.HDel(ctx, s.wrapperKey(JWTRedisUsers+key), values...)
	_, err := s.checkError(result)
	return err
}

func (s *RedisTokenStore) removeToken(ctx context.Context, tokenString string, isRefresh bool) error {
	basicID, err := s.getBasicID(ctx, tokenString)
	if err != nil {
		return err
	} else if basicID == "" {
		return nil
	}

	err = s.remove(ctx, tokenString)
	if err != nil {
		return err
	}

	token, err := s.getToken(ctx, basicID)
	if err != nil {
		return err
	} else if token == nil {
		return nil
	}

	checkToken := token.GetRefresh()
	if isRefresh {
		checkToken = token.GetAccess()
	}
	iresult := s.cli.Exists(ctx, s.wrapperKey(checkToken))
	if err := iresult.Err(); err != nil && err != redis.Nil {
		return err
	} else if iresult.Val() == 0 {
		return s.remove(ctx, basicID)
	}

	return nil
}

func (s *RedisTokenStore) parseToken(result *redis.StringCmd) (jwts.TokenInfo, error) {
	if ok, err := s.checkError(result); err != nil {
		return nil, err
	} else if ok {
		return nil, nil
	}

	buf, err := result.Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var token models.Token
	if err := jsonUnmarshal(buf, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *RedisTokenStore) getToken(ctx context.Context, key string) (jwts.TokenInfo, error) {
	result := s.cli.Get(ctx, s.wrapperKey(JWTRedis+key))
	return s.parseToken(result)
}

func (s *RedisTokenStore) parseBasicID(result *redis.StringCmd) (string, error) {
	if ok, err := s.checkError(result); err != nil {
		return "", err
	} else if ok {
		return "", nil
	}
	return result.Val(), nil
}

func (s *RedisTokenStore) getBasicID(ctx context.Context, token string) (string, error) {
	result := s.cli.Get(ctx, s.wrapperKey(JWTRedis+token))
	return s.parseBasicID(result)
}

func (s *RedisTokenStore) cleanByUser(ctx context.Context, userID string) {
	//遍历用户下的keys，如果某个key的value的value没有了，那就是无效key，就清除
	basicIDs := s.cli.HKeys(ctx, JWTRedisUsers+userID).Val()
	for _, v := range basicIDs {
		otherAccess := s.cli.HGet(ctx, JWTRedisUsers+userID, v).Val()
		b := s.cli.Get(ctx, JWTRedis+otherAccess).Val()
		if b == "" {
			s.cli.HDel(ctx, JWTRedisUsers+userID, v)
		}
	}
}

// Create Create and store the new token information
func (s *RedisTokenStore) Create(ctx context.Context, info jwts.TokenInfo) error {
	ct := time.Now()
	jv, err := jsonMarshal(info)
	if err != nil {
		return err
	}
	var userID = ""
	pipe := s.cli.TxPipeline()

	userID = info.GetUserID()
	basicID := uuid.Must(uuid.NewRandom()).String()
	aexp := info.GetAccessExpiresIn()
	rexp := aexp

	if refresh := info.GetRefresh(); refresh != "" {
		rexp = info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn()).Sub(ct)
		if aexp.Seconds() > rexp.Seconds() {
			aexp = rexp
		}
		pipe.Set(ctx, s.wrapperKey(JWTRedis+refresh), basicID, rexp)
	}

	pipe.Set(ctx, s.wrapperKey(JWTRedis+info.GetAccess()), basicID, aexp)
	pipe.Set(ctx, s.wrapperKey(JWTRedis+basicID), jv, rexp)

	pipe.HSet(ctx, s.wrapperKey(JWTRedisUsers+userID), basicID, info.GetAccess())
	pipe.Expire(ctx, s.wrapperKey(JWTRedisUsers+userID), info.GetRefreshExpiresIn())
	if _, err := pipe.Exec(ctx); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// RemoveByCode Use the authorization code to delete the token information
func (s *RedisTokenStore) RemoveByCode(ctx context.Context, code string) error {
	return s.remove(ctx, code)
}

// RemoveByAccess Use the access token to delete the token information
func (s *RedisTokenStore) RemoveByAccess(ctx context.Context, access string) error {
	err := s.removeToken(ctx, access, false)

	return err
}

// RemoveByRefresh Use the refresh token to delete the token information
func (s *RedisTokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	tokenInfo, _ := s.GetByRefresh(ctx, refresh)
	err := s.removeToken(ctx, refresh, true)
	if tokenInfo != nil {
		userID := tokenInfo.GetUserID()
		s.cleanByUser(ctx, userID)
	}

	return err
}

// GetByCode Use the authorization code for token information data
func (s *RedisTokenStore) GetByCode(ctx context.Context, code string) (jwts.TokenInfo, error) {
	return s.getToken(ctx, code)
}

// GetByAccess Use the access token for token information data
func (s *RedisTokenStore) GetByAccess(ctx context.Context, access string) (jwts.TokenInfo, error) {
	basicID, err := s.getBasicID(ctx, access)
	if err != nil || basicID == "" {
		return nil, err
	}
	return s.getToken(ctx, basicID)
}

// GetByRefresh Use the refresh token for token information data
func (s *RedisTokenStore) GetByRefresh(ctx context.Context, refresh string) (jwts.TokenInfo, error) {
	basicID, err := s.getBasicID(ctx, refresh)
	if err != nil || basicID == "" {
		return nil, err
	}
	return s.getToken(ctx, basicID)
}

// RemoveToken Use the jti to delete the token information data
func (s *RedisTokenStore) RemoveToken(ctx context.Context, jti string) error {
	keys := s.cli.HKeys(ctx, JWTRedisUsers+jti).Val()
	for _, v := range keys {
		access := s.cli.HGet(ctx, JWTRedisUsers+jti, v).Val()
		tokenInfo, _ := s.GetByAccess(ctx, access)
		if tokenInfo == nil {
			continue
		}
		_ = s.RemoveByAccess(ctx, tokenInfo.GetAccess())
		if tokenInfo != nil {
			userID := tokenInfo.GetUserID()
			s.cleanByUser(ctx, userID)
		}
		_ = s.RemoveByRefresh(ctx, tokenInfo.GetRefresh())
		s.cli.Del(ctx, JWTRedis+v)
		s.cli.HDel(ctx, JWTRedisUsers+jti, v)
	}
	return nil
}
