package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	catalogURI          = "rsk://catalog"
	catalogMIME         = "application/json"
	descriptionLimit    = 200
	descriptionEllipsis = "…"
)

// catalogEntry is one skill in the rsk://catalog resource payload. The full
// version list is omitted: search and profile need only the latest.
type catalogEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Latest      string `json:"latest"`
	Personal    bool   `json:"personal"`
}

func registerCatalogResource(srv *sdk.Server, deps Deps) {
	srv.AddResource(&sdk.Resource{
		URI:   catalogURI,
		Name:  "rsk-catalog",
		Title: "ralvaskills catalog",
		Description: "Catálogo de skills disponibles en ralvaskills. Descripciones " +
			"recortadas; usa search_skills para filtrar por proyecto.",
		MIMEType: catalogMIME,
	}, catalogHandler(deps))
}

func catalogHandler(deps Deps) sdk.ResourceHandler {
	return func(ctx context.Context, _ *sdk.ReadResourceRequest) (*sdk.ReadResourceResult, error) {
		entries, err := deps.catalog(ctx)
		if err != nil {
			return nil, fmt.Errorf("read catalog: %w", err)
		}

		payload := make([]catalogEntry, 0, len(entries))
		for _, e := range entries {
			payload = append(payload, catalogEntry{
				Name:        e.Name,
				Description: truncate(e.Description, descriptionLimit),
				Latest:      e.Latest,
				Personal:    e.Personal,
			})
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal catalog: %w", err)
		}

		return &sdk.ReadResourceResult{Contents: []*sdk.ResourceContents{{
			URI:      catalogURI,
			MIMEType: catalogMIME,
			Text:     string(data),
		}}}, nil
	}
}

// truncate caps s at limit runes, appending an ellipsis when it cuts. It counts
// runes, not bytes, so it never splits a multi-byte character.
func truncate(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit]) + descriptionEllipsis
}
