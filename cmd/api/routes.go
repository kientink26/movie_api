package main

import (
	"github.com/julienschmidt/httprouter"
	"movie_api/internal/data"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission(data.MoviesRead, app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission(data.MoviesWrite, app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission(data.MoviesRead, app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission(data.MoviesWrite, app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission(data.MoviesWrite, app.deleteMovieHandler))

	router.HandlerFunc(http.MethodGet, "/v1/movies/:id/comments", app.requirePermission(data.CommentsRead, app.listCommentsHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies/:id/comments", app.requirePermission(data.CommentsWrite, app.createCommentHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.HandlerFunc(http.MethodGet, "/v1/users", app.requirePermission(data.UsersRead, app.listUsersHandler))

	router.HandlerFunc(http.MethodPut, "/v1/users/:id/permissions", app.requirePermission(data.PermissionsWrite, app.addUserPermissionsHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/users/:id/permissions", app.requirePermission(data.PermissionsWrite, app.deleteUserPermissionsHandler))
	router.HandlerFunc(http.MethodGet, "/v1/users/:id/permissions", app.requirePermission(data.PermissionsRead, app.getUserPermissionsHandler))

	return app.recoverPanic(app.enableCORS(app.authenticate(router)))
}
