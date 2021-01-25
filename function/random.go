package function

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	firebase "firebase.google.com/go"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/api/option"
)

const (
	WORK   = "work"
	WEALTH = "wealth"
	HEALTH = "health"
	LOVE   = "love"
	STUDY  = "study"
)

type QueryResult struct {
	Action     string `json:"action,omitempty"`
	Parameters struct {
		Topic string `json:"Topic,omitempty"`
	} `json:"Parameters"`
}

type WebhookRequest struct {
	Session     string      `json:"string,omitempty"`
	ResponseId  string      `json:"responseId,omitempty"`
	QueryResult QueryResult `json:"queryResult"`
}

type Tarot struct {
	Name   string `firestore:"name,omitempty"`
	ImgURL string `firestore:"imgURL,omitempty"`
	Topic  struct {
		Health string `firestore:"health,omitempty"`
		Love   string `firestore:"love,omitempty"`
		Study  string `firestore:"study,omitempty"`
		Wealth string `firestore:"wealth,omitempty"`
		Work   string `firestore:"work,omitempty"`
	} `firestore:"topic,omitempty"`
}

type Reply struct {
	FulfillmentText string  `json:"fulfillmentText,omitempty"`
	Payload         Payload `json:"payload,omitempty"`
}

type Payload struct {
	Line Line `json:"line,omitempty"`
}

type Line struct {
	Type string `json:"type,omitempty"`
	Text string `json:"Text,omitempty"`
}

func RandomCard(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = httprouter.ParamsFromContext(r.Context())

	defer r.Body.Close()

	var message WebhookRequest
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	opt := option.WithCredentialsFile("config/serviceAccountKey.json")
	config := &firebase.Config{ProjectID: "darris-tarot"}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	fmt.Println("Pass")
	client, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	cardNo := rand.Intn(1)
	result, err := client.Collection("tarots").Doc(fmt.Sprint(cardNo)).Get(context.Background())
	if err != nil {
		log.Fatalf("error fetching database: %v\n", err)
	}

	var tarot Tarot
	err = result.DataTo(&tarot)
	if err != nil {
		log.Fatalf("error mapped data: %v\n", err)
	}

	var topic, text string
	switch t := message.QueryResult.Parameters.Topic; t {
	case WORK:
		topic, text = "การงาน", tarot.Topic.Work
	case WEALTH:
		topic, text = "การเงิน", tarot.Topic.Wealth
	case HEALTH:
		topic, text = "สุขภาพ", tarot.Topic.Health
	case STUDY:
		topic, text = "การเรียน", tarot.Topic.Study
	case LOVE:
		topic, text = "ความรัก", tarot.Topic.Love
	}

	resp, err := newReply(tarot.Name, tarot.ImgURL, topic, text)
	if err != nil {
		log.Fatalf("error reponse: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func newReply(cardName, imgURL, topic, text string) ([]byte, error) {
	// reply := &Reply{
	// 	FulfillmentText: text,
	// 	Payload: Payload{
	// 		Line: Line{
	// 			Type: "text",
	// 			Text: "ความรัก",
	// 		},
	// 	},
	// }

	reply := fmt.Sprintf(`
	{
		"fulfillmentMessages":[
		   {
			  "payload":{
				 "line":{
					"type":"flex",
					"altText":"Flex Message",
					"contents":{
					   "type":"bubble",
					   "direction":"ltr",
					   "header":{
						  "type":"box",
						  "layout":"vertical",
						  "contents":[
							 {
								"type":"text",
								"text":"%s",
								"size":"xl",
								"align":"center",
								"weight":"bold"
							 }
						  ]
					   },
					   "hero":{
						  "type":"image",
						  "url":"%s",
						  "size":"full",
						  "aspectRatio":"1.51:1",
						  "aspectMode":"fit"
					   },
					   "body":{
						  "type":"box",
						  "layout":"vertical",
						  "contents":[
							 {
								"type":"text",
								"text":"%s",
								"size":"xl",
								"align":"center",
								"weight":"bold"
							 },
							 {
								"type":"text",
								"text":"%s",
								"align":"center"
							 }
						  ]
					   }
					}
				 }
			  }
		   }
		]
	 }

	`, cardName, imgURL, topic, text)
	fmt.Println(reply)
	return []byte(reply), nil
	// resp, err := json.Marshal(reply)
	// if err != nil {
	// 	return nil, err
	// }
	// return resp, nil
}
