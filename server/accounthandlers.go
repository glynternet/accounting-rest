package server

import (
	"encoding/json"
	"net/http"

	"github.com/glynternet/go-accounting-storage"
	"github.com/pkg/errors"
)

func (s *server) accounts(w http.ResponseWriter, _ *http.Request) (int, error) {
	if w == nil {
		return http.StatusInternalServerError, errors.New("nil ResponseWriter")
	}
	store, err := s.NewStorage()
	if err != nil {
		return http.StatusServiceUnavailable, errors.Wrap(err, "creating new storage")
	}
	var as *storage.Accounts
	as, err = store.SelectAccounts()
	if err != nil {
		return http.StatusServiceUnavailable, errors.Wrap(err, "selecting accounts from client")
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set(`Content-Type`, `application/json; charset=UTF-8`)
	return http.StatusOK, errors.Wrap(
		json.NewEncoder(w).Encode(as),
		"error encoding accounts json",
	)
}

//func (s *server) muxAccountIDHandlerfunc(w http.ResponseWriter, r *http.Request) (int, error) {
//	vars := mux.Vars(r)
//	if vars == nil {
//		return http.StatusBadRequest, errors.New("no context variables")
//		return http.StatusBadRequest, fmt.Errorf("no account_id context variable")
//}
//
//key := "account_id"
//idString := vars[key]
//id, err := strconv.ParseUint(idString, 10, 64)
//if err != nil {
//	return http.StatusBadRequest, errors.Wrapf(err, "parsing %s to uint", key)
//}
//return s.accountHandlerWithID(id)(w, r)
//}

func (s *server) accountHandlerWithID(id uint64) appHandler {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		if w == nil {
			return http.StatusInternalServerError, errors.New("nil ResponseWriter")
		}
		_, err := s.NewStorage()
		if err != nil {
			return http.StatusServiceUnavailable, errors.Wrap(err, "creating new storage")
		}
		return 0, errors.New("not implemented")
	}
}
