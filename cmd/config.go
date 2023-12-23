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

	mongodbURIEnv     = "MONGODB_URI"
	mongodbURIDefault = "mongodb://admin:password@localhost:27017"

	mongodbDatabaseEnv     = "MONGODB_DB"
	mongodbDatabaseDefault = "test"

	mongodbCollectionEnv     = "MONGODB_COLLECTION"
	mongodbCollectionDefault = "tasks"

	rabbitmqURIEnv     = "RABBITMQ_URI"
	rabbitmqURIDefault = "amqp://guest:guest@localhost:5672/"

	rabbitmqQueueEnv     = "RABBITMQ_QUEUE"
	rabbitmqQueueDefault = "task_queue"

	taskReqTimeoutEnv     = "TASK_REQ_TIMEOUT"
	taskReqTimeoutDefault = 10

	workersCountEnv     = "WORKERS_COUNT"
	workersCountDefault = 10

	debugEnv     = "DEBUG"
	debugDefault = false
)

type config struct {
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
	WorkersCount   int
	TaskReqTimeout int // in seconds

	Debug bool
}

//nolint: funlen
func newConfig() config {
	conf := config{}

	// Init config via env variables
	viper.SetEnvPrefix(envPrefix)

	//nolint: errcheck
	viper.BindEnv(httpServerPortEnv)
	viper.SetDefault(httpServerPortEnv, httpServerPortDefault)
	conf.HTTPServer.Port = viper.GetUint16(httpServerPortEnv)

	//nolint: errcheck
	viper.BindEnv(httpReadinessEndpointEnv)
	viper.SetDefault(httpReadinessEndpointEnv, httpReadinessEndpointDefault)
	conf.HTTPServer.ReadinessEndpoint = viper.GetString(httpReadinessEndpointEnv)

	//nolint: errcheck
	viper.BindEnv(httpLivenessEndpointEnv)
	viper.SetDefault(httpLivenessEndpointEnv, httpLivenessEndpointDefault)
	conf.HTTPServer.LivenessEndpoint = viper.GetString(httpLivenessEndpointEnv)

	//nolint: errcheck
	viper.BindEnv(mongodbURIEnv)
	viper.SetDefault(mongodbURIEnv, mongodbURIDefault)
	conf.MongoDB.URI = viper.GetString(mongodbURIEnv)

	//nolint: errcheck
	viper.BindEnv(mongodbDatabaseEnv)
	viper.SetDefault(mongodbDatabaseEnv, mongodbDatabaseDefault)
	conf.MongoDB.Database = viper.GetString(mongodbDatabaseEnv)

	//nolint: errcheck
	viper.BindEnv(mongodbCollectionEnv)
	viper.SetDefault(mongodbCollectionEnv, mongodbCollectionDefault)
	conf.MongoDB.Collection = viper.GetString(mongodbCollectionEnv)

	//nolint: errcheck
	viper.BindEnv(rabbitmqURIEnv)
	viper.SetDefault(rabbitmqURIEnv, rabbitmqURIDefault)
	conf.RabbitMQ.URI = viper.GetString(rabbitmqURIEnv)

	//nolint: errcheck
	viper.BindEnv(rabbitmqQueueEnv)
	viper.SetDefault(rabbitmqQueueEnv, rabbitmqQueueDefault)
	conf.RabbitMQ.QueueName = viper.GetString(rabbitmqQueueEnv)

	//nolint: errcheck
	viper.BindEnv(taskReqTimeoutEnv)
	viper.SetDefault(taskReqTimeoutEnv, taskReqTimeoutDefault)
	conf.TaskReqTimeout = viper.GetInt(taskReqTimeoutEnv)

	//nolint: errcheck
	viper.BindEnv(workersCountEnv)
	viper.SetDefault(workersCountEnv, workersCountDefault)
	conf.WorkersCount = viper.GetInt(workersCountEnv)

	//nolint: errcheck
	viper.BindEnv(debugEnv)
	viper.SetDefault(debugEnv, debugDefault)
	conf.Debug = viper.GetBool(debugEnv)

	// Init config via flags
	pflag.Uint16Var(&conf.HTTPServer.Port, "httpserver.port", conf.HTTPServer.Port,
		fmt.Sprintf("HTTP Server port; env: %s", fmt.Sprintf("%s_%s", envPrefix, httpServerPortEnv)))
	pflag.StringVar(&conf.HTTPServer.ReadinessEndpoint, "httpserver.readiness", conf.HTTPServer.ReadinessEndpoint,
		fmt.Sprintf("HTTP Server readiness endpoint name; env: %s", fmt.Sprintf("%s_%s", envPrefix, httpReadinessEndpointEnv)))
	pflag.StringVar(&conf.HTTPServer.LivenessEndpoint, "httpserver.liveness", conf.HTTPServer.LivenessEndpoint,
		fmt.Sprintf("HTTP Server liveness endpoint name; env: %s", fmt.Sprintf("%s_%s", envPrefix, httpLivenessEndpointEnv)))
	pflag.StringVar(&conf.MongoDB.Database, "mongodb.db", conf.MongoDB.Database,
		fmt.Sprintf("MongoDB database; env: %s", fmt.Sprintf("%s_%s", envPrefix, mongodbDatabaseEnv)))
	pflag.StringVar(&conf.MongoDB.Collection, "mongodb.collection", conf.MongoDB.Collection,
		fmt.Sprintf("MongoDB collection; env: %s", fmt.Sprintf("%s_%s", envPrefix, mongodbCollectionEnv)))
	pflag.StringVar(&conf.RabbitMQ.QueueName, "rabbitmq.queue", conf.RabbitMQ.QueueName,
		fmt.Sprintf("RabbitMQ queue name; env: %s", fmt.Sprintf("%s_%s", envPrefix, rabbitmqQueueEnv)))
	pflag.IntVar(&conf.TaskReqTimeout, "task.req.timeout", conf.TaskReqTimeout,
		fmt.Sprintf("Timeout for task requests in seconds; env: %s", fmt.Sprintf("%s_%s", envPrefix, taskReqTimeoutEnv)))
	pflag.IntVar(&conf.WorkersCount, "workers", conf.WorkersCount,
		fmt.Sprintf("number of workers for simultaneous processing of tasks; env: %s", fmt.Sprintf("%s_%s", envPrefix, workersCountEnv)))
	pflag.BoolVar(&conf.Debug, "debug", conf.Debug,
		fmt.Sprintf("sets log level to debug; env: %s", fmt.Sprintf("%s_%s", envPrefix, debugEnv)))

	pflag.Parse()

	return conf
}
