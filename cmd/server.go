package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	deliveryhttp "github.com/ceperapl/requester/pkg/delivery/http"
	"github.com/ceperapl/requester/pkg/domain"
	httpstuff "github.com/ceperapl/requester/pkg/http"
	"github.com/ceperapl/requester/pkg/logging"
	"github.com/ceperapl/requester/pkg/mq"
	"github.com/ceperapl/requester/pkg/mq/rabbitmq"
	"github.com/ceperapl/requester/pkg/repository/mongodb"
	"github.com/ceperapl/requester/pkg/taskexec"
	"github.com/ceperapl/requester/pkg/usecase"
	"github.com/ceperapl/requester/pkg/validator"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	readHeaderTimeout       = 3 * time.Second
	gracefulShutdownTimeout = 20 * time.Second
)

var (
	errRabbitMQConn = errors.New("RabbitMQ connection closed")
)

// RunServer is a function that runs a HTTP server that handles tasks
// using MongoDB, RabbitMQ, and other packages.
func RunServer() error {
	// Init config
	config := newConfig()

	logging.LogInit(config.Debug)

	mongoDBClient, err := mongodb.NewClient(config.MongoDB.URI)
	if err != nil {
		return fmt.Errorf("couldn't connect to MongoDB: %w", err)
	}
	defer mongodb.Close(context.Background(), mongoDBClient)
	taskRepo := mongodb.NewTaskRepo(mongoDBClient, config.MongoDB.Database, config.MongoDB.Collection)

	rabbitMQConn, rabbitMQChannel, err := rabbitmq.NewConnection(config.RabbitMQ.URI)
	if err != nil {
		return fmt.Errorf("couldn't connect to RabbitMQ: %w", err)
	}
	//nolint: errcheck
	defer rabbitmq.CloseConnection(rabbitMQConn, rabbitMQChannel)
	workQueue, err := rabbitmq.New(rabbitMQConn, rabbitMQChannel, config.RabbitMQ.QueueName)
	if err != nil {
		return fmt.Errorf("couldn't create work queue: %w", err)
	}

	taskUseCase := usecase.NewTaskUseCase(taskRepo, workQueue, taskexec.NewTaskExecution())

	valid, err := validator.New(validator.WithJSONNamesForStructFields())
	if err != nil {
		return fmt.Errorf("couldn't create validator: %w", err)
	}

	rootMux := mux.NewRouter()
	if err := deliveryhttp.Handle(rootMux, taskUseCase, valid); err != nil {
		//nolint: wrapcheck
		return err
	}

	// Configure health checks
	healthchecker := httpstuff.NewHealthChecker()
	healthchecker.AddReadinessChecks(checkDB(mongoDBClient))
	healthchecker.AddReadinessChecks(checkMQ(rabbitMQConn))
	rootMux.Handle(config.HTTPServer.ReadinessEndpoint, healthchecker.ReadinessHandler())
	rootMux.Handle(config.HTTPServer.LivenessEndpoint, healthchecker.LivenessHandler())
	// Configure metrics
	rootMux.Handle("/metrics", promhttp.Handler())

	httpServerAddr := fmt.Sprintf("0.0.0.0:%d", config.HTTPServer.Port)
	httpSrv := &http.Server{
		Addr:              httpServerAddr,
		Handler:           rootMux,
		ReadHeaderTimeout: readHeaderTimeout, // to prevent DDoS attack called Slowloris attack
	}

	// channel for receiving errors from all goroutines
	doneC := make(chan error)

	// Start the HTTP server
	go func() {
		log.Info().Msg(fmt.Sprintf("run HTTP server on %s", httpServerAddr))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			doneC <- err
		}
	}()

	log.Info().Msg(fmt.Sprintf("start task processing; workers = %d", config.WorkersCount))
	runWorkers(config.WorkersCount, doneC, workQueue, taskUseCase)

	go gracefulShutdown(gracefulShutdownTimeout, httpSrv, rabbitMQConn, rabbitMQChannel, mongoDBClient)

	// waiting for the errors from http server and workers
	if err := <-doneC; err != nil {
		return err
	}

	return nil
}

func runWorkers(workersCount int, doneC chan<- error, mq mq.TaskQueuer, taskUseCase usecase.TaskUsecaser) {
	for i := 1; i <= workersCount; i++ {
		go func(workerNumber int) {
			err := mq.Consume(context.Background(), func(task domain.Task) {
				log.Debug().Msg(fmt.Sprintf("Worker #%d/%d is processing task with id: %q", workerNumber, workersCount, task.ID))
				if _, err := taskUseCase.ProcessTask(context.Background(), task); err != nil {
					log.Debug().Msg(fmt.Sprintf("Worker #%d/%d failed to process task with id: %q", workerNumber, workersCount, task.ID))

					return
				}
				log.Debug().Msg(fmt.Sprintf("Worker #%d/%d successfully processed task with id: %q", workerNumber, workersCount, task.ID))
			})
			if err != nil {
				doneC <- err
			}
		}(i)
	}
}

func gracefulShutdown(timeout time.Duration, httpServer *http.Server, rabbitMQConn *amqp.Connection,
	rabbitMQChannel *amqp.Channel, mongoDBClient *mongo.Client) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// shutdown http server
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Info().Msg(fmt.Sprintf("HTTP Server shutdown failed: %v\n", err))
		}
		log.Info().Msg("HTTP Server shutdown gracefully")

		// disconnect from RabbitMQ
		if err := rabbitmq.CloseConnection(rabbitMQConn, rabbitMQChannel); err != nil {
			log.Info().Msg(fmt.Sprintf("failed to disconnect from RabbitMQ: %v\n", err))
		}
		log.Info().Msg("successfully disconnected from RabbitMQ")

		// disconnect from MongoDB
		if err := mongodb.Close(ctx, mongoDBClient); err != nil {
			log.Info().Msg(fmt.Sprintf("failed to disconnect from MongoDB: %v\n", err))
		}
		log.Info().Msg("successfully disconnected from MongoDB")

		os.Exit(0)
	}()
}

func checkDB(client *mongo.Client) httpstuff.Check {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := client.Ping(ctx, nil); err != nil {
			//nolint: wrapcheck
			return err
		}

		return nil
	}
}

func checkMQ(conn *amqp.Connection) httpstuff.Check {
	return func() error {
		if conn.IsClosed() {
			return errRabbitMQConn
		}

		return nil
	}
}
