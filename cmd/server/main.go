package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/IsJordanBraz/go-multithreading/internal/entity"
)

func main() {
	http.HandleFunc("/{cep}", BuscaCepHandler)
	http.ListenAndServe(":8080", nil)
}
func BuscaCepHandler(w http.ResponseWriter, r *http.Request) {
	cep := r.PathValue("cep")

	if cep == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ch1 := make(chan entity.ViaCepMessage)
	ch2 := make(chan entity.BrasilApiMessage)

	go func() {
		req, err := http.Get("http://viacep.com.br/ws/" + cep + "/json/")

		if err != nil {
			fmt.Println("error while requesting viacep: " + err.Error())
		}

		defer req.Body.Close()

		res, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Println("error while reading viacep: " + err.Error())
		}

		var viacep entity.ViaCepMessage
		err = json.Unmarshal(res, &viacep)

		if err != nil {
			fmt.Println("error while Unmarshal viacep: " + err.Error())
		}

		ch1 <- viacep
		close(ch1)
	}()

	go func() {
		req, err := http.Get("https://brasilapi.com.br/api/cep/v1/" + cep)
		if err != nil {
			fmt.Println("error while requesting brasilapi: " + err.Error())
		}
		defer req.Body.Close()

		res, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Println("error while reading brasilapi: " + err.Error())
		}

		var brasilapi entity.BrasilApiMessage
		err = json.Unmarshal(res, &brasilapi)

		if err != nil {
			fmt.Println("error while Unmarshal brasilapi: " + err.Error())
		}

		ch2 <- brasilapi
	}()

	var message string

	select {
	case msg1 := <-ch1:
		message = fmt.Sprintf("Enviado por: viacep, Endereco: %s, %s, %s, %s - %s", msg1.Logradouro, msg1.Bairro, msg1.Localidade, msg1.Estado, msg1.Cep)
	case msg2 := <-ch2:
		message = fmt.Sprintf("Enviado por: brasilapi, Endereco: %s, %s, %s, %s - %s", msg2.Street, msg2.Neighborhood, msg2.City, msg2.State, msg2.Cep)
	case <-time.After(time.Second):
		message = "timeout"
	}

	fmt.Println(message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}
