package placeholder_uuid_generator

import (
	"context"
	"github.com/google/uuid"
)

type Generator struct{}

func (g *Generator) GenerateUUID(_ context.Context) string {
	return uuid.New().String()
}
