package http

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func logEndpoint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the method and url from the request
		method := r.Method
		url := r.URL
		// Print the method and url to the console
		log.Info().Msg(fmt.Sprintf("HTTP Request >> %s %s", method, url))
		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
