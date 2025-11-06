package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"bitbucket.org/Amartha/go-x/environment"
	xlog "bitbucket.org/Amartha/go-x/log"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/http"
	"bitbucket.org/Amartha/go-megatron/internal/kafka/consumer"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/flag"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/metrics"

	"github.com/newrelic/go-agent/v3/integrations/nrzap"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "consumer",
	Short: "Consumer is a kafka consumer",
	Long:  ``,
}

var (
	runJobCmd = &cobra.Command{
		Use:     "run",
		Short:   "Run consumer",
		Long:    fmt.Sprintf("Run consumer for handling messages, available consumer type: %s", strings.Join(consumer.ListConsumerName, ", ")),
		Example: "consumer run -n={consumer-type-name}",
		Run:     runConsumer,
	}
	runConsumerCmdName = "name"
)

func init() {
	rootCmd.AddCommand(runJobCmd)

	runJobCmd.Flags().StringP(runConsumerCmdName, "n", "", "job name")
	runJobCmd.MarkFlagRequired(runConsumerCmdName)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func runConsumer(ccmd *cobra.Command, args []string) {
	name, _ := ccmd.Flags().GetString(runConsumerCmdName)
	var (
		ctx      = context.Background()
		starters []graceful.ProcessStarter
		stoppers []graceful.ProcessStopper
	)
	ctx, cancel := context.WithCancel(ctx)

	cfg, err := config.New(ctx)
	if err != nil {
		log.Fatalf("failed initializing service configuration: %v", err)
	}

	xlog.Init(cfg.App.Name,
		xlog.WithLogToOption(cfg.Log.Option),
		xlog.WithLogEnvOption(cfg.App.Env),
		xlog.WithCaller(true),
		xlog.AddCallerSkip(2),
		xlog.DebugLogLevel())

	nr := setupNR(ctx, cfg)
	mtc := metrics.New()

	flagClient, err := flag.New(cfg)
	if err != nil {
		log.Fatalf("failed initializing flag client: %v", err)
	}

	stoppers = append(stoppers, func(_ context.Context) error {
		flagClient.Close()
		return nil
	})

	if os.Getenv("USE_DB_MIGRATION") == "true" {
		// This replaces need for consume transaction for new DB migration
		cfg.Kafka.Consumers.PASAcuanTransactionNotif = cfg.Kafka.Consumers.PASAcuanTransactionNotifMigration
		cfg.Kafka.Publishers.JournalStream = cfg.Kafka.Publishers.JournalStreamMigration
		cfg.GoAccounting.BaseURL = cfg.GoAccounting.BaseURLMigration
	}

	consumer, stopperConsumer, err := consumer.New(ctx, name, cfg, nr, mtc, flagClient)
	if err != nil {
		xlog.Fatal(ctx, "error initializing kafka consumer", xlog.Err(err))
	}

	starterConsumer := consumer.Start()
	starters = append(starters, starterConsumer)
	stoppers = append(stoppers, stopperConsumer...)

	http := http.NewHealthCheck(cfg)
	starterApi, stopperApi := http.Start()
	starters = append(starters, starterApi)
	stoppers = append(stoppers, stopperApi)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		graceful.StartProcessAtBackground(starters...)
		graceful.StopProcessAtBackground(cfg.App.GracefulTimeout, stoppers...)
		wg.Done()
	}()

	wg.Wait()
	cancel()
	xlog.Info(ctx, "consumer server stopped!")
}

func setupNR(ctx context.Context, cfg *config.Configuration) *newrelic.Application {
	if env := environment.ToEnvironment(cfg.App.Env); env == environment.PROD_ENV {
		logger, ok := xlog.Loggers.Load(xlog.DefaultLogger)
		if !ok {
			return nil
		}
		app, err := newrelic.NewApplication(
			newrelic.ConfigAppName(cfg.App.Name),
			newrelic.ConfigLicense(cfg.NewRelicLicenseKey),
			func(config *newrelic.Config) {
				config.Logger = nrzap.Transform(logger)
			},
			newrelic.ConfigDistributedTracerEnabled(true),
		)
		if err != nil {
			xlog.Errorf(ctx, "setupNR.NewApplication - %v", err)
		}
		if err = app.WaitForConnection(15 * time.Second); nil != err {
			xlog.Errorf(ctx, "setupNR.WaitForConnection - %v", err)
		}
		return app
	}
	return nil
}
