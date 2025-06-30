package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/gitslim/go-ragger/internal/util"
	"github.com/gitslim/go-ragger/internal/web/errs"
	"github.com/gitslim/go-ragger/internal/web/tpl"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	datastar "github.com/starfederation/datastar/sdk/go"
	"golang.org/x/crypto/bcrypt"
)

func SetupAuthRoutes(r chi.Router, log *slog.Logger, db *sqlc.Queries, sessionStore sessions.Store) {

	r.Route("/auth", func(authRouter chi.Router) {
		authRouter.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
			sess, err := sessionStore.Get(r, util.SessionKey)
			if err != nil {
				http.Error(w, "failed to get session", http.StatusInternalServerError)
				return
			}

			delete(sess.Values, util.UserIDKey)
			if err := sess.Save(r, w); err != nil {
				http.Error(w, "failed to save session", http.StatusInternalServerError)
				return
			}

			sse := datastar.NewSSE(w, r)
			sse.Redirect("/auth/login")
		})

		authRouter.Route("/login", func(loginRouter chi.Router) {
			loginRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
				if _, ok := util.UserFromContext(r.Context()); ok {
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}

				signals := &LoginSignals{}

				PageAuthenticationLogin(r, signals, nil).Render(r.Context(), w)
			})

			loginRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {
				var appError error

				// if err := r.ParseMultipartForm(1 << 20); err != nil {
				// 	http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
				// 	return
				// }

				signals := &LoginSignals{}
				if err := datastar.ReadSignals(r, &signals); err != nil {
					http.Error(w, fmt.Sprintf("error unmarshalling form: %s", err), http.StatusBadRequest)
				}

				user, err := db.GetUserByEmail(r.Context(), signals.Email)
				if err != nil {
					appError = errs.ErrBadCredentials
				}

				err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(signals.Password))
				if err != nil {
					appError = errs.ErrBadCredentials
				} else {
					sess, err := sessionStore.Get(r, util.SessionKey)
					if err != nil {
						http.Error(w, "failed to get session", http.StatusInternalServerError)
						return
					}

					sess.Values[util.UserIDKey] = user.ID
					if err := sess.Save(r, w); err != nil {
						http.Error(w, "failed to save session", http.StatusInternalServerError)
						return
					}
				}

				sse := datastar.NewSSE(w, r)

				errs.ShowErrors(sse)
				if appError != nil {
					errs.ShowErrors(sse, appError)
				} else {
					sse.Redirect("/")
				}
			})
		})

		authRouter.Route("/register", func(registerRouter chi.Router) {

			registerRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
				if _, ok := util.UserFromContext(r.Context()); ok {
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}

				signals := &RegisterSignals{}

				PageAuthenticationRegister(r, signals, nil).Render(r.Context(), w)
			})

			registerRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {
				if _, ok := util.UserFromContext(r.Context()); ok {
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}

				signals := &RegisterSignals{}
				if err := datastar.ReadSignals(r, &signals); err != nil {
					http.Error(w, fmt.Sprintf("error unmarshalling form: %s", err), http.StatusBadRequest)
				}

				email := strings.TrimSpace(signals.Email)
				password := strings.TrimSpace(signals.Password)

				sse := datastar.NewSSE(w, r)

				validationErrors := []error{}
				appendAndSendValidationErrors := func(errs ...error) {
					validationErrors = append(validationErrors, errs...)
					ec := tpl.ErrorMessages(validationErrors...)
					sse.MergeFragmentTempl(ec)
				}
				appendAndSendValidationErrors()

				if email == "" {
					appendAndSendValidationErrors(errors.New("Email обязателен"))
				}
				if len(password) < 8 {
					appendAndSendValidationErrors(errors.New("Минимальная длина пароля 8 символов"))
				}

				if len(validationErrors) > 0 {
					return
				}

				passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					http.Error(w, "failed to create user", http.StatusInternalServerError)
					return
				}

				_, err = db.CreateUserIfNotExists(r.Context(), sqlc.CreateUserIfNotExistsParams{Email: email,
					PasswordHash: string(passwordHash)})
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						appendAndSendValidationErrors(errors.New("Пользователь с таким email уже зарегистрирован"))
					} else {
						http.Error(w, "failed to create user", http.StatusInternalServerError)
						return
					}
				}

				if len(validationErrors) > 0 {
					return
				}

				sse.Redirect("/auth/login")
			})
		})

	})
}
