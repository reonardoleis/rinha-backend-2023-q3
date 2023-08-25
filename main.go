package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	person_controller "github.com/reonardoleis/rinha-backend-2023/controllers"
	"github.com/reonardoleis/rinha-backend-2023/db"
	"github.com/reonardoleis/rinha-backend-2023/queue"
)

func main() {
	godotenv.Overload(".env")

	err := db.Connect()
	if err != nil {
		log.Fatalln(err)
	}

	if os.Getenv("IS_QUEUE") == "false" {
		controller, err := person_controller.Instance()
		if err != nil {
			log.Fatalln(err)
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				controller.CreatePerson(w, r)
			} else if r.Method == http.MethodGet {
				if r.URL.Path == "/pessoas" {
					controller.SearchPeople(w, r)
				} else if strings.Contains(r.URL.Path, "/pessoas/") {
					controller.GetPerson(w, r)
				} else {
					controller.CountPeople(w, r)
				}
			}
		})

		err = http.ListenAndServe(":80", nil)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		queue, err := queue.Instance()
		if err != nil {
			log.Fatalln(err)
		}

		queue.Init()
	}
}
