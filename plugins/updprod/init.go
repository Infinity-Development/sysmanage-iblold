package updprod

import (
	"net/http"

	"github.com/infinitybotlist/eureka/crypto"
	"github.com/infinitybotlist/sysmanage-web/plugins/actions"
	"github.com/infinitybotlist/sysmanage-web/types"
)

func InitPlugin(c *types.PluginConfig) error {
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
