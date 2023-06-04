package updprod

import (
	"net/http"

	"github.com/infinitybotlist/eureka/crypto"
	"github.com/infinitybotlist/sysmanage-web/core/plugins"
	"github.com/infinitybotlist/sysmanage-web/plugins/actions"
	"github.com/infinitybotlist/sysmanage-web/types"
)

var GitRepo string
var GithubUsername string
var VercelDeployHook string

func InitPlugin(c *types.PluginConfig) error {
	opts, err := plugins.GetConfig(c.Name)

	if err != nil {
		panic(err)
	}

	GitRepo, err = opts.GetString("git_repo")

	if err != nil {
		panic(err)
	}

	GithubUsername, err = opts.GetString("github_username")

	if err != nil {
		panic(err)
	}

	VercelDeployHook, err = opts.GetString("vercel_deploy_hook")

	if err != nil {
		panic(err)
	}

	actions.RegisterActions(&actions.Action{
		Name:        "Update Production",
		Description: "Bump the Infinity-Next repository with the latest master branch.",
		Handler: func(*actions.ActionContext) (*actions.ActionResponse, error) {
			taskId := crypto.RandString(128)

			go updateProdTask(taskId)

			return &actions.ActionResponse{
				StatusCode: http.StatusNoContent,
				TaskID:     taskId,
			}, nil
		},
	})

	return nil
}
