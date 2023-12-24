package main

import (
	"log"

	"github.com/ceperapl/requester/cmd"
)

//go:generate mockery --dir "./pkg/repository" --filename taskRepository.go --output "./pkg/repository/mocks" --outpkg "mocks" --name TaskRepository
//go:generate mockery --dir "./pkg/mq" --filename taskQueue.go --output "./pkg/mq/mocks" --outpkg "mocks" --name TaskQueuer
//go:generate mockery --dir "./pkg/taskexec" --filename taskExecutor.go --output "./pkg/taskexec/mocks" --outpkg "mocks" --name TaskExecutor
//go:generate mockery --dir "./pkg/usecase" --filename taskService.go --output "./pkg/usecase/mocks" --outpkg "mocks" --name TaskUsecaser

func main() {
	if err := cmd.RunServer(); err != nil {
		log.Fatal(err)
	}
}
