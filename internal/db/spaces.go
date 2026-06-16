package db

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Space struct {
	attrs SpaceAttrs
	id    string
	slug  string
}

func (s *Space) Attrs() SpaceAttrs {
	return s.attrs
}

func (s *Space) Id() string {
	return s.id
}

func (s *Space) Rev() int32 {
	return s.attrs.Rev
}

func (s *Space) Slug() string {
	return s.slug
}

func (s *Space) insert(ctx context.Context) error {
	if err := s.validate(); err != nil {
		return err
	}

	s.id = uuidv7()
	s.attrs.Rev = rand.Int32()

	qry := `
		insert into spaces (
			id,
			slug,
			attrs
		) values (
			$1::text,
			$2::text,
			$3::jsonb
		);
	`

	if _, err := pg.Exec(ctx, qry,
		s.id,
		s.slug,
		s.attrs,
	); err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			switch pgxErr.Code {
			case "23505": // unique violation
				return wrapError(err, ErrConflict)
			default:
				return wrapError(err, ErrUnknown)
			}
		}

		return wrapError(err, ErrUnknown)
	}

	return nil
}

func (s *Space) validate() error {
	if !rxSpaceSlug().MatchString(s.slug) {
		return wrapError(fmt.Errorf("invalid slug"), ErrInvalidData)
	}

	if rxSpaceSlugReserved().MatchString(s.slug) {
		return wrapError(fmt.Errorf("reserved words in slug"), ErrInvalidData)
	}

	if strings.Contains(s.slug, "--") {
		return wrapError(fmt.Errorf("invalid slug"), ErrInvalidData)
	}

	return nil
}

type SpaceAttrs struct {
	Rev int32 `json:"_rev,omitempty"`
}

func NewSpace(ctx context.Context, slug string, attrs SpaceAttrs) (*Space, error) {
	space := &Space{
		attrs: attrs,
		slug:  slug,
	}

	if err := space.insert(ctx); err != nil {
		return nil, err
	}

	return space, nil
}

func RetrieveSpace(ctx context.Context, id string) (*Space, error) {
	qry := `
		select
			id,
			slug,
			attrs
		from spaces
		where id = $1::text;
	`

	var space Space
	if err := pg.QueryRow(ctx, qry, id).Scan(&space.id, &space.slug, &space.attrs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, err
	}

	return &space, nil
}

func RetrieveSpaceBySlug(ctx context.Context, slug string) (*Space, error) {
	qry := `
		select
			id,
			slug,
			attrs
		from spaces
		where slug = $1::text;
	`

	var space Space
	if err := pg.QueryRow(ctx, qry, slug).Scan(&space.id, &space.slug, &space.attrs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, err
	}

	return &space, nil
}
