package main

import (
	"errors"
	"movie_api/internal/data"
	"movie_api/internal/validator"
	"net/http"
)

func (app *application) addUserPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var input struct {
		Permissions data.Permissions `json:"permissions"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	permissions := input.Permissions
	v := validator.New()
	data.ValidatePermissions(v, permissions)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Permissions.AddForUser(id, permissions...)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNonExistUser):
			v.AddError("userID", err.Error())
			app.failedValidationResponse(w, r, v.Errors)
			//		case errors.Is(err, data.ErrDuplicatePermission):
			//			v.AddError("permission", err.Error())
			//			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "permissions successfully added"}, nil)
}

func (app *application) deleteUserPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	var input struct {
		Permissions data.Permissions `json:"permissions"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	permissions := input.Permissions
	v := validator.New()
	data.ValidatePermissions(v, permissions)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Permissions.DeleteForUser(id, permissions...)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNonExistPermission):
			v.AddError("permission", err.Error())
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "permissions successfully deleted"}, nil)
}

func (app *application) getUserPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	permissions, err := app.models.Permissions.GetAllForUser(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"permissions": permissions}, nil)
}
