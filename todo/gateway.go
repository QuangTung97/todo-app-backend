package todo

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/glog"
	"github.com/jmoiron/sqlx"
)

// Gateway : struct for Gateway
type Gateway struct {
	db          *sqlx.DB
	redisClient *redis.Client
}

// NewGateway : Create a new Gateway
func NewGateway(db *sqlx.DB, redisClient *redis.Client) *Gateway {
	return &Gateway{
		db:          db,
		redisClient: redisClient,
	}
}

func (g *Gateway) setRedisValue(ctx context.Context) valueSetter {
	return func(key, value string, expiration time.Duration) error {
		return g.redisClient.Set(ctx, key, value, expiration).Err()
	}
}

func (g *Gateway) getRedisValue(ctx context.Context) valueGetter {
	return func(key string) (string, error) {
		value, err := g.redisClient.Get(ctx, key).Result()
		if err == redis.Nil {
			return value, errKeyNotExist
		}
		return value, err
	}
}

func (g *Gateway) setRedisExpiration(ctx context.Context) expirationSetter {
	return func(key string, expiration time.Duration) error {
		return g.redisClient.Expire(ctx, key, expiration).Err()
	}
}

func (g *Gateway) getAccount(ctx context.Context) accountGetter {
	return func(username string) (int, string, error) {

		type Account struct {
			Hash string `db:"password_hash"`
			ID   int    `db:"id"`
		}
		query := `SELECT id, password_hash FROM account WHERE username = ?`

		var account Account
		err := g.db.GetContext(ctx, &account, query, username)
		if err == sql.ErrNoRows {
			return account.ID, account.Hash, errAccountNotExist
		}
		return account.ID, account.Hash, err
	}
}

// Authenticated : authentication middleware
func (g *Gateway) Authenticated(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		info := basicAuthInfo{
			username: username,
			password: password,
			ok:       ok,
		}
		token := r.Header.Get("X-Auth-Token")

		ctx := r.Context()
		id, token, ok, err := verifyCredentials(
			info, token,
			g.setRedisValue(ctx),
			g.getRedisValue(ctx),
			g.setRedisExpiration(ctx),
			g.getAccount(ctx),
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			glog.Error(err)
			return
		}

		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			glog.Error("unauthenticated")
			return
		}

		w.Header().Add("X-Auth-Token", token)
		r.Header.Add("Grpc-Metadata-Account-ID", strconv.FormatInt(int64(id), 10))
		handler.ServeHTTP(w, r)
	})
}
