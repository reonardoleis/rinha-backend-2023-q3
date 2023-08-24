package person_controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gofrs/uuid"
	"github.com/reonardoleis/rinha-backend-2023/db"
	"github.com/reonardoleis/rinha-backend-2023/models"
	"github.com/reonardoleis/rinha-backend-2023/queue"
	"github.com/reonardoleis/rinha-backend-2023/repositories"
)

type PersonController struct {
	personRepository *repositories.PersonRepository
	queue            *queue.Queue
}

var singleton *PersonController

func PersonControllerInstance() (*PersonController, error) {
	if singleton != nil {
		return singleton, nil
	}

	personRepository, err := repositories.PersonRepositoryInstance()
	if err != nil {
		return nil, err
	}

	queue, err := queue.Instance()
	if err != nil {
		return nil, err
	}

	singleton = &PersonController{
		personRepository,
		queue,
	}

	return singleton, nil
}

func (pc *PersonController) CreatePerson(w http.ResponseWriter, r *http.Request) {
	person := &models.Person{}

	err := json.
		NewDecoder(r.Body).
		Decode(person)
	if err != nil {
		log.Println("error decoding person", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if person.Nickname == "" || person.BirthDate == "" || person.Name == "" {
		log.Println("error validating person", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	birthdateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	if utf8.RuneCountInString(person.Nickname) > 32 ||
		utf8.RuneCountInString(person.Name) > 100 ||
		!birthdateRegex.MatchString(string(person.BirthDate)) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	for _, stack := range person.Stack {
		stackLen := utf8.RuneCountInString(stack)
		if stackLen == 0 || stackLen > 32 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
	}

	if person.Stack == nil {
		person.Stack = make([]string, 0)
	}

	personExists, err := pc.personRepository.PersonExists(person.Nickname)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if personExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	} else {
		uuid, err := uuid.NewV4()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		person.ID = db.CustomUUID(uuid.String())

		err = pc.personRepository.InsertPerson(person)
		if err != nil {
			log.Println(err)
		}
	}

	alreadyEnqueued, err := pc.queue.Enqueue([]*models.Person{person})
	if err != nil {
		log.Println(err)
	}

	if alreadyEnqueued {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/pessoas/%s", person.ID))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte{})
}

func (pc *PersonController) GetPerson(w http.ResponseWriter, r *http.Request) {
	splittedURL := strings.Split(r.URL.Path, "/")
	id := splittedURL[len(splittedURL)-1]

	person, err := pc.personRepository.FindPerson(id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json, err := person.ToJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (pc *PersonController) SearchPeople(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	term := query.Get("t")
	if term == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	people, err := pc.personRepository.SearchPeople(term)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(people)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (pc *PersonController) CountPeople(w http.ResponseWriter, r *http.Request) {
	count, err := pc.personRepository.CountPeople()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%d", count)))
}
