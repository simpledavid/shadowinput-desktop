package application

import (
	"arion.local/core/domain"
	"arion.local/core/domain/ports"
	"context"
)

type Writer struct{ Generator ports.WritingGenerator }

func (w Writer) Write(ctx context.Context, request domain.WritingRequest, preview func(string)) (string, error) {
	if err := request.Validate(); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return w.Generator.Generate(ctx, request, preview)
}
