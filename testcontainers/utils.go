package testcontainer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cpuguy83/dockercfg"
)

func registryCred() {
	host, ok := os.LookupEnv("REGISTRY_HOST")
	if !ok {
		return
	}
	userName := os.Getenv("REGISTRY_USER")
	password := os.Getenv("REGISTRY_PASSWORD")
	authStr := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", userName, password)))

	cfg := dockercfg.Config{}
	cfg.AuthConfigs = map[string]dockercfg.AuthConfig{host: dockercfg.AuthConfig{
		Username: userName,
		Password: password,
		Auth:     authStr,
	}}

	data, _ := json.Marshal(cfg)

	_ = os.Setenv("DOCKER_AUTH_CONFIG", string(data))
}

func setContainerPortByEnv(internalPort string, envKey string) string {
	if env := os.Getenv(envKey); env != "" {
		return env + ":" + internalPort
	}
	return internalPort
}
