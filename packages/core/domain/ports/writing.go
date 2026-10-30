package ports

import (
	"arion.local/core/domain"
	"context"
)

type WritingGenerator interface {
	Generate(context.Context, domain.WritingRequest, func(string)) (string, error)
}
