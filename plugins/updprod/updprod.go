package updprod

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/infinitybotlist/sysmanage-web/core/logger"
	"github.com/infinitybotlist/sysmanage-web/core/state"
)

func updateProdTask(taskId string) {
	defer logger.LogMap.MarkDone(taskId)

	logger.LogMap.Add(taskId, "Going to update production...", true)

	// Look for .protect-deploy, if it exists, exit

	f, err := os.ReadFile("/.protect-deploy")

	if err == nil {
		logger.LogMap.Add(taskId, "FATAL: Deploy protected: "+string(f), true)
		logger.LogMap.Add(taskId, "", true)
		logger.LogMap.Add(taskId, "Please do not attempt to manually override this. The protection is most likely for a reason", true)
		return
	}

	logger.LogMap.Add(taskId, "", true)
	logger.LogMap.Add(taskId, "=> Deleting old production branch", true)

	var bpDisable bool // Whether branch protection is working on the Github repo

	client := http.Client{
		Timeout: time.Second * 10,
	}

	// Disable enforce_admin
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"https://api.github.com/repos/%s/branches/production/protection/enforce_admins",
			GitRepo,
		),
		nil,
	)

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to create request: "+err.Error(), true)
		return
	}

	req.SetBasicAuth(GithubUsername, state.Config.GithubPat)
	req.Header.Set("User-Agent", GithubUsername)

	resp, err := client.Do(req)

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to make request to enforce_admin: "+err.Error(), true)
	}

	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		// Try reading body as well to check for github pro
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			logger.LogMap.Add(taskId, "FATAL: Failed to read response body: "+err.Error(), true)
			return
		}

		// HACK but it works for our case
		if strings.Contains(string(body), "GitHub Pro") {
			logger.LogMap.Add(taskId, "WARNING: Branch protection is disabled due to Github Pro subscription expiry: "+string(body), true)
			bpDisable = true
		} else {
			logger.LogMap.Add(taskId, "FATAL: Failed to disable enforce_admin: "+string(body), true)
			return
		}
	}

	logger.LogMap.Add(taskId, "SUCCESS: Disabled enforce_admins", true)

	if !bpDisable {
		if ok := deleteBranchProtection(taskId, client); !ok {
			return
		}

		logger.LogMap.Add(taskId, "SUCCESS: Deleted existing branch protection", true)
	}

	// Delete old production branch
	if ok := deleteProdBranch(taskId, client); !ok {
		return
	}

	logger.LogMap.Add(taskId, "SUCCESS: Deleted old production branch", true)

	// Create new production branch
	logger.LogMap.Add(taskId, "", true)
	logger.LogMap.Add(taskId, "=> Creating new production branch", true)

	sha := getMainBranchSHA(taskId, client)

	if sha == "" {
		return
	}

	logger.LogMap.Add(taskId, "Main branch SHA: "+sha, true)

	if ok := createProdBranch(taskId, sha, client); !ok {
		return
	}

	logger.LogMap.Add(taskId, "SUCCESS: Created new production branch", true)

	// Create new production branch protection
	if !bpDisable {
		if ok := createBranchProtection(taskId, client); !ok {
			return
		}

		logger.LogMap.Add(taskId, "SUCCESS: Created branch protection", true)

		// Admin enforce
		if ok := enableAdminEnforce(taskId, client); !ok {
			return
		}

		logger.LogMap.Add(taskId, "SUCCESS: Enabled enforce_admins", true)
	}

	logger.LogMap.Add(taskId, "", true)
	logger.LogMap.Add(taskId, "SUCCESS", true)
}
