package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BrasilAPI struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type ViaCep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
}

type ApiResponse struct {
	API  string
	Data interface{}
	Err  error
}

func main() {
	http.HandleFunc("/", BuscaCepHandler)
 	http.ListenAndServe(":8080", nil)
}

func BuscaCepHandler(w http.ResponseWriter, r *http.Request) {
	cep := r.URL.Query().Get("cep")
	if cep == "" {
		http.Error(w, "CEP não fornecido", http.StatusBadRequest)
		return
	}

	ch := make(chan ApiResponse, 2)

	go GetBrasilAPI(cep, ch)
	go GetViaCep(cep, ch)

	select {
	case result := <-ch:
		if result.Err != nil {
			fmt.Printf("Erro na API %s: %v\n", result.API, result.Err)
			return
		}

		fmt.Printf("\nResposta da API mais rápida: %s\n\n", result.API)
		switch res := result.Data.(type) {
		case BrasilAPI:
			fmt.Printf("CEP: %s\nRua: %s\nBairro: %s\nCidade: %s\nEstado: %s\n",
				res.Cep, res.Street, res.Neighborhood, res.City, res.State)
		case ViaCep:
			fmt.Printf("CEP: %s\nRua: %s\nBairro: %s\nCidade: %s\nEstado: %s\n",
				res.Cep, res.Logradouro, res.Bairro, res.Localidade, res.Uf)
		}

	case <-time.After(1 * time.Second):
		fmt.Println("\nErro: Tempo limite excedido!")
	}
}

func GetBrasilAPI(cep string, ch chan<- ApiResponse) {
	url := "https://brasilapi.com.br/api/cep/v1/" + cep

	resp, err := http.Get(url)
	if err != nil {
		ch <- ApiResponse{API: "BrasilAPI", Err: err}
		return
	}
	defer resp.Body.Close()

	var data BrasilAPI
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		ch <- ApiResponse{API: "BrasilAPI", Err: err}
		return
	}
	
	ch <- ApiResponse{API: "BrasilAPI", Data: data}
}

func GetViaCep(cep string, ch chan<- ApiResponse) {
	url := "http://viacep.com.br/ws/" + cep + "/json/"

	resp, err := http.Get(url)
	if err != nil {
		ch <- ApiResponse{API: "ViaCEP", Err: err}
		return
	}
	defer resp.Body.Close()

	var data ViaCep
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		ch <- ApiResponse{API: "ViaCEP", Err: err}
		return
	}

	ch <- ApiResponse{API: "ViaCEP", Data: data}
}