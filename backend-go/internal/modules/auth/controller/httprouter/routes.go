// Package httprouter
package httprouter

import (
	"github.com/adexcell/crm-ai/internal/config"
	"github.com/adexcell/crm-ai/internal/infra/http/middleware"
	"github.com/adexcell/crm-ai/internal/infra/jwt"
	v1 "github.com/adexcell/crm-ai/internal/modules/auth/controller/httprouter/v1"
	"github.com/go-chi/chi/v5"
)

func AuthRouter(r chi.Router, srv v1.Service, cfg *config.Config, jwtManager *jwt.Manager) {
	ver1 := v1.New(srv, cfg)

	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/email/send-code", ver1.EmailSendCode)
		r.Post("/email/verify-code", ver1.EmailVerifyCode)
		r.Post("/register", ver1.Register)
		r.Post("/register/invite", ver1.Invite)
		r.Post("/login", ver1.Login)
		r.Post("/forgot-password", ver1.ForgotPassword)
		r.Post("/reset-password", ver1.ResetPassword)

		r.Post("/refresh", ver1.RefreshToken)
		r.Post("/logout", ver1.Logout)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtManager))

			r.Get("/me", ver1.GetProfile)
			r.Patch("/me", ver1.UpdateProfile)
			r.Post("/switch-org", ver1.SwitchOrg)
			r.Get("/sessions", ver1.GetSessions)
			r.Post("/sessions/revoke-all", ver1.RevokeAllSessions)
			r.Delete("/sessions/{id}", ver1.RevokeSession)
		})
	})
}
