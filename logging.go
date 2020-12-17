/*
 * Shea Odland, Von Castro, Eric Wedemire
 * CMPT315
 * Group Project: Codenames
 */

package main

import (
	"log"
	"net/http"
)

//loggingMiddleware, it simply will log any time an API call is made and what
// method was used to call it
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		// add item to log
		log.Println(request.RequestURI, request.Method)

		// Call the next handler or middleware
		// in this case, it would be statsGetHandle or statsDeleteHandle
		next.ServeHTTP(writer, request)
	})
}
