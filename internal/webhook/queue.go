package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type PayloadIncident struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Description   *string `json:"description,omitempty"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	DangerRadiusM int     `json:"danger_radius_m"`
	DistanceM     float64 `json:"distance_m"`
}

type Payload struct {
	CheckID   int64             `json:"check_id"`
	UserID    string            `json:"user_id"`
	Latitude  float64           `json:"latitude"`
	Longitude float64           `json:"longitude"`
	Incidents []PayloadIncident `json:"incidents"`
	CreatedAt time.Time         `json:"created_at"`
}

type job struct {
	Payload Payload `json:"payload"`
	Attempt int     `json:"attempt"`
}

type Queue struct {
	redis      *redis.Client
	webhookURL string
	queueKey   string
	maxRetries int
	retryBase  time.Duration
	client     *http.Client
}

func NewQueue(redisClient *redis.Client, webhookURL string, maxRetries int, retryBase time.Duration) *Queue {
	return &Queue{
		redis:      redisClient,
		webhookURL: webhookURL,
		queueKey:   "queue:webhook",
		maxRetries: maxRetries,
		retryBase:  retryBase,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (q *Queue) Enqueue(ctx context.Context, payload Payload) error {
	if q == nil || q.redis == nil || q.webhookURL == "" {
		return nil
	}
	b, err := json.Marshal(job{Payload: payload, Attempt: 0})
	if err != nil {
		return err
	}
	return q.redis.LPush(ctx, q.queueKey, b).Err()
}

func (q *Queue) Run(ctx context.Context) {
	if q == nil || q.redis == nil || q.webhookURL == "" {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		items, err := q.redis.BRPop(ctx, 2*time.Second, q.queueKey).Result()
		if err == redis.Nil || err == context.DeadlineExceeded {
			continue
		}
		if err != nil || len(items) < 2 {
			continue
		}

		var j job
		if err := json.Unmarshal([]byte(items[1]), &j); err != nil {
			continue
		}

		if err := q.send(ctx, j.Payload); err != nil {
			if j.Attempt+1 <= q.maxRetries {
				j.Attempt++
				q.scheduleRetry(ctx, j)
			}
		}
	}
}

func (q *Queue) send(ctx context.Context, payload Payload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, q.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errHTTPStatus(resp.StatusCode)
	}
	return nil
}

func (q *Queue) scheduleRetry(ctx context.Context, j job) {
	delay := q.retryBase * time.Duration(1<<uint(j.Attempt-1))
	time.AfterFunc(delay, func() {
		b, err := json.Marshal(j)
		if err != nil {
			return
		}
		_ = q.redis.LPush(ctx, q.queueKey, b).Err()
	})
}

type errHTTPStatus int

func (e errHTTPStatus) Error() string {
	return "webhook status " + http.StatusText(int(e))
}
