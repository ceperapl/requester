package cmd

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	envPrefix = "REQUESTER"

	httpServerPortEnv     = "HTTP_SERVER_PORT"
	httpServerPortDefault = 8080

	httpReadinessEndpointEnv     = "HTTP_READINESS_ENDPOINT"
	httpReadinessEndpointDefault = "/ready"

	httpLivenessEndpointEnv     = "HTTP_LIVENESS_ENDPOINT"
	httpLivenessEndpointDefault = "/health"

	mongodbUriEnv     = "MONGODB_URI"
	mongodbUriDefault = "mongodb://admin:password@localhost:27017"

	mongodbDatabaseEnv     = "MONGODB_DB"
	mongodbDatabaseDefault = "test"

	mongodbCollectionEnv     = "MONGODB_COLLECTION"
	mongodbCollectionDefault = "tasks"

	rabbitmqUriEnv     = "RABBITMQ_URI"
	rabbitmqUriDefault = "amqp://guest:guest@localhost:5672/"

	rabbitmqQueueNameEnv     = "RABBITMQ_QUEUE_NAME"
	rabbitmqQueueNameDefault = "task_queue"

	workersCountEnv     = "WORKERS_COUNT"
	workersCountDefault = 10

	debugEnv     = "DEBUG"
	debugDefault = false
)

type Config struct {
	HTTPServer struct {
		Port              uint16
		ReadinessEndpoint string
		LivenessEndpoint  string
	}
	MongoDB struct {
		URI        string
		Database   string
		Collection string
	}
	RabbitMQ struct {
		URI       string
		QueueName string
	}
	WorkersCount int

	Debug bool
}

func NewConfig() Config {
	config := Config{}

	// Init config via env variables
	viper.SetEnvPrefix(envPrefix)

	viper.BindEnv(httpServerPortEnv)
	viper.SetDefault(httpServerPortEnv, httpServerPortDefault)
	config.HTTPServer.Port = viper.GetUint16(httpServerPortEnv)

	viper.BindEnv(httpReadinessEndpointEnv)
	viper.SetDefault(httpReadinessEndpointEnv, httpReadinessEndpointDefault)
	config.HTTPServer.ReadinessEndpoint = viper.GetString(httpReadinessEndpointEnv)

	viper.BindEnv(httpLivenessEndpointEnv)
	viper.SetDefault(httpLivenessEndpointEnv, httpLivenessEndpointDefault)
	config.HTTPServer.LivenessEndpoint = viper.GetString(httpLivenessEndpointEnv)

	viper.BindEnv(mongodbUriEnv)
	viper.SetDefault(mongodbUriEnv, mongodbUriDefault)
	config.MongoDB.URI = viper.GetString(mongodbUriEnv)

	viper.BindEnv(mongodbDatabaseEnv)
	viper.SetDefault(mongodbDatabaseEnv, mongodbDatabaseDefault)
	config.MongoDB.Database = viper.GetString(mongodbDatabaseEnv)

	viper.BindEnv(mongodbCollectionEnv)
	viper.SetDefault(mongodbCollectionEnv, mongodbCollectionDefault)
	config.MongoDB.Collection = viper.GetString(mongodbCollectionEnv)

	viper.BindEnv(rabbitmqUriEnv)
	viper.SetDefault(rabbitmqUriEnv, rabbitmqUriDefault)
	config.RabbitMQ.URI = viper.GetString(rabbitmqUriEnv)

	viper.BindEnv(rabbitmqQueueNameEnv)
	viper.SetDefault(rabbitmqQueueNameEnv, rabbitmqQueueNameDefault)
	config.RabbitMQ.QueueName = viper.GetString(rabbitmqQueueNameEnv)

	viper.BindEnv(workersCountEnv)
	viper.SetDefault(workersCountEnv, workersCountDefault)
	config.WorkersCount = viper.GetInt(workersCountEnv)

	viper.BindEnv(debugEnv)
	viper.SetDefault(debugEnv, debugDefault)
	config.Debug = viper.GetBool(debugEnv)

	// Init config via flags
	pflag.Uint16Var(&config.HTTPServer.Port, "httpserver.port", config.HTTPServer.Port,
		fmt.Sprintf("HTTP Server port; env: %s", httpServerPortEnv))
	pflag.StringVar(&config.HTTPServer.ReadinessEndpoint, "httpserver.readiness", config.HTTPServer.ReadinessEndpoint,
		fmt.Sprintf("HTTP Server readiness endpoint name; env: %s", httpReadinessEndpointEnv))
	pflag.StringVar(&config.HTTPServer.LivenessEndpoint, "httpserver.liveness", config.HTTPServer.LivenessEndpoint,
		fmt.Sprintf("HTTP Server liveness endpoint name; env: %s", httpLivenessEndpointEnv))
	pflag.StringVar(&config.MongoDB.URI, "mongodb.uri", config.MongoDB.URI,
		fmt.Sprintf("MongoDB URI; env: %s", mongodbUriEnv))
	pflag.StringVar(&config.MongoDB.Database, "mongodb.db", config.MongoDB.Database,
		fmt.Sprintf("MongoDB database; env: %s", mongodbDatabaseEnv))
	pflag.StringVar(&config.MongoDB.Collection, "mongodb.collection", config.MongoDB.Collection,
		fmt.Sprintf("MongoDB collection; env: %s", mongodbCollectionEnv))
	pflag.StringVar(&config.RabbitMQ.URI, "rabbitmq.uri", config.RabbitMQ.URI,
		fmt.Sprintf("RabbitMQ URI; env: %s", rabbitmqUriEnv))
	pflag.StringVar(&config.RabbitMQ.QueueName, "rabbitmq.queue", config.RabbitMQ.QueueName,
		fmt.Sprintf("RabbitMQ queue name; env: %s", rabbitmqQueueNameEnv))
	pflag.IntVar(&config.WorkersCount, "workers", config.WorkersCount,
		fmt.Sprintf("number of workers for simultaneous processing of tasks; env: %s", workersCountEnv))
	pflag.BoolVar(&config.Debug, "debug", config.Debug,
		fmt.Sprintf("sets log level to debug; env: %s", debugEnv))

	pflag.Parse()

	return config
}
