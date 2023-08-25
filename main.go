package main

import (
	"log"
	"net/http"
	"os"

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
		queue, err := queue.Instance()
		if err != nil {
			log.Fatalln(err)
		}

		for i := 0; i < 4; i++ {
			go queue.MonitorSetAndEnqueue()
		}

		controller, err := person_controller.PersonControllerInstance()
		if err != nil {
			log.Fatalln(err)
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/pessoas", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				controller.CreatePerson(w, r)
			} else {
				controller.SearchPeople(w, r)
			}
		})

		mux.HandleFunc("/pessoas/", controller.GetPerson)

		mux.HandleFunc("/contagem-pessoas", controller.CountPeople)

		err = http.ListenAndServe(":80", mux)
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
