package plugins

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/movio/bramble"
)

func init() {
	bramble.RegisterPlugin(&AuthKelvinGreen{})
}

type AuthKelvinGreenConfig struct {
	TokenLength int `json:"tokenLength"`
}

type AuthKelvinGreen struct {
	bramble.BasePlugin
	config AuthKelvinGreenConfig
}

func (p *AuthKelvinGreen) ID() string {
	return "auth_kelvingreen"
}

func (p *AuthKelvinGreen) ApplyMiddlewarePublicMux(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ab := r.Header.Get("Authorization")
		ctx := r.Context()

		parts := strings.Split(ab, "Bearer ")
		if len(parts) == 2 {
			if p.config.TokenLength > 0 && len(parts[1]) != p.config.TokenLength {
				writeGraphqlError(rw, "invalid authorization")
				return
			}

			ctx = bramble.AddOutgoingRequestsHeaderToContext(ctx, "Authorization", ab)
		}

		h.ServeHTTP(rw, r.WithContext(ctx))
	})
}

func (p *AuthKelvinGreen) Configure(_ *bramble.Config, data json.RawMessage) error {
	err := json.Unmarshal(data, &p.config)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &p.config)
}
