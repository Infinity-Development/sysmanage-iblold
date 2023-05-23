package updprod

import "github.com/infinitybotlist/sysmanage-web/core/logger"

func updateProdTask(taskId string) {
	logger.LogMap.Add(taskId, "Going to update production...", true)
}
