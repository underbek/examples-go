package pgtaskpool

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	ctxerrors "github.com/underbek/examples-go/errors"
	"github.com/underbek/examples-go/logger"
	goKitPgx "github.com/underbek/examples-go/storage/pgx"
	"github.com/underbek/examples-go/tracing"
)

func (w *worker) createTask(ctx context.Context, task Task) error {
	ctx, span := tracing.StartSpan(ctx, "pgtaskpool", "Create task")
	defer span.End()

	meta := logger.ParseCtxMeta(ctx)

	sql, args, err := sq.Insert(w.config.Table).
		Columns(
			"transaction_id",
			"status",
			"custom_schedule",
			"schedule_type",
			"process_at",
			"trace_meta",
		).
		Values(
			task.TransactionID,
			w.config.ProcessStatus,
			task.CustomScheduleSlice,
			task.Type,
			time.Now().Add(time.Second*time.Duration(task.InitDelaySec)),
			Trace{
				TraceID: span.SpanContext().TraceID().String(),
				SpanID:  span.SpanContext().SpanID().String(),
				Meta:    meta,
			},
		).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to get query string")
	}

	_, err = w.storage.Exec(ctx, sql, args...)
	if err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to make query")
	}

	return nil
}

func (w *worker) cancelTask(ctx context.Context, id uint64) error {
	sql, args, err := sq.Update(w.config.Table).
		Where(sq.Eq{"transaction_id": id}).
		Where(sq.Eq{"status": w.config.ProcessStatus}).
		Set("status", w.config.CanceledStatus).
		Set("updated_at", "now()").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to get query string")
	}

	_, err = w.storage.Exec(ctx, sql, args...)
	if err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to make query")
	}
	return nil
}

func (w *worker) getTask(ctx context.Context) (Task, error) {
	sql, args, err := sq.Select(
		"id",
		"transaction_id",
		"attempts",
		"custom_schedule",
		"schedule_type",
		"trace_meta",
		"last_error_code",
		"last_error_message",
		"status",
		"process_at",
		"lock_time",
	).
		From(w.config.Table).
		Where(
			sq.And{
				sq.Eq{"status": w.config.ProcessStatus},
				sq.LtOrEq{"process_at": "now()"},
				sq.Or{
					sq.Eq{"lock_time": nil},
					sq.LtOrEq{"lock_time": "now()"},
				},
			},
		).
		OrderBy("id ASC").
		Suffix("FOR UPDATE SKIP LOCKED").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return Task{}, ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to get query string")
	}

	rows, err := w.storage.Query(ctx, sql, args...)
	if err != nil {
		return Task{}, err
	}

	defer rows.Close()

	task, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Task])
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func (w *worker) updateTask(ctx context.Context, task Task) error {
	upd := sq.Update(w.config.Table).
		Set("attempts", task.Attempts).
		Set("updated_at", "now()").
		Set("lock_time", nil)

	if task.Status != nil {
		upd = upd.Set("status", task.Status)
	}

	if task.LastErrorCode != nil {
		upd = upd.Set("last_error_code", task.LastErrorCode)
	}

	if task.LastErrorMessage != nil {
		upd = upd.Set("last_error_message", task.LastErrorMessage)
	}

	if task.NextProcessingTime != nil {
		upd = upd.Set("process_at", task.NextProcessingTime)
	}

	sql, args, err := upd.Where(sq.Eq{"id": task.ID}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to get query string")
	}

	if _, err := w.storage.Exec(ctx, sql, args...); err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to exec update query")
	}

	return nil
}

func (w *worker) isTaskExists(ctx context.Context, id uint64) (bool, error) {
	sql, args, err := sq.Select("count(transaction_id)").From(w.config.Table).
		Where(sq.And{
			sq.Eq{"status": w.config.ProcessStatus},
			sq.Eq{"transaction_id": id},
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to get query string")
	}

	rows, err := w.storage.Query(ctx, sql, args...)
	if err != nil {
		return false, err
	}

	count, err := pgx.CollectOneRow(rows, pgx.RowTo[int64])
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

type tasksFilter struct {
	status        string
	processBefore time.Time
}

func (w *worker) countTasks(ctx context.Context, filter tasksFilter) (uint32, error) {
	query := sq.Select("COUNT(1)").
		From(w.config.Table).
		PlaceholderFormat(sq.Dollar)

	if filter.status != "" {
		query = query.Where(sq.Eq{"status": filter.status})
	}
	if !filter.processBefore.IsZero() {
		query = query.Where(sq.Lt{"process_at": filter.processBefore})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to build query")
	}

	var count uint32
	if err = w.storage.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to query and scan result")
	}

	return count, nil
}

func (w *worker) blockTask(ctx context.Context, tx goKitPgx.Transaction, task Task) error {
	if task.LockTime == nil {
		return ctxerrors.New(ctxerrors.TypeInternal, "lock time is nil")
	}

	sql, args, err := sq.Update(w.config.Table).
		Set("lock_time", task.LockTime).
		Where(sq.Eq{"id": task.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to get query string")
	}

	if _, err := tx.Exec(ctx, sql, args...); err != nil {
		return ctxerrors.Wrap(err, ctxerrors.TypeDatabase, "failed to exec update query")
	}

	return nil
}
