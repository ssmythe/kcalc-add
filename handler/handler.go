package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ssmythe/kcalc-add/service"
)

type addRequest struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

type addResponse struct {
	Result float64 `json:"result"`
}

func AddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var req addRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	res := addResponse{Result: service.Add(req.A, req.B)}
	// optional: log inputs+result
	log.Printf("add: a=%v b=%v result=%v", req.A, req.B, res.Result)
	writeJSON(w, http.StatusOK, res)

}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	type errBody struct {
		Error string `json:"error"`
	}
	writeJSON(w, code, errBody{Error: msg})
}
