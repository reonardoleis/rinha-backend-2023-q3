package person_controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

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

func Instance() (*PersonController, error) {
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

	var generatedUUID uuid.UUID
	var personExists bool

	personExists, err = pc.personRepository.PersonExists(person.Nickname)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if personExists {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	generatedUUID, err = uuid.NewV4()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	birthdateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	if len(person.Nickname) > 32 ||
		len(person.Name) > 100 ||
		!birthdateRegex.MatchString(string(person.BirthDate)) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if person.Stack == nil {
		person.Stack = make([]string, 0)
	}

	for _, stack := range person.Stack {
		stackLen := len(stack)
		if stackLen == 0 || stackLen > 32 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
	}

	person.ID = db.CustomUUID(generatedUUID.String())

	err = pc.queue.Enqueue([]*models.Person{person})
	if err != nil {
		log.Println(err)
	}

	err = pc.personRepository.InsertPerson(person)
	if err != nil {
		log.Println(err)
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

	w.WriteHeader(http.StatusOK)
	w.Write(person.JSON())
}

func (pc *PersonController) SearchPeople(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	term := query.Get("t")

	termLen := len(term)
	if termLen == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if termLen > 100 {
		w.Write([]byte("[]"))
		return
	}

	people, err := pc.personRepository.SearchPeople(term, termLen)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(models.PersonListJSON(people))
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
