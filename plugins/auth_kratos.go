package plugins

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/movio/bramble"
)

func init() {
	bramble.RegisterPlugin(&AuthKratos{})
}

type AuthKratosConfig struct {
	TokenLength int `json:"tokenLength"`
}

type AuthKratos struct {
	bramble.BasePlugin
	config AuthKratosConfig
}

func (p *AuthKratos) ID() string {
	return "auth_kratos"
}

func (p *AuthKratos) ApplyMiddlewarePublicMux(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ab := r.Header.Get("Authorization")
		ctx := r.Context()

		parts := strings.Split(ab, "Bearer ")
		if len(parts) == 2 {
			if len(parts[1]) != p.config.TokenLength {
				writeGraphqlError(rw, "invalid authorization")
				return
			}

			ctx = bramble.AddOutgoingRequestsHeaderToContext(ctx, "Authorization", ab)
		}

		h.ServeHTTP(rw, r.WithContext(ctx))
	})
}

func (p *AuthKratos) Configure(_ *bramble.Config, data json.RawMessage) error {
	err := json.Unmarshal(data, &p.config)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &p.config)
}
