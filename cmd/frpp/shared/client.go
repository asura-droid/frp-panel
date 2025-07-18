package shared

import (
	"context"

	bizclient "github.com/VaalaCat/frp-panel/biz/client"
	"github.com/VaalaCat/frp-panel/conf"
	"github.com/VaalaCat/frp-panel/defs"
	"github.com/VaalaCat/frp-panel/pb"
	"github.com/VaalaCat/frp-panel/services/app"
	"github.com/VaalaCat/frp-panel/services/clientrpc"
	"github.com/VaalaCat/frp-panel/services/rpc"
	"github.com/VaalaCat/frp-panel/services/tunnel"
	"github.com/VaalaCat/frp-panel/services/watcher"
	"github.com/VaalaCat/frp-panel/utils/logger"
	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
)

type runClientParam struct {
	fx.In

	Lc fx.Lifecycle

	Ctx            *app.Context
	AppInstance    app.Application
	TaskManager    watcher.Client `name:"clientTaskManager"`
	WorkersManager app.WorkersManager

	Cfg conf.Config
}

func runClient(param runClientParam) {
	var (
		ctx          = param.Ctx
		clientID     = param.AppInstance.GetConfig().Client.ID
		clientSecret = param.AppInstance.GetConfig().Client.Secret
		appInstance  = param.AppInstance
	)
	logger.Logger(ctx).Infof("start to run client")
	if len(clientSecret) == 0 {
		logger.Logger(ctx).Fatal("client secret cannot be empty")
	}

	if len(clientID) == 0 {
		logger.Logger(ctx).Fatal("client id cannot be empty")
	}

	param.TaskManager.AddDurationTask(defs.PullConfigDuration,
		bizclient.PullConfig, appInstance, clientID, clientSecret)
	param.TaskManager.AddDurationTask(defs.PullClientWorkersDuration,
		bizclient.PullWorkers, appInstance, clientID, clientSecret)

	var wg conc.WaitGroup
	param.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			appInstance.SetRPCCred(NewClientCred(appInstance))
			appInstance.SetMasterCli(rpc.NewMasterCli(appInstance))
			appInstance.SetClientController(tunnel.NewClientController())

			cliRpcHandler := clientrpc.NewClientRPCHandler(
				appInstance,
				clientID,
				clientSecret,
				pb.Event_EVENT_REGISTER_CLIENT,
				bizclient.HandleServerMessage,
			)
			appInstance.SetClientRPCHandler(cliRpcHandler)

			// --- init once start ---
			initClientOnce(appInstance, clientID, clientSecret)
			initClientWorkerOnce(appInstance, clientID, clientSecret)
			// --- init once stop ----

			wg.Go(cliRpcHandler.Run)
			wg.Go(param.TaskManager.Run)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			param.TaskManager.Stop()
			appInstance.GetClientRPCHandler().Stop()

			wg.Wait()
			return nil
		},
	})
}

func initClientOnce(appInstance app.Application, clientID, clientSecret string) {
	err := bizclient.PullConfig(appInstance, clientID, clientSecret)
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Errorf("cannot pull client config, wait for retry")
	}
}

func initClientWorkerOnce(appInstance app.Application, clientID, clientSecret string) {
	err := bizclient.PullWorkers(appInstance, clientID, clientSecret)
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Errorf("cannot pull client workers, wait for retry")
	}
}
