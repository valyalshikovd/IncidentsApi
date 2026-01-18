package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewHandler(db *pgxpool.Pool, redis *redis.Client) *Handler {
	return &Handler{db: db, redis: redis}
}

func (h *Handler) Health(c *gin.Context) {
	dbOK := true
	if h.db != nil {
		if err := h.db.Ping(c.Request.Context()); err != nil {
			dbOK = false
		}
	}

	redisOK := true
	if h.redis != nil {
		if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
			redisOK = false
		}
	}

	status := http.StatusOK
	if !dbOK || !redisOK {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status":   "ok",
		"db_ok":    dbOK,
		"redis_ok": redisOK,
	})
}
