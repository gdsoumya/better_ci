package ci

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

type Config struct {
	ctx     context.Context
	client  *github.Client
	dockeru string
	dockerp string
	websec  string
}

func Init() (Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("CANNOT READ ENV DATA")
		//return Config{}, err
	}
	var c Config
	accessKey, exists := os.LookupEnv("ACCESS_KEY")
	if !exists {
		log.Fatal("ERROR ENV: ACCESS KEY MISSING")
		//return Config{}, errors.New("ERROR ENV: ACCESS KEY MISSING")
	}
	c.dockeru, exists = os.LookupEnv("DOCKER_USER")
	if !exists {
		log.Fatal("ERROR ENV: DOCKER USERNAME MISSING")
		//return Config{}, errors.New("ERROR ENV: DOCKER USERNAME MISSING")
	}
	c.dockerp, exists = os.LookupEnv("DOCKER_PASS")
	if !exists {
		log.Fatal("ERROR ENV: DOCKER PASSWORD MISSING")
		//return Config{}, errors.New("ERROR ENV: DOCKER PASSWORD MISSING")
	}
	c.websec, exists = os.LookupEnv("WEBHOOK_SECRET")
	if !exists {
		log.Fatal("ERROR ENV: WEBHOOK SECRET MISSING, REMEMBER TO SECURE YOUR HOOKS ALWAYS")
		//return Config{}, errors.New("ERROR ENV: DOCKER PASSWORD MISSING")
	}
	c.ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessKey},
	)
	tc := oauth2.NewClient(c.ctx, ts)

	c.client = github.NewClient(tc)
	return c, nil
}
