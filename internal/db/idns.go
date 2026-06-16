package db

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	IdnStatusPending   IdnStatus = 0
	IdnStatusActive    IdnStatus = 1
	IdnStatusSuspended IdnStatus = -1
	IdnStatusDeleted   IdnStatus = -2

	IdnSrcAttrsPwdTypArgon2 IdnSrcAttrsPwdTyp = "argon2"
)

type IdnStatus int32
type IdnSrcAttrsPwdTyp string

func (s IdnStatus) String() string {
	switch s {
	case IdnStatusActive:
		return "active"
	case IdnStatusDeleted:
		return "deleted"
	case IdnStatusPending:
		return "pending"
	case IdnStatusSuspended:
		return "suspended"
	default:
		return "unknown"
	}
}

type Idn struct {
	appId        string
	attrs        IdnAttrs
	federationId *string
	id           string
	login        string
}

func (idn *Idn) AppId() string {
	return idn.appId
}

func (idn *Idn) Attrs() IdnAttrs {
	return idn.attrs
}

func (idn *Idn) Email() string {
	return idn.attrs.Email
}

func (idn *Idn) FederationId() string {
	if idn.federationId == nil {
		return ""
	}

	return *idn.federationId
}

func (idn *Idn) Id() string {
	return idn.id
}

func (idn *Idn) Login() string {
	return idn.login
}

func (idn *Idn) Name() string {
	return idn.attrs.Name
}

func (idn *Idn) Picture() string {
	return idn.attrs.Picture
}

func (idn *Idn) Rev() int32 {
	return idn.attrs.Rev
}

func (idn *Idn) SetAttrs(ctx context.Context, attrs IdnAttrs) error {
	attrs.Rev = idn.Rev() + 1

	qry := `
		update idns
		set
			attrs = $3::jsonb
		where
			id = $1::text
			and attrs['_rev']::integer =  $2::integer
		returning attrs['_rev']::integer;
	`

	if err := pg.QueryRow(ctx, qry, idn.id, idn.Rev(), attrs).Scan(&attrs.Rev); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return wrapError(err, ErrConflict)
		}

		return wrapError(err, ErrUnknown)
	}

	idn.attrs = attrs
	return nil
}

func (idn *Idn) SpaceId() string {
	return idn.attrs.SpaceId
}

func (idn *Idn) Status() IdnStatus {
	return idn.attrs.Status
}

func (idn *Idn) insert(ctx context.Context) error {
	if idn.appId == "" {
		return wrapError(fmt.Errorf("missing app id"), ErrInvalidData)
	}

	idn.id = uuidv7()
	idn.attrs.Rev = rand.Int32()

	qry := `
		insert into idns (
			id,
			app_id,
			login,
			federation_id,
			attrs
		) values (
			$1::text,
			$2::text,
			$3::text,
			$4::text,
			$5::jsonb
		);
	`

	if _, err := pg.Exec(
		ctx,
		qry,
		idn.id,
		idn.appId,
		idn.login,
		idn.federationId,
		idn.attrs,
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

type IdnAttrs struct {
	Email   string    `json:"email,omitempty"`
	Name    string    `json:"name,omitempty"`
	Picture string    `json:"picture,omitempty"`
	SpaceId string    `json:"space_id,omitempty"`
	Status  IdnStatus `json:"status,omitempty"`

	Rev int32 `json:"_rev,omitempty"`
}

type IdnSrc struct {
	attrs      IdnSrcAttrs
	idnId      string
	providerId string
	sub        string
}

func (src *IdnSrc) Cr() IdnSrcAttrsCr {
	if src.attrs.Cr == nil {
		return IdnSrcAttrsCr{}
	}

	return *src.attrs.Cr
}

func (src *IdnSrc) IdnId() string {
	return src.idnId
}

func (src *IdnSrc) ProviderId() string {
	return src.providerId
}

func (src *IdnSrc) Pwd() IdnSrcAttrsPwd {
	if src.attrs.Pwd == nil {
		return IdnSrcAttrsPwd{}
	}

	return *src.attrs.Pwd
}

func (src *IdnSrc) Rev() int32 {
	return src.attrs.Rev
}

func (src *IdnSrc) SetAttrsCr(ctx context.Context, cr IdnSrcAttrsCr) error {
	attrs := src.attrs
	attrs.Rev += 1
	attrs.Cr = &cr

	qry := `
		update idn_srcs
		set
			attrs = $4::jsonb
		where
			idn_id = $1::text
			and provider_id = $2::text
			and attrs['_rev']::integer = $3::integer
		returning attrs['_rev']::integer;
	`

	if err := pg.QueryRow(ctx, qry, src.idnId, src.providerId, src.Rev(), attrs).Scan(&attrs.Rev); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return wrapError(err, ErrConflict)
		}

		return wrapError(err, ErrUnknown)
	}

	src.attrs = attrs
	return nil
}

func (src *IdnSrc) Sub() string {
	return src.sub
}

func (src *IdnSrc) insert(ctx context.Context) error {
	src.attrs.Rev = rand.Int32()

	qry := `
		insert into idn_srcs (
			idn_id,
			provider_id,
			sub,
			attrs
		) values (
			$1::text,
			$2::text,
			$3::text,
			$4::jsonb
		);
	`

	if _, err := pg.Exec(ctx, qry, src.idnId, src.providerId, src.sub, src.attrs); err != nil {
		return wrapError(err, ErrUnknown)
	}

	return nil
}

type IdnSrcAttrs struct {
	Cr  *IdnSrcAttrsCr  `json:"cr,omitempty"`
	Pwd *IdnSrcAttrsPwd `json:"pwd,omitempty"`

	Rev int32 `json:"_rev,omitempty"`
}

type IdnSrcAttrsCr struct {
	// Credential record
	// Ref: https://w3c.github.io/webauthn/#credential-record

	AttestationObject         Base64Url `json:"attestation_object,omitempty"`
	AttestationClientDataJSON Base64Url `json:"attestation_clientdata_json,omitempty"`
	RpId                      string    `json:"rp_id,omitempty"`

	BackupEligible bool      `json:"backup_eligible"`
	BackupState    bool      `json:"backup_state"`
	PublicKey      Base64Url `json:"public_key,omitempty"`
	SignCount      uint32    `json:"sign_count,omitempty"`
	Transports     []string  `json:"transports,omitempty"`
	UvInitialized  bool      `json:"uv_initialized"`
}

type IdnSrcAttrsPwd struct {
	// Password credentials

	Typ     IdnSrcAttrsPwdTyp `json:"typ,omitempty"`
	Salt    []byte            `json:"salt,omitempty"`
	Key     []byte            `json:"key,omitempty"`
	Time    uint32            `json:"time,omitempty"`
	Memory  uint32            `json:"memory,omitempty"`
	Threads uint8             `json:"threads,omitempty"`
}

func FindIdnSrcBySub(ctx context.Context, sub string) (bool, error) {
	qry := `
		select
			count(*)
		from idn_srcs
		where
			sub = $1::text
		limit 1;
	`

	var count int
	if err := pg.QueryRow(ctx, qry, sub).Scan(&count); err != nil {
		return false, wrapError(err, ErrUnknown)
	}

	return (count > 0), nil
}

func ListIdnSrcByProviderIdAndLogin(ctx context.Context, providerId string, login string) ([]IdnSrc, error) {
	qry := `
		select
			s.idn_id,
			s.provider_id,
			s.sub,
			s.attrs
		from idn_srcs s
			inner join idns i on s.idn_id = i.id
		where
			s.provider_id = $1::text
			and i.login = $2::text
		;
	`

	rows, err := pg.Query(ctx, qry, providerId, login)
	if err != nil {
		return nil, wrapError(err, ErrUnknown)
	}

	var srcs []IdnSrc
	for rows.Next() {
		var src IdnSrc
		if err := rows.Scan(&src.idnId, &src.providerId, &src.sub, &src.attrs); err != nil {
			return nil, wrapError(err, ErrUnknown)
		}

		srcs = append(srcs, src)
	}

	return srcs, nil
}

func NewIdn(ctx context.Context, appId string, login string, attrs IdnAttrs) (*Idn, error) {
	idn := &Idn{
		attrs: attrs,
		appId: appId,
		login: login,
	}

	if err := idn.insert(ctx); err != nil {
		return nil, err
	}

	return idn, nil
}

func NewIdnSrc(ctx context.Context, idnId string, providerId string, sub string) (*IdnSrc, error) {
	src := &IdnSrc{
		idnId:      idnId,
		providerId: providerId,
		sub:        sub,
	}

	if err := src.insert(ctx); err != nil {
		return nil, err
	}

	return src, nil
}

func NewIdnSrcWithCr(ctx context.Context, idnId string, providerId string, sub string, cr IdnSrcAttrsCr) (*IdnSrc, error) {
	src := &IdnSrc{
		attrs: IdnSrcAttrs{
			Cr: &cr,
		},
		idnId:      idnId,
		providerId: providerId,
		sub:        sub,
	}

	if err := src.insert(ctx); err != nil {
		return nil, err
	}

	return src, nil
}

func RetrieveIdnByFederatedLogin(ctx context.Context, appId string, login string, iss string, aud string) (*Idn, error) {
	// Note: we need to ensure the query uses unique indexes on both tables
	//   - idns (app_id, login)
	//   - federations (app_id, iss)
	qry := `
		select
			i.id,
			i.app_id,
			i.login,
			i.federation_id,
			i.attrs
		from idns i inner join federations f
			on i.federation_id = f.id and i.app_id = f.app_id
		where
			i.app_id = $1::text
			and i.login = $2::text
			and f.iss = $3::text
			and f.attrs->>'aud' = $4::text
		;
	`

	var idn Idn
	err := pg.QueryRow(ctx, qry, appId, login, iss, aud).Scan(&idn.id, &idn.appId, &idn.login, &idn.federationId, &idn.attrs)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, err
	}

	return &idn, nil
}

func RetrieveIdnByLogin(ctx context.Context, appId string, login string) (*Idn, error) {
	qry := `
		select
			id,
			app_id,
			login,
			federation_id,
			attrs
		from idns
		where
			app_id = $1::text
			and login = $2::text
		;
	`

	idn := Idn{}
	if err := pg.QueryRow(ctx, qry, appId, login).Scan(&idn.id, &idn.appId, &idn.login, &idn.federationId, &idn.attrs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, err
	}

	return &idn, nil
}

func RetrieveIdnByAuthToken(ctx context.Context, token []byte) (*Idn, error) {
	qry := `
		select
			i.id,
			i.app_id,
			i.login,
			i.federation_id,
			i.attrs
		from tokens o inner join idns i
			on o.attrs->>'sub' = i.id
		where
			o.id = $1::text
			and o.exp > $2::timestamp;
	`

	tokenId := authTokenPrefix + encoding.EncodeToString(token)
	var idn Idn
	if err := pg.QueryRow(ctx, qry, tokenId, time.Now().UTC()).Scan(&idn.id, &idn.appId, &idn.login, &idn.federationId, &idn.attrs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, wrapError(err, ErrUnknown)
	}

	return &idn, nil
}

func RetrieveIdnBySrc(ctx context.Context, providerId string, sub string) (*Idn, error) {
	qry := `
		select
			i.id,
			i.app_id,
			i.login,
			i.federation_id,
			i.attrs
		from idns i inner join idn_srcs s
			on i.id = s.idn_id
		where
			s.provider_id = $1::text
			and s.sub = $2::text
		;
	`

	idn := Idn{}
	if err := pg.QueryRow(ctx, qry, providerId, sub).Scan(&idn.id, &idn.appId, &idn.login, &idn.federationId, &idn.attrs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, err
	}

	return &idn, nil
}

func RetrieveIdnSrc(ctx context.Context, providerId string, sub string) (*IdnSrc, error) {
	qry := `
		select
			idn_id,
			provider_id,
			sub,
			attrs
		from idn_srcs
		where
			provider_id = $1::text
			and sub = $2::text
		;
	`

	var src IdnSrc
	if err := pg.QueryRow(ctx, qry, providerId, sub).Scan(&src.idnId, &src.providerId, &src.sub, &src.attrs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, wrapError(err, ErrNotFound)
		}

		return nil, wrapError(err, ErrUnknown)
	}

	return &src, nil
}
