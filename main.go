package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", chatbotHandler)
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err.Error())
	}
}

func chatbotHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		verifyWebhook(w, r)
	case "POST":
		processWebhook(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Không hỗ trợ phương thức HTTP %v", r.Method)
	}
}

func verifyWebhook(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	challenge := r.URL.Query().Get("hub.challenge")
	token := r.URL.Query().Get("hub.verify_token")

	if mode == "subscribe" && token == "GoBot" {
		w.WriteHeader(200)
		w.Write([]byte(challenge))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Error, wrong validation token"))
	}
}

func processWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("Message not supported"))
		return
	}

	if req.Object == "page" {
		for _, entry := range req.Entry {
			for _, event := range entry.Messaging {
				if event.Message != nil {
					sendText(event.Sender, strings.ToUpper(event.Message.Text))
				}
			}
		}
		w.WriteHeader(200)
		w.Write([]byte("Got your message"))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Message not supported"))
	}
}

const (
	FBMessageURL    = "https://graph.facebook.com/v15.0/me/messages"
	PageToken       = "your_token"
	MessageResponse = "RESPONSE"
)

func sendText(recipient *User, message string) error {
	m := ResponseMessage{
		MessageType: MessageResponse,
		Recipient:   recipient,
		Message: &ResMessage{
			Text: message,
		},
	}

	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&m)
	if err != nil {
		log.Printf("Json: " + err.Error())
		return err
	}

	req, err := http.NewRequest("POST", FBMessageURL, body)
	if err != nil {
		log.Printf("http:" + err.Error())
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.URL.RawQuery = "access_token=" + PageToken
	client := &http.Client{Timeout: time.Second * 30}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("request: " + err.Error())
		return err
	}
	defer resp.Body.Close()

	return nil
}
