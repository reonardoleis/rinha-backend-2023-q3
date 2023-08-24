package queue

import (
	"log"
	"math"
	"time"

	"github.com/reonardoleis/rinha-backend-2023/cache"
	"github.com/reonardoleis/rinha-backend-2023/repositories"
	"github.com/reonardoleis/rinha-backend-2023/utils"
)

type Queue struct {
	personRepository *repositories.PersonRepository
	cache            *cache.Cache
	lastRun          int64
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
	}

	return singleton, nil
}

func (q *Queue) shouldInsertByInterval() bool {
	intervalSeconds := utils.GetIntEnv("INSERT_BATCH_INTERVAL", 1)
	intervalNano := int64(intervalSeconds * 1000000000)

	shouldInsert := q.lastRun < time.Now().UnixNano()-intervalNano

	return shouldInsert
}

func (q *Queue) Init() {
	for {
		time.Sleep(1000 * time.Millisecond)
		size := q.cache.QueueSize()
		batchSize := utils.GetIntEnv("INSERT_BATCH_SIZE", 10)

		if size >= batchSize || (size != 0 && q.shouldInsertByInterval()) {
			people, err := q.cache.PopFirstN(batchSize)
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
				q.cache.Enqueue(people)
				log.Println(err)
			} else {
				q.lastRun = time.Now().UnixNano()
				exponentialWaitExp = 2
			}

		}
	}

}
