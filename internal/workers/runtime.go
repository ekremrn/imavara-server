package workers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"imavara.local/server/internal/platform/observe"
	"log/slog"
	"os"
	"time"
)

const extractQueue = "assets.extract.v1"
const indexQueue = "assets.index.v1"

type Job struct {
	Asset   string `json:"asset_id"`
	Object  string `json:"object_key"`
	Version int    `json:"version"`
	Attempt int    `json:"attempt"`
}
type runtime struct {
	db  *sql.DB
	pub *amqp.Channel
}

func (r runtime) publish(ctx context.Context, queue string, j Job) error {
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	confirmation, err := r.pub.PublishWithDeferredConfirmWithContext(ctx, "", queue, false, false, amqp.Publishing{DeliveryMode: amqp.Persistent, ContentType: "application/json", Body: b})
	if err != nil {
		return err
	}
	ok, err := confirmation.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("broker rejected job")
	}
	return nil
}
func (r runtime) dispatch(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var j Job
	err = tx.QueryRowContext(ctx, `SELECT j.asset_id,a.object_key,j.version FROM catalog_jobs j JOIN assets a ON a.id=j.asset_id WHERE NOT j.published ORDER BY j.asset_id LIMIT 1 FOR UPDATE OF j SKIP LOCKED`).Scan(&j.Asset, &j.Object, &j.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = r.publish(ctx, extractQueue, j); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE catalog_jobs SET published=true WHERE asset_id=$1 AND version=$2", j.Asset, j.Version); err != nil {
		return err
	}
	return tx.Commit()
}
func Run(extract bool) error {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()
	conn, err := amqp.Dial(os.Getenv("AMQP_URL"))
	if err != nil {
		return err
	}
	defer conn.Close()
	sub, err := conn.Channel()
	if err != nil {
		return err
	}
	defer sub.Close()
	pub, err := conn.Channel()
	if err != nil {
		return err
	}
	defer pub.Close()
	for _, q := range []string{extractQueue, indexQueue, extractQueue + ".failed", indexQueue + ".failed"} {
		if _, err = sub.QueueDeclare(q, true, false, false, false, nil); err != nil {
			return err
		}
	}
	if err = pub.Confirm(false); err != nil {
		return err
	}
	rt := runtime{db, pub}
	queue := indexQueue
	if extract {
		queue = extractQueue
	}
	if err = sub.Qos(1, 0, false); err != nil {
		return err
	}
	deliveries, err := sub.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if extract {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				err = rt.dispatch(ctx)
				cancel()
				if err != nil {
					return err
				}
			}
		case d, open := <-deliveries:
			if !open {
				return errors.New("consumer disconnected")
			}
			observe.Add("jobs_received_total", 1)
			var j Job
			err = json.Unmarshal(d.Body, &j)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			if err == nil {
				if extract {
					err = rt.extract(ctx, j)
				} else {
					err = rt.index(ctx, j)
				}
			}
			cancel()
			if err != nil {
				observe.Add("jobs_failed_total", 1)
				observe.Error(err)
				slog.Warn("catalog job failed", "asset_id", j.Asset, "version", j.Version, "attempt", j.Attempt, "error", err)
				target := queue
				j.Attempt++
				if j.Attempt >= 3 {
					target = queue + ".failed"
				}
				ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
				err = rt.publish(ctx, target, j)
				cancel()
				if err != nil {
					return err
				}
			}
			if err = d.Ack(false); err != nil {
				return err
			}
		}
	}
}
func analysisKey(j Job) string { return fmt.Sprintf("%s/%d/extract.v1.json", j.Asset, j.Version) }

// Replay moves a bounded snapshot of failed jobs back to their original queue.
// The failed delivery is acknowledged only after the broker confirms publication.
func Replay(queue string, limit int) (int, error) {
	if (queue != extractQueue && queue != indexQueue) || limit < 1 {
		return 0, errors.New("expected an extract or index queue and a positive limit")
	}
	conn, err := amqp.Dial(os.Getenv("AMQP_URL"))
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return 0, err
	}
	defer ch.Close()
	if err = ch.Confirm(false); err != nil {
		return 0, err
	}
	snapshot, err := ch.QueueInspect(queue + ".failed")
	if err != nil {
		return 0, err
	}
	count := min(limit, snapshot.Messages)
	for i := 0; i < count; i++ {
		delivery, ok, err := ch.Get(queue+".failed", false)
		if err != nil {
			return i, err
		}
		if !ok {
			return i, nil
		}
		var job Job
		if err = json.Unmarshal(delivery.Body, &job); err != nil {
			return i, err
		}
		job.Attempt = 0
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = (runtime{pub: ch}).publish(ctx, queue, job)
		cancel()
		if err != nil {
			return i, err
		}
		if err = delivery.Ack(false); err != nil {
			return i, err
		}
	}
	return count, nil
}
