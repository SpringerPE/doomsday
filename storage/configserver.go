package storage

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/credhub-cli/credhub"
	"doomsday/storage/uaa"
)

type ConfigServerAccessor struct {
	credhub *credhub.CredHub
}

type ConfigServerConfig struct {
	Address            string `yaml:"address"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	Auth               struct {
		GrantType    string `yaml:"grant_type"`
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
	} `yaml:"auth"`
}

func newConfigServerAccessor(conf ConfigServerConfig) (*ConfigServerAccessor, error) {
	var err error
	var authResp *uaa.AuthResponse

	c, err := credhub.New(
		conf.Address,
		credhub.SkipTLSValidation(conf.InsecureSkipVerify),
	)
	if err != nil {
		return nil, fmt.Errorf("Could not create config server client: %s", err)
	}

	authURL, err := c.AuthURL()
	if err != nil {
		return nil, fmt.Errorf("Could not get auth endpoint: %s", err)
	}

	fmt.Printf("AuthURL: %s\n", authURL)

	uaaClient := uaa.Client{
		URL:               authURL,
		SkipTLSValidation: conf.InsecureSkipVerify,
	}

	var isClientCredentials bool

	switch conf.Auth.GrantType {
	case "client_credentials", "client credentials", "clientcredentials":
		fmt.Println("Performing client credentials grant auth")
		isClientCredentials = true
		authResp, err = uaaClient.ClientCredentials(
			conf.Auth.ClientID,
			conf.Auth.ClientSecret,
		)

	case "resource_owner", "resource owner", "resourceowner", "password":
		fmt.Println("Performing password grant auth")
		authResp, err = uaaClient.Password(
			conf.Auth.ClientID,
			conf.Auth.ClientSecret,
			conf.Auth.Username,
			conf.Auth.Password,
		)

	case "none", "noop": //The default is the noop builder
	default:
		err = fmt.Errorf("Unknown auth grant_type `%s'", conf.Auth.GrantType)
	}
	if err != nil {
		return nil, err
	}

	fmt.Println("Auth complete")

	c, err = credhub.New(
		conf.Address,
		credhub.SkipTLSValidation(conf.InsecureSkipVerify),
	)

	c.Auth = &refreshTokenStrategy{
		ClientID:            conf.Auth.ClientID,
		ClientSecret:        conf.Auth.ClientSecret,
		UAAClient:           &uaaClient,
		APIClient:           c.Client(),
		IsClientCredentials: isClientCredentials,
	}

	c.Auth.(*refreshTokenStrategy).SetTokens(authResp.AccessToken, authResp.RefreshToken)

	go func() {
		for range time.Tick(authResp.TTL / 2) {
			err = c.Auth.(*refreshTokenStrategy).Refresh()
			if err != nil {
				fmt.Printf("Could not refresh token: %s", err)
			}
		}
	}()

	return &ConfigServerAccessor{credhub: c}, nil
}

//List attempts to get all of the paths in the config server
func (v *ConfigServerAccessor) List() (PathList, error) {
	paths, err := v.credhub.FindByPath("/")
	if err != nil {
		return nil, fmt.Errorf("Could not get paths in config server: %s", err)
	}

	ret := make(PathList, 0, len(paths.Credentials))
	for _, entry := range paths.Credentials {
		ret = append(ret, entry.Name)
	}

	return ret, nil
}

func (v *ConfigServerAccessor) Get(path string) (map[string]string, error) {
	cred, err := v.credhub.GetLatestVersion(path)
	if err != nil {
		return nil, err
	}

	if cred.Type == "certificate" {
		if certInterface, found := cred.Value.(map[string]interface{})["certificate"]; found {
			return map[string]string{"certificate": certInterface.(string)}, nil
		}
	}

	return nil, nil
}

type refreshTokenStrategy struct {
	lock                sync.RWMutex
	accessToken         string
	refreshToken        string
	ClientID            string
	ClientSecret        string
	IsClientCredentials bool
	UAAClient           *uaa.Client
	APIClient           *http.Client
}

func (r *refreshTokenStrategy) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+r.AccessToken())
	return r.APIClient.Do(req)
}

func (r *refreshTokenStrategy) AccessToken() string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.accessToken
}

func (r *refreshTokenStrategy) RefreshToken() string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.refreshToken
}

func (r *refreshTokenStrategy) Refresh() error {
	var authResp *uaa.AuthResponse
	var err error
	if r.IsClientCredentials {
		authResp, err = r.UAAClient.ClientCredentials(r.ClientID, r.ClientSecret)
	} else {
		authResp, err = r.UAAClient.Refresh(r.ClientID, r.ClientSecret, r.RefreshToken())
	}

	if err != nil {
		return err
	}

	r.SetTokens(authResp.AccessToken, authResp.RefreshToken)

	return nil
}

func (r *refreshTokenStrategy) SetTokens(accessToken, refreshToken string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.accessToken = accessToken
	r.refreshToken = refreshToken
}
