package cmd

import (
	"context"
	"encoding/json"
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
	"github.com/ceperapl/requester/pkg/repository"
	"github.com/ceperapl/requester/pkg/repository/mongodb"
	"github.com/ceperapl/requester/pkg/usecase"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func RunServer() error {
	doneC := make(chan error)

	// Init config
	config := NewConfig()

	logging.LogInit(config.Debug)

	log.Info().Msg("connecting to MongoDB...")

	taskRepo, err := mongodb.NewTaskRepo(config.MongoDB.URI, config.MongoDB.Database, config.MongoDB.Collection)
	if err != nil {
		return err
	}
	defer taskRepo.Close(context.Background())

	log.Info().Msg("connecting to RabbitMQ...")

	mq, err := rabbitmq.New(config.RabbitMQ.URI, config.RabbitMQ.QueueName)
	if err != nil {
		return err
	}
	defer mq.Close()

	taskService, err := usecase.NewTaskService(taskRepo, mq)
	if err != nil {
		return err
	}

	rootMux := mux.NewRouter()

	httpHandler, err := deliveryhttp.NewTaskHandler(rootMux, taskService)
	if err != nil {
		return err
	}

	// Create a new HTTP server
	httpServerAddr := fmt.Sprintf("0.0.0.0:%d", config.HTTPServer.Port)
	httpSrv := &http.Server{
		Addr:    httpServerAddr,
		Handler: httpHandler,
	}

	// Start the HTTP server
	go func() {
		log.Info().Msg(fmt.Sprintf("run HTTP server on %s", httpServerAddr))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			doneC <- err
		}
	}()

	log.Info().Msg(fmt.Sprintf("start task processing; workers = %d", config.WorkersCount))
	runWorkers(config.WorkersCount, doneC, mq, taskService)

	go gracefulShutdown(httpSrv, mq, taskRepo)

	// waiting for the errors from servers
	if err := <-doneC; err != nil {
		return err
	}

	return nil
}

func runWorkers(workersCount int, doneC chan<- error, mq mq.WorkQueue, taskService usecase.TaskService) {
	for i := 1; i <= workersCount; i++ {
		go func(workerNumber int) {
			err := mq.Consume(func(msg string) {
				log.Debug().Msg(fmt.Sprintf("Worker #%d/%d is processing task: %s", workerNumber, workersCount, msg))
				var task domain.Task
				json.Unmarshal([]byte(msg), &task)
				if err := taskService.ProcessTask(&task); err != nil {
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

func gracefulShutdown(httpServer *http.Server, mq mq.WorkQueue, taskRepo repository.TaskRepository) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("graceful shutdown")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// shutdown http server
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Info().Msg(fmt.Sprintf("HTTP Server shutdown failed: %v\n", err))
		}
		log.Info().Msg("HTTP Server shutdown gracefully")

		// disconnect from RabbitMQ
		if err := mq.Close(); err != nil {
			log.Info().Msg(fmt.Sprintf("failed to disconnect from RabbitMQ: %v\n", err))
		}
		log.Info().Msg("successfully disconnected from RabbitMQ")

		// disconnect from MongoDB
		if err := taskRepo.Close(ctx); err != nil {
			log.Info().Msg(fmt.Sprintf("failed to disconnect from MongoDB: %v\n", err))
		}
		log.Info().Msg("successfully disconnected from MongoDB")

		os.Exit(0)
	}()
}
