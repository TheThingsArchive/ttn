package util

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/apex/log"
	"github.com/spf13/viper"
)

type App struct {
	EUI        string   `json:"eui"` // TODO: Change to []string
	Name       string   `json:"name"`
	Owner      string   `json:"owner"`
	AccessKeys []string `json:"accessKeys"`
	Valid      bool     `json:"valid"`
}

func GetApplications(ctx log.Interface) ([]*App, error) {
	server := viper.GetString("ttn-account-server")
	uri := fmt.Sprintf("%s/applications", server)
	req, err := NewRequestWithAuth(server, "GET", uri, nil)
	if err != nil {
		ctx.WithError(err).Fatal("Failed to create authenticated request")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.WithError(err).Fatal("Failed to get applications")
	}
	if resp.StatusCode != http.StatusOK {
		ctx.Fatalf("Failed to get applications: %s", resp.Status)
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var apps []*App
	err = decoder.Decode(&apps)
	if err != nil {
		ctx.WithError(err).Fatal("Failed to read applications")
	}

	return apps, nil
}
