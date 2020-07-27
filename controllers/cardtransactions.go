package controllers

import (
	"encoding/json"
	"github.com/donohutcheon/gowebserver/models/auth"
	"net/http"

	"github.com/donohutcheon/gowebserver/controllers/errors"
	"github.com/donohutcheon/gowebserver/controllers/response"
	"github.com/donohutcheon/gowebserver/datalayer"
	"github.com/donohutcheon/gowebserver/models"
	"github.com/donohutcheon/gowebserver/models/pagination"
	"github.com/donohutcheon/gowebserver/state"
)

func CreateCardTransaction(w http.ResponseWriter, r *http.Request, state *state.ServerState) error {
	if r.Method == http.MethodOptions {
		return nil
	}

	userID := r.Context().Value(auth.UserKey).(int64) //Grab the id of the userID that send the request
	cardTransaction := models.NewCardTransaction(state)

	err := json.NewDecoder(r.Body).Decode(cardTransaction)
	if err != nil {
		resp := response.New(false, "Error while decoding request body")
		errResp := resp.Respond(w)
		if errResp != nil {
			return errResp
		}

		errors.WriteError(w, err, http.StatusBadRequest)
		return err
	}

	cardTransaction.UserID = userID
	data, err := cardTransaction.CreateCardTransaction()
	if err != nil {
		errors.WriteError(w, err)
		return err
	}

	resp := response.New(true, "success")
	resp.Set("cardTransaction", data)

	return resp.Respond(w)
}

func GetCardTransactions(w http.ResponseWriter, r *http.Request, state *state.ServerState) error {
	if r.Method == http.MethodOptions {
		return nil
	}

	cardTransaction := models.NewCardTransaction(state)
	err := pagination.ParsePagination(state.Logger, r.URL.Query(), cardTransaction)
	if err != nil {
		errors.WriteError(w, err, http.StatusBadRequest)
		return err
	}
	err = cardTransaction.SetFilterCriteria(r.URL.Query())
	if err != nil {
		errors.WriteError(w, err, http.StatusBadRequest)
		return err
	}

	userID := r.Context().Value(auth.UserKey).(int64)
	data, err := cardTransaction.GetCardTransactionsByUserID(userID)
	if err != nil && err != datalayer.ErrNoData {
		errors.WriteError(w, err, http.StatusInternalServerError)
		return err
	}

	resp := response.New(true, "success")
	resp.Set("cardTransactions", data)

	return resp.Respond(w)
}
