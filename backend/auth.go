package backend

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/avatar"
	"github.com/go-pkgz/auth/provider"
	"github.com/go-pkgz/auth/token"
)

type Auth struct {
	Service   *auth.Service
	Providers []string
}

func NewAuth() *Auth {
	auth := &Auth{}
	auth.newService()

	return auth
}

func (a *Auth) newService() {
	options := auth.Opts{
		SecretReader: token.SecretFunc(func(_ string) (string, error) {
			return "qwerty1234", nil
		}),
		TokenDuration:  time.Minute,
		CookieDuration: time.Hour * 24,
		Issuer:         "secretshare",
		URL:            "http://localhost:8000",
		Validator:      token.ValidatorFunc(a.tokenValidator),
		AvatarStore:    avatar.NewNoOp(),
		UseGravatar:    false,
	}

	a.Service = auth.NewService(options)

	githubClientID := os.Getenv("APP_GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("APP_GITHUB_CLIENT_SECRET")
	if githubClientID != "" && githubClientSecret != "" {
		a.Service.AddProvider("github", githubClientID, githubClientSecret)
		a.Providers = append(a.Providers, "github")
	}

	if token := os.Getenv("APP_TELEGRAM_TOKEN"); token != "" {
		telegram := provider.TelegramHandler{
			ProviderName: "telegram",
			ErrorMsg:     "❌ Invalid auth request. Please try clicking link again.",
			SuccessMsg:   "✅ You have successfully authenticated!",
			Telegram:     provider.NewTelegramAPI(token, http.DefaultClient),
			TokenService: a.Service.TokenService(),
			AvatarSaver:  a.Service.AvatarProxy(),
		}

		go func() {
			if err := telegram.Run(context.Background()); err != nil {
				log.Fatal(err)
			}
		}()

		a.Service.AddCustomHandler(&telegram)
		a.Providers = append(a.Providers, "telegram")
	}
}

func (a *Auth) tokenValidator(token string, claims token.Claims) bool {
	if claims.User != nil {
		log.Printf("user: %#v\n", claims.User)
		if strings.HasPrefix(claims.User.ID, "github_") {
			return true
		}
	}
	return false
}
