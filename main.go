package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const (
	expectedToken = "my_secret_token"
)

type TestResult struct {
	IdTest        int    `json:"id_test"`
	DeliveryToken int    `json:"delivery_token"`
	Token         string `json:"token"`
}

type RequestBody struct {
	IdTest int `json:"id_test"`
}

// Функция для преобразования числа в соответствующее слово
func getStatusWord(status int) string {

	switch status {
	case 1:
		return "Успех"
	case 2:
		return "Неуспех"
	default:
		return "Неизвестный статус"
	}
}

func main() {
	http.HandleFunc("/api/async_calc/", handleProcess)
	fmt.Println("Server running at port :5000")
	http.ListenAndServe(":5000", nil)
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "The method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody RequestBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		fmt.Println("Error decoding JSON:", err)
		return
	}

	id_test := requestBody.IdTest
	fmt.Println(id_test)

	var delivery_token int = rand.Intn(9000) + 1000

	// Успешный ответ в формате JSON
	successMessage := map[string]interface{}{
		"message":        "Successful",
		"delivery_token": getStatusWord(delivery_token),
		"data": TestResult{
			IdTest:        id_test,
			DeliveryToken: delivery_token,
			Token:         expectedToken,
		},
	}

	jsonResponse, err := json.Marshal(successMessage)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		fmt.Println("Ошибка кодирования JSON:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

	go func() {
		// Задержка 5 секунд
		delay := 5
		time.Sleep(time.Duration(delay) * time.Second)

		// Отправка результата на другой сервер
		result := TestResult{
			IdTest:        id_test,
			DeliveryToken: delivery_token,
			Token:         expectedToken,
		}

		fmt.Println("json", result)
		jsonValue, err := json.Marshal(result)
		if err != nil {
			fmt.Println("Ошибка при маршализации JSON:", err)
			return
		}
		resultURL := "http://localhost:8000/async_token/" + fmt.Sprint(requestBody.IdTest) + "/"
		req, err := http.NewRequest(http.MethodPut, resultURL, bytes.NewBuffer(jsonValue))
		if err != nil {
			fmt.Println("Ошибка при создании запроса на обновление:", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Ошибка при отправке запроса на обновление:", err)
			return
		}
		defer resp.Body.Close()

		fmt.Println("Ответ от сервера обновления:", resp.Status)
	}()
}
