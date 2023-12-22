package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	deliveryhttp "github.com/ceperapl/requester/pkg/delivery/http"
	"github.com/ceperapl/requester/pkg/domain"
	"github.com/ceperapl/requester/pkg/logging"
	"github.com/ceperapl/requester/pkg/mq"
	"github.com/ceperapl/requester/pkg/mq/rabbitmq"
	"github.com/ceperapl/requester/pkg/repository/mongodb"
	"github.com/ceperapl/requester/pkg/usecase"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	readHeaderTimeout       = 3 * time.Second
	gracefulShutdownTimeout = 20 * time.Second
)

func RunServer() error {
	// Init config
	config := NewConfig()

	logging.LogInit(config.Debug)

	log.Info().Msg("connecting to MongoDB...")
	mongoDBClient, err := mongodb.NewClient(config.MongoDB.URI)
	if err != nil {
		return fmt.Errorf("couldn't connect to MongoDB: %w", err)
	}
	defer mongodb.Close(context.Background(), mongoDBClient)
	taskRepo := mongodb.NewTaskRepo(mongoDBClient, config.MongoDB.Database, config.MongoDB.Collection)

	log.Info().Msg("connecting to RabbitMQ...")
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

	taskService, err := usecase.NewTaskService(taskRepo, workQueue)
	if err != nil {
		return fmt.Errorf("couldn't create task usecase: %w", err)
	}

	rootMux := mux.NewRouter()
	httpHandler, err := deliveryhttp.NewTaskHandler(rootMux, taskService)
	if err != nil {
		return fmt.Errorf("coudn't create http task handler: %w", err)
	}

	httpServerAddr := fmt.Sprintf("0.0.0.0:%d", config.HTTPServer.Port)
	httpSrv := &http.Server{
		Addr:              httpServerAddr,
		Handler:           httpHandler,
		ReadHeaderTimeout: readHeaderTimeout,
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
	runWorkers(config.WorkersCount, doneC, workQueue, taskService)

	go gracefulShutdown(gracefulShutdownTimeout, httpSrv, rabbitMQConn, rabbitMQChannel, mongoDBClient)

	// waiting for the errors from http server and workers
	if err := <-doneC; err != nil {
		return err
	}

	return nil
}

func runWorkers(workersCount int, doneC chan<- error, mq mq.WorkQueuer, taskService usecase.TaskUsecaser) {
	for i := 1; i <= workersCount; i++ {
		go func(workerNumber int) {
			err := mq.Consume(context.Background(), func(msg string) {
				log.Debug().Msg(fmt.Sprintf("Worker #%d/%d is processing task: %s", workerNumber, workersCount, msg))
				var task domain.Task
				//nolint: errcheck
				json.Unmarshal([]byte(msg), &task)
				if err := taskService.ProcessTask(context.Background(), &task); err != nil {
					log.Debug().Msg(fmt.Sprintf("Worker #%d/%d failed to process task: %s", workerNumber, workersCount, msg))

					return
				}
				log.Debug().Msg(fmt.Sprintf("Worker #%d/%d successfully processed task: %s", workerNumber, workersCount, msg))
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
		log.Info().Msg("graceful shutdown")

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
