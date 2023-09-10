package env

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"github.com/namsral/flag"
)

var (
	Port           int
	PublicURL      string
	Env            string
	CommitLongSha  string
	CommitShortSha string
	Version        string
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading env vars: %s", err)
	}
	flag.StringVar(&PublicURL, "PUBLIC_URL", "delphi.market", "Public URL of website")
	flag.IntVar(&Port, "PORT", 4321, "Server port")
	flag.StringVar(&Env, "ENV", "development", "Specify for which environment files should be built")
	flag.Parse()
	CommitLongSha = execCmd("git", "rev-parse", "HEAD")
	CommitShortSha = execCmd("git", "rev-parse", "--short", "HEAD")
	Version = fmt.Sprintf("v0.0.0+%s", CommitShortSha)
}

func execCmd(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(stdout))
}
