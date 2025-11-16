package config

import (
	"context"
	"time"

	confLoader "bitbucket.org/Amartha/go-config-loader-library"
)

type Configuration struct {
	App                AppConfiguration            `json:"app"`
	Log                LogConfiguration            `json:"log"`
	Database           DatabaseConfiguration       `json:"database"`
	Kafka              KafkaConfiguration          `json:"kafka"`
	GoAccounting       HTTPConfiguration           `json:"go_accounting"`
	FeatureFlagSDK     FeatureFlagSDKConfiguration `json:"feature_flag_sdk"`
	DDDNotification    DDDNotification             `json:"ddd_notification"`
	ConfigAccount      MapConfigAccount            `json:"config_account"`
	ExponentialBackoff ExponentialBackOffConfig    `json:"exponential_backoff"`
	AcuanLibConfig     AcuanLibConfig              `json:"go_acuan_lib"`
	NewRelicLicenseKey string                      `json:"new_relic_license_key"`
}

type AppConfiguration struct {
	Env             string        `json:"env"`
	Name            string        `json:"name"`
	GracefulTimeout time.Duration `json:"graceful_timeout"`
}

type LogConfiguration struct {
	Option string `json:"option"`
	Level  string `json:"level"`
}

type DatabaseConfiguration struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type KafkaConfiguration struct {
	Publishers KafkaPublishers `json:"publishers"`
	Consumers  KafkaConsumers  `json:"consumers"`
}

type KafkaPublishers struct {
	OrderTransaction       KafkaPublisherConfiguration
	JournalStream          KafkaPublisherConfiguration `json:"journal_stream"`
	TransformerStreamDLQ   KafkaPublisherConfiguration `json:"transformer_stream_dlq"`
	JournalStreamMigration KafkaPublisherConfiguration `json:"journal_stream_migration"`
}

type KafkaConsumers struct {
	HealthCheckPort                   string                     `json:"health_check_port"`
	PAPATransactionStream             KafkaConsumerConfiguration `json:"papa_transaction_stream"`
	LSMLoanLogs                       KafkaConsumerConfiguration `json:"lsm_loan_logs"`
	BREBillingRepaymentLogs           KafkaConsumerConfiguration `json:"bre_billing_repayment_logs"`
	PASAcuanTransactionNotif          KafkaConsumerConfiguration `json:"acuan_transaction_notif"`
	TransformerStreamDLQ              KafkaConsumerConfiguration `json:"transformer_stream_dlq"`
	PASAcuanTransactionNotifMigration KafkaConsumerConfiguration `json:"acuan_transaction_notif_migration"`
}

type KafkaConsumerConfiguration struct {
	Brokers       []string `json:"brokers"`
	Topic         string   `json:"topic"`
	ConsumerGroup string   `json:"consumer_group"`
}

type KafkaPublisherConfiguration struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
}

type FeatureFlagSDKConfiguration struct {
	URL             string        `json:"url"`
	Token           string        `json:"token"`
	Env             string        `json:"env"`
	RefreshInterval time.Duration `json:"refresh_interval"`
}

type HTTPConfiguration struct {
	BaseURL          string        `json:"base_url"`
	BaseURLMigration string        `json:"base_url_migration"`
	SecretKey        string        `json:"secret_key"`
	RetryCount       int           `json:"retry_count"`
	RetryWaitTime    int           `json:"retry_wait_time"`
	Timeout          time.Duration `json:"timeout"`
}

type DDDNotification struct {
	BaseURL       string `json:"base_url"`
	RetryCount    int    `json:"retry_count"`
	RetryWaitTime int    `json:"retry_wait_time"`

	SlackChannel string `json:"slack_channel"`
	TitleBot     string `json:"title_bot"`
}

type MapConfigAccount struct {
	MapAccountOrderTypesEntityProduct map[string]map[string]string `json:"map_account_order_types_entity_product"`
	MapAccountEntityTrxTypes          map[string]map[string]string `json:"map_account_entity_trx_types"`
	MapRepaymentCashInTransitEntity   map[string]string            `json:"map_repayment_cit_entity"`
	MapBankHORepaymentPooling         map[string]string            `json:"map_bank_ho_repayment_pooling_entity"`
	MapDSBPIAccountEntity             map[string]string            `json:"map_dsbpi_account_entity"`
}

type ExponentialBackOffConfig struct {
	MaxRetries        uint64        `json:"max_retries"`
	InitialInterval   time.Duration `json:"initial_interval"`
	MaxBackoffTime    time.Duration `json:"max_backoff_time"`
	MaxElapsedTime    time.Duration `json:"max_elapsed_time"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
}

type AcuanLibConfig struct {
	Kafka                 AcuanLibKafkaConfig `json:"kafka"`
	SourceSystem          string              `json:"source_system"`
	Topic                 string              `json:"topic"`
	TopicAccounting       string              `json:"topic_accounting"`
	TopUpKey              string              `json:"topup_key"`
	InvestmentKey         string              `json:"investment_key"`
	CashoutKey            string              `json:"cashout_key"`
	DisbursementKey       string              `json:"disbursement_key"`
	DisbursementFailedKey string              `json:"disbursement_failed_key"`
	RepaymentKey          string              `json:"repayment_key"`
	RefundKey             string              `json:"refund_key"`
}

type AcuanLibKafkaConfig struct {
	BrokerList        string `json:"broker_list"`
	PartitionStrategy string `json:"partition_strategy"`
}

func New(ctx context.Context) (*Configuration, error) {
	var cfg Configuration
	l := confLoader.New("", "", "",
		confLoader.WithConfigFileSearchPaths(
			"./config",
			"./../config",
			"./../../config"),
	)
	if err := l.Load(&cfg); err != nil {
		return &cfg, err
	}

	return &cfg, nil
}
