package ci

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/gdsoumya/better_ci/utils"

	"github.com/google/go-github/v32/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

type Config struct {
	ctx        context.Context
	client     *github.Client
	dockeru    string
	dockerp    string
	websec     string
	Port       string
	host       string
	permission []string
}

func Init() (Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("CANNOT READ ENV DATA")
	}
	var c Config
	accessKey, exists := os.LookupEnv("ACCESS_KEY")
	if !exists {
		log.Fatal("ERROR ENV: ACCESS KEY MISSING")
	}
	c.dockeru, exists = os.LookupEnv("DOCKER_USER")
	if !exists {
		log.Fatal("ERROR ENV: DOCKER USERNAME MISSING")
	}
	c.dockerp, exists = os.LookupEnv("DOCKER_PASS")
	if !exists {
		log.Fatal("ERROR ENV: DOCKER PASSWORD MISSING")
	}
	c.websec, exists = os.LookupEnv("WEBHOOK_SECRET")
	if !exists {
		log.Fatal("ERROR ENV: WEBHOOK SECRET MISSING, REMEMBER TO SECURE YOUR HOOKS ALWAYS")
	}
	c.Port, exists = os.LookupEnv("PORT")
	if !exists {
		log.Fatal("ERROR ENV: PORT FOR CI MISSING")
	}
	c.host, exists = os.LookupEnv("HOST")
	if !exists {
		c.host, err = utils.GetPublicIP()
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	permission, exist := os.LookupEnv("AUTHOR_PERMISSION")
	if !exist {
		log.Fatal("ERROR ENV: AUTHOR PERMISSION FOR CI MISSING")
	} else {
		c.permission = strings.Split(permission, " ")
	}
	log.Print("CI HOST : ", c.host)
	log.Print("WEBHOOK URL : http://", c.host, ":", c.Port, "/webhook")
	c.ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessKey},
	)
	tc := oauth2.NewClient(c.ctx, ts)

	c.client = github.NewClient(tc)
	return c, nil
}
