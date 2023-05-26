package updprod

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/infinitybotlist/sysmanage-web/core/logger"
	"github.com/infinitybotlist/sysmanage-web/core/state"
)

func authReq(client http.Client, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(GithubUsername, state.Config.GithubPat)
	req.Header.Set("User-Agent", GithubUsername)

	return client.Do(req)
}

func deleteBranchProtection(taskId string, client http.Client) bool {
	// Remove branch protection
	resp, err := authReq(
		client,
		"DELETE",
		fmt.Sprintf(
			"https://api.github.com/repos/%s/branches/production/protection",
			GitRepo,
		),
		nil,
	)

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to make request to disable branch protection: "+err.Error(), true)
	}

	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			logger.LogMap.Add(taskId, "FATAL: Failed to read response body: "+err.Error(), true)
			return false
		}

		logger.LogMap.Add(taskId, "FATAL: Failed to disable branch protection: "+string(body), true)
		return false
	}

	return true
}

func deleteProdBranch(taskId string, client http.Client) bool {
	// Delete old production branch
	resp, err := authReq(
		client,
		"DELETE",
		fmt.Sprintf(
			"https://api.github.com/repos/%s/git/refs/heads/production",
			GitRepo,
		),
		nil,
	)

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to make request to delete the branch: "+err.Error(), true)
	}

	if resp.StatusCode >= 400 && resp.StatusCode != 404 && resp.StatusCode != 422 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			logger.LogMap.Add(taskId, "FATAL: Failed to read response body: "+err.Error(), true)
			return false
		}

		logger.LogMap.Add(taskId, "FATAL: Failed to delete branch: "+string(body), true)
		return false
	}

	if resp.StatusCode == 422 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			logger.LogMap.Add(taskId, "FATAL: Failed to read response body: "+err.Error(), true)
			return false
		}

		logger.LogMap.Add(taskId, "WARNING: Branch does not exist, ignoring error: "+string(body), true)
		return true
	}

	return true
}

func getMainBranchSHA(taskID string, client http.Client) string {
	resp, err := authReq(
		client,
		"GET",
		fmt.Sprintf(
			"https://api.github.com/repos/%s/git/refs/heads/master",
			GitRepo,
		),
		nil,
	)

	if err != nil {
		logger.LogMap.Add(taskID, "FATAL: Failed to make request to get SHA of master branch: "+err.Error(), true)
		return ""
	}

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			logger.LogMap.Add(taskID, "FATAL: Failed to read response body: "+err.Error(), true)
			return ""
		}

		logger.LogMap.Add(taskID, "FATAL: Failed to get SHA of master branch: "+string(body), true)
		return ""
	}

	var respData struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}

	err = json.NewDecoder(resp.Body).Decode(&respData)

	if err != nil {
		logger.LogMap.Add(taskID, "FATAL: Failed to decode response body: "+err.Error(), true)
		return ""
	}

	return respData.Object.SHA
}

func createProdBranch(taskId string, sha string, client http.Client) bool {
	body, err := json.Marshal(map[string]string{
		"ref": "refs/heads/production",
		"sha": sha,
	})

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to marshal request body: "+err.Error(), true)
		return false
	}

	resp, err := authReq(
		client,
		"POST",
		fmt.Sprintf(
			"https://api.github.com/repos/%s/git/refs",
			GitRepo,
		),
		bytes.NewReader(body),
	)

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to make request to create branch: "+err.Error(), true)
	}

	if resp.StatusCode > 400 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			logger.LogMap.Add(taskId, "FATAL: Failed to read response body: "+err.Error(), true)
			return false
		}

		logger.LogMap.Add(taskId, "FATAL: Failed to create branch: "+string(body), true)
		return false
	}

	logger.LogMap.Add(taskId, "SUCCESS: Created new production branch", true)

	return true
}

func createBranchProtection(taskId string, client http.Client) bool {
	body, err := json.Marshal(map[string]any{
		"required_status_checks": nil,
		"enforce_admins":         true,
		"required_pull_request_reviews": map[string]any{
			"dismissal_restrictions": map[string]any{
				"users": []any{},
				"teams": []any{},
			},
			"dismiss_stale_reviews":           true,
			"require_code_owner_reviews":      true,
			"required_approving_review_count": 1,
		},
		"restrictions": map[string]any{
			"users": []any{},
			"teams": []any{},
			"apps":  []any{},
		},
		"allow_deletions": false,
		"block_creations": false,
		"lock_branch":     true,
	})

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to marshal request body: "+err.Error(), true)
		return false
	}

	resp, err := authReq(
		client,
		"PUT",
		fmt.Sprintf(
			"https://api.github.com/repos/%s/branches/production/protection",
			GitRepo,
		),
		bytes.NewReader(body),
	)

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to make request to create branch protection: "+err.Error(), true)
		return false
	}

	if resp.StatusCode > 400 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			logger.LogMap.Add(taskId, "FATAL: Failed to read response body: "+err.Error(), true)
			return false
		}

		logger.LogMap.Add(taskId, "FATAL: Failed to create branch protection: "+string(body), true)
		return false
	}

	logger.LogMap.Add(taskId, "SUCCESS: Created branch protection", true)

	return true
}

func enableAdminEnforce(taskId string, client http.Client) bool {
	resp, err := authReq(
		client,
		"POST",
		fmt.Sprintf(
			"https://api.github.com/repos/%s/branches/production/protection/enforce_admins",
			GitRepo,
		),
		nil,
	)

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to make request to enable enforce_admin: "+err.Error(), true)
		return false
	}

	if resp.StatusCode > 400 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			logger.LogMap.Add(taskId, "FATAL: Failed to read response body: "+err.Error(), true)
			return false
		}

		logger.LogMap.Add(taskId, "FATAL: Failed to enable enforce_admin: "+string(body), true)
		return false
	}

	logger.LogMap.Add(taskId, "SUCCESS: Enabled enforce_admins", true)

	return true
}

func ackVercelDeployHook(taskId string, client http.Client) bool {
	// Ack Vercel deploy hook
	req, err := http.NewRequest(
		"POST",
		VercelDeployHook,
		nil,
	)

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to create request: "+err.Error(), true)
	}

	req.Header.Set("User-Agent", GithubUsername)

	resp, err := client.Do(req)

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to make request to Vercel: "+err.Error(), true)
		return false
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		logger.LogMap.Add(taskId, "FATAL: Failed to read response body: "+err.Error(), true)
		return false
	}

	if resp.StatusCode >= 400 {
		logger.LogMap.Add(taskId, "FATAL: Failed to ack Vercel deploy hook: "+string(body), true)
		return false
	}

	logger.LogMap.Add(taskId, "SUCCESS: Acked Vercel deploy hook: "+string(body), true)

	return true
}
