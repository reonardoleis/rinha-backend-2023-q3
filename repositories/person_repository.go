package repositories

import (
	"context"
	"fmt"
	"log"

	"github.com/reonardoleis/rinha-backend-2023/cache"
	"github.com/reonardoleis/rinha-backend-2023/db"
	"github.com/reonardoleis/rinha-backend-2023/models"
)

type PersonRepository struct {
	DB    *db.Database
	Cache *cache.Cache
}

var singleton *PersonRepository

func PersonRepositoryInstance() (*PersonRepository, error) {
	if singleton != nil {
		return singleton, nil
	}

	db, err := db.Instance()
	if err != nil {
		log.Println(err)
		return nil, nil
	}

	cache, err := cache.Instance()
	if err != nil {
		log.Println(err)
		return nil, nil
	}

	singleton = &PersonRepository{
		db,
		cache,
	}

	return singleton, nil
}

func (p PersonRepository) InsertPerson(person *models.Person) error {
	err := p.Cache.SetPerson(string(person.ID), person)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (p PersonRepository) CreatePeople(people []*models.Person) error {
	query := "INSERT INTO person (id, nickname, name, birth_date, stack) VALUES "
	args := make([]interface{}, 0)
	for _, person := range people {
		args = append(args, person.ID, person.Nickname, person.Name, person.BirthDate, person.Stack)
		l := len(args)
		query += fmt.Sprintf(`($%d, $%d, $%d, $%d, $%d),`, l-4, l-3, l-2, l-1, l)
	}

	query = query[:len(query)-1]
	_, err := p.DB.Conn.Exec(context.Background(), query, args...)
	if err != nil {
		log.Println(err)
		return err
	}

	p.Cache.CleanAllTermSearch()

	return nil
}

func (p PersonRepository) FindPerson(id string) (*models.Person, error) {
	person, isCached, err := p.Cache.GetPersonByID(id)
	if err != nil {
		log.Println(err)
	} else {
		if isCached {
			return person, nil
		}
	}

	person = &models.Person{
		ID: db.CustomUUID(id),
	}
	query := `SELECT name, nickname, birth_date, stack FROM person
			  WHERE id = $1`

	err = p.DB.Conn.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(&person.Name, &person.Nickname, &person.BirthDate, &person.Stack)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, err
		}
		log.Println(err)
		return nil, err
	}

	p.Cache.SetPerson(id, person)

	return person, nil
}

func (p PersonRepository) SearchPeople(term string) ([]*models.Person, error) {
	people, isCached, err := p.Cache.GetTermSearch(term)
	if err != nil {
		log.Println(err)
	} else {
		if isCached {
			return people, nil
		}
	}

	query := fmt.Sprintf(
		`SELECT id, nickname, name, birth_date, stack FROM person WHERE
	idx @@ to_tsquery('simple', '%s:*')
	ORDER BY id DESC
	LIMIT 50`,
		term,
	)

	rows, err := p.DB.Conn.Query(context.Background(), query)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for rows.Next() {
		person := &models.Person{}
		err := rows.Scan(&person.ID, &person.Nickname, &person.Name, &person.BirthDate, &person.Stack)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		people = append(people, person)
	}

	p.Cache.SetTermSearch(term, people)

	return people, nil
}

func (p PersonRepository) CountPeople() (uint, error) {
	query := `SELECT COUNT(id) FROM person`

	var count uint
	err := p.DB.Conn.QueryRow(context.Background(), query).Scan(&count)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return count, nil
}

func (p PersonRepository) PersonExists(nickname string) (bool, error) {
	_, exists, err := p.Cache.GetPersonByNickname(nickname)
	if err != nil {
		log.Println(err)
	}

	if err == nil {
		if exists {
			return exists, nil
		}
	}

	query := `SELECT id FROM person WHERE nickname = $1 LIMIT 1`

	var id string

	err = p.DB.Conn.QueryRow(context.Background(), query, nickname).Scan(&id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil
		}

		log.Println(err)
		return false, err
	}

	return true, nil
}
