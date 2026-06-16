package srv

import (
	"context"
	"errors"
	"fmt"

	"github.com/felk-ai/idaas/internal/db"
	"github.com/felk-ai/idaas/pb"
)

type spaces struct {
	pb.UnimplementedSpacesServer
}

func (s *spaces) Create(ctx context.Context, req *pb.SpacesCreateRequest) (*pb.SpacesCreateResponse, error) {
	token, err := accessTokenFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	idn, err := checkAccessToken(ctx, token)
	if err != nil {
		if errors.Is(err, errInvalidAccessToken) {
			return nil, NewError(ErrPermissionDenied, err)
		}

		if errors.Is(err, errInvalidAccessTokenPrefix) {
			// TODO: check for federated tokens
			return nil, NewErrorf(ErrUnauthenticated, "invalid or unsupported credentials")
		}

		return nil, err
	}

	if idn.Status() < db.IdnStatusPending {
		return nil, NewError(
			ErrPermissionDenied,
			fmt.Errorf("forbidden due to identity status (%v)", idn.Status()),
		)
	}

	if idn.SpaceId() != "" {
		return nil, NewErrorf(ErrResourceExhausted, "quota exceeded")
	}

	space, err := db.NewSpace(ctx, req.GetSlug(), db.SpaceAttrs{})
	if err != nil {
		if errors.Is(err, db.ErrInvalidData) {
			return nil, ErrInvalidData
		}

		if errors.Is(err, db.ErrConflict) {
			return nil, ErrAlreadyExists
		}

		return nil, err
	}

	// Due to the lack of access control, we assign the newly created space to the request actor
	// and activate the actor identity (if pending).
	attrs := idn.Attrs()
	attrs.SpaceId = space.Id()
	if attrs.Status == db.IdnStatusPending {
		attrs.Status = db.IdnStatusActive
	}
	if err = idn.SetAttrs(ctx, attrs); err != nil {
		return nil, err
	}

	return &pb.SpacesCreateResponse{
		Space: mapSpaceToPb(space),
	}, nil
}

func mapSpaceToPb(space *db.Space) *pb.Space {
	return &pb.Space{
		Id:   space.Id(),
		Slug: space.Slug(),
	}
}
