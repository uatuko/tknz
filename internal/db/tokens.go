package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type token struct {
	attrs any
	exp   time.Time
	id    string
}

func (t *token) Exp() time.Time {
	return t.exp
}

func (t *token) Expire(ctx context.Context) error {
	if t.exp.IsZero() {
		// Already set as expired, avoid unnecessary db writes
		return nil
	}

	qry := `
		update tokens
		set exp = $3::timestamp
		where
			id = $1::text
			and exp = $2::timestamp
		;
	`

	tm := time.Time{}
	cmd, err := pg.Exec(ctx, qry, t.id, t.exp, tm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return wrapError(err, ErrUnknown)
	}

	if cmd.RowsAffected() != 1 {
		return ErrConflict
	}

	t.exp = tm

	return nil
}

func (t *token) Expired() bool {
	return t.exp.Before(time.Now())
}

func (t *token) insert(ctx context.Context) error {
	attrs, err := json.Marshal(t.attrs)
	if err != nil {
		return wrapError(err, ErrInvalidData)
	}

	t.exp = time.UnixMicro(t.exp.UnixMicro()).UTC()

	qry := `
		insert into tokens (
			id,
			exp,
			attrs
		) values (
			$1::text,
			$2::timestamp,
			$3::jsonb
		);
	`

	if _, err := pg.Exec(ctx, qry, t.id, t.exp, attrs); err != nil {
		return err
	}

	return nil
}

func retrieveToken(ctx context.Context, id string, out *token) error {
	qry := `
		select
			id,
			exp,
			attrs
		from tokens
		where id = $1::text;
	`

	var attrs []byte
	if err := pg.QueryRow(ctx, qry, id).Scan(&out.id, &out.exp, &attrs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return wrapError(err, ErrNotFound)
		}

		return wrapError(err, ErrUnknown)
	}

	if len(attrs) > 0 {
		out.attrs = attrs
	}

	return nil
}
