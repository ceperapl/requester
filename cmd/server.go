package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	deliveryhttp "github.com/ceperapl/requester/pkg/delivery/http"
	"github.com/ceperapl/requester/pkg/domain"
	"github.com/ceperapl/requester/pkg/logging"
	"github.com/ceperapl/requester/pkg/mq/rabbitmq"
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
	defer taskRepo.Close()

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

	// Start the HTTP server
	httpServerAddr := fmt.Sprintf("0.0.0.0:%d", config.HTTPServer.Port)
	go func() {
		log.Info().Msg(fmt.Sprintf("run HTTP server on %s", httpServerAddr))
		doneC <- http.ListenAndServe(httpServerAddr, httpHandler)
	}()

	log.Info().Msg(fmt.Sprintf("start task processing; workers = %d", config.WorkersCount))
	for i := 0; i < config.WorkersCount; i++ {
		go func() {
			doneC <- mq.Consume(func(msg string) error {
				var task domain.Task
				if err := json.Unmarshal([]byte(msg), &task); err != nil {
					return err
				}
				if err := taskService.ProcessTask(&task); err != nil {
					return err
				}
				return nil
			})
		}()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigChan

		log.Info().Msg("graceful shutdown")
		os.Exit(0)
	}()

	// waiting for the errors from servers
	if err := <-doneC; err != nil {
		return err
	}

	return nil
}
