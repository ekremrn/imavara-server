// Package credits serializes mutations of each browser's credit account.
package credits

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
)

var ErrInsufficient = errors.New("insufficient credits")
var ErrKeyConflict = errors.New("operation key already belongs to a different request")
var ErrState = errors.New("invalid generation transition")

type Ledger struct{ DB *sql.DB }
type Generation struct {
	Status  string
	AssetID sql.NullString
}

// ReserveOnlyOnce returns start=true only to the caller that acquired the reservation.
func (l Ledger) ReserveOnlyOnce(ctx context.Context, browser, key string, digest []byte) (record Generation, start bool, err error) {
	tx, err := l.DB.BeginTx(ctx, nil)
	if err != nil {
		return record, false, err
	}
	defer tx.Rollback()
	if err = lockAccount(ctx, tx, browser); err != nil {
		return record, false, err
	}
	var saved []byte
	err = tx.QueryRowContext(ctx, "SELECT request_hash,status,asset_id FROM generations WHERE browser_id=$1 AND operation_key=$2", browser, key).Scan(&saved, &record.Status, &record.AssetID)
	if err == nil {
		if !bytes.Equal(saved, digest) {
			return record, false, ErrKeyConflict
		}
		return record, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return record, false, err
	}
	var remaining int
	err = tx.QueryRowContext(ctx, `SELECT (SELECT COALESCE(sum(amount),0) FROM credit_movements WHERE browser_id=$1)
 - (SELECT COALESCE(sum(amount),0) FROM credit_holds WHERE browser_id=$1)`, browser).Scan(&remaining)
	if err != nil {
		return record, false, err
	}
	if remaining < 3 {
		return record, false, ErrInsufficient
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO generations(browser_id,operation_key,request_hash,status) VALUES($1,$2,$3,'reserved')", browser, key, digest); err != nil {
		return record, false, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO credit_holds(browser_id,operation_key,amount) VALUES($1,$2,3)", browser, key); err != nil {
		return record, false, err
	}
	if err = tx.Commit(); err != nil {
		return record, false, err
	}
	record.Status = "reserved"
	return record, true, nil
}

func lockAccount(ctx context.Context, tx *sql.Tx, browser string) error {
	var exists int
	return tx.QueryRowContext(ctx, "SELECT 1 FROM browsers WHERE id=$1 FOR UPDATE", browser).Scan(&exists)
}

// Submit claims the right to contact the provider. An ambiguous outcome keeps its hold.
func (l Ledger) Submit(ctx context.Context, browser, key string) error {
	result, err := l.DB.ExecContext(ctx, "UPDATE generations SET status='submitted' WHERE browser_id=$1 AND operation_key=$2 AND status='reserved'", browser, key)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrState
	}
	return nil
}

// Finish runs in the same transaction as asset registration. A nil asset means definite failure.
func Finish(ctx context.Context, tx *sql.Tx, browser, key string, asset *string) error {
	if err := lockAccount(ctx, tx, browser); err != nil {
		return err
	}
	var state string
	var existing sql.NullString
	err := tx.QueryRowContext(ctx, "SELECT status,asset_id FROM generations WHERE browser_id=$1 AND operation_key=$2 FOR UPDATE", browser, key).Scan(&state, &existing)
	if err != nil {
		return err
	}
	if state == "failed" && asset == nil || state == "delivered" && asset != nil && existing.String == *asset {
		return nil
	}
	if state != "submitted" && !(state == "reserved" && asset == nil) {
		return ErrState
	}
	next := "failed"
	if asset != nil {
		next = "delivered"
		if _, err = tx.ExecContext(ctx, "INSERT INTO credit_movements(browser_id,operation_key,reason,amount) VALUES($1,$2,'generation',-3)", browser, "generation:"+key); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, "UPDATE generations SET status=$3,asset_id=$4 WHERE browser_id=$1 AND operation_key=$2", browser, key, next, asset); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM credit_holds WHERE browser_id=$1 AND operation_key=$2", browser, key)
	return err
}

// ReleaseExpired closes interrupted operations after the maximum provider and storage deadline.
// It never resubmits a provider request and leaves committed deliveries intact.
func (l Ledger) ReleaseExpired(ctx context.Context) error {
	_, err := l.DB.ExecContext(ctx, `WITH expired AS (
 UPDATE generations SET status='failed',
 provider_outcome=CASE WHEN status='submitted' THEN 'unknown' ELSE 'unsubmitted' END
 WHERE status IN ('reserved','submitted') AND created_at < now()-interval '15 minutes'
 RETURNING browser_id, operation_key
 )
 DELETE FROM credit_holds h USING expired e
 WHERE h.browser_id=e.browser_id AND h.operation_key=e.operation_key`)
	return err
}
