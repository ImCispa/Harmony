package autentication

import (
	"context"
	"log"
	"os"

	"github.com/coreos/go-oidc"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/oauth2"
)

type Service struct {
	Oauth2Config *oauth2.Config
	OidcVerifier *oidc.IDTokenVerifier
}

var (
	clientID     = os.Getenv("AUTH_CLIENT_ID")
	clientSecret = os.Getenv("AUTH_CLIENT_SECRET")
	redirectURL  = os.Getenv("AUTH_REDIRECT_URL")
	providerURL  = os.Getenv("AUTH_PROVIDER_URL")
)

func New() Service {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		log.Fatalf("Failed to get provider: %v", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	oidcVerifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return Service{
		Oauth2Config: oauth2Config,
		OidcVerifier: oidcVerifier,
	}
}
