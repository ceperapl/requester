package http

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

// logEndpoint is a middleware function that logs the method and url of the endpoint before calling the next handler
func logEndpoint(next http.Handler) http.Handler {
	// Return a new handler function
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
