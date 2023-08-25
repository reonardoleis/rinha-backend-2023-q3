package queue

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/reonardoleis/rinha-backend-2023/cache"
	"github.com/reonardoleis/rinha-backend-2023/models"
	"github.com/reonardoleis/rinha-backend-2023/repositories"
	"github.com/reonardoleis/rinha-backend-2023/utils"
)

type Queue struct {
	personRepository *repositories.PersonRepository
	cache            *cache.Cache
	lastRun          int64
	C                chan *models.Person
}

var singleton *Queue

var exponentialWaitSeconds = 2
var exponentialWaitExp = 2
var exponentialWaitMax = 16

func Instance() (*Queue, error) {
	if singleton != nil {
		return singleton, nil
	}

	personRepository, err := repositories.PersonRepositoryInstance()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	cache, err := cache.Instance()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	singleton = &Queue{
		personRepository,
		cache,
		time.Now().UnixNano(),
		make(chan *models.Person, 10_000),
	}

	return singleton, nil
}

func (q *Queue) shouldInsertByInterval() bool {
	intervalSeconds := utils.GetIntEnv("INSERT_BATCH_INTERVAL", 1)
	intervalNano := int64(intervalSeconds * 1000000000)

	shouldInsert := q.lastRun < time.Now().UnixNano()-intervalNano

	return shouldInsert
}

func (q *Queue) Enqueue(person []*models.Person) error {
	var jsons = make([]interface{}, len(person))
	for idx, p := range person {
		json := p.JSON()
		jsons[idx] = json
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := q.cache.Client().LPush(ctx, "queue", jsons).Err()
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (q *Queue) PopFirstN(n int) ([]*models.Person, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	people := make([]*models.Person, 0)
	for i := 0; i < n; i++ {
		val, err := q.cache.Client().RPop(ctx, "queue").Result()
		if err != nil {
			if err == redis.Nil {
				break
			}

			log.Println(err)
			return nil, err
		}

		person := &models.Person{}

		err = person.FromJSON([]byte(val))
		if err != nil {
			log.Println(err)
			return nil, err
		}

		people = append(people, person)
	}

	return people, nil
}

func (q *Queue) QueueSize() int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	size, err := q.cache.Client().LLen(ctx, "queue").Result()
	if err != nil {
		log.Println(err)
		return 0
	}

	return int(size)
}

func (q *Queue) Init() {
	for {
		time.Sleep(1000 * time.Millisecond)
		size := q.QueueSize()
		batchSize := utils.GetIntEnv("INSERT_BATCH_SIZE", 10)

		if size >= batchSize || (size != 0 && q.shouldInsertByInterval()) {
			people, err := q.PopFirstN(batchSize)
			if err != nil {
				val := math.Pow(float64(exponentialWaitSeconds), float64(exponentialWaitExp))
				time.Sleep(time.Duration(int(val) * int(time.Second)))

				if val <= float64(exponentialWaitMax) {
					exponentialWaitExp++
				}

				log.Println(err)
				continue
			}

			err = q.personRepository.CreatePeople(people)
			if err != nil {
				q.Enqueue(people)
				log.Println(err)
			} else {
				q.lastRun = time.Now().UnixNano()
				exponentialWaitExp = 2
			}

		}
	}

}

func (q *Queue) SendToMonitor(person *models.Person) bool {
	if len(q.C) >= cap(q.C) {
		return false
	}

	q.C <- person
	return true
}

func (q *Queue) MonitorSetAndEnqueue() {
	for {
		person, ok := <-q.C

		if !ok {
			continue
		}

		q.cache.SetPerson(string(person.ID), person)
		q.Enqueue([]*models.Person{person})
	}
}
