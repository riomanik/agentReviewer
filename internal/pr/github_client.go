package pr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type PRFile struct {
	Filename string `json:"filename"`
	Patch    string `json:"patch"`
}

func FetchPRFiles(info PRInfo) ([]PRFile, error) {
	token := os.Getenv("GIT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GIT_TOKEN not set")
	}

	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/pulls/%d/files",
		info.Owner,
		info.Repo,
		info.Number,
	)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// ⛔️ PENTING
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errResp)

		return nil, fmt.Errorf(
			"github error: status=%d message=%v",
			resp.StatusCode,
			errResp["message"],
		)
	}

	var files []PRFile
	err = json.NewDecoder(resp.Body).Decode(&files)
	return files, err

}
