package queue

import (
	"log"
	"sync"
	"time"

	"github.com/reonardoleis/rinha-backend-2023/models"
	"github.com/reonardoleis/rinha-backend-2023/repositories"
	"github.com/reonardoleis/rinha-backend-2023/utils"
)

type Queue struct {
	personRepository *repositories.PersonRepository
	pendingInsertion []*models.Person
	lock             sync.Mutex
	lastRun          int64
}

var singleton *Queue

func Instance() (*Queue, error) {
	if singleton != nil {
		return singleton, nil
	}

	personRepository, err := repositories.PersonRepositoryInstance()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	singleton = &Queue{
		personRepository,
		make([]*models.Person, 0),
		sync.Mutex{},
		time.Now().UnixNano(),
	}

	return singleton, nil
}

func (q *Queue) Enqueue(person *models.Person) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.pendingInsertion = append(q.pendingInsertion, person)
}

func (q *Queue) PendingSize() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.pendingInsertion)
}

func (q *Queue) PopFirstN(n int) []*models.Person {
	q.lock.Lock()
	defer q.lock.Unlock()

	if n > len(q.pendingInsertion) {
		n = len(q.pendingInsertion)
	}

	pending := q.pendingInsertion[:n]
	q.pendingInsertion = q.pendingInsertion[n:]

	return pending
}

func (q *Queue) shouldInsertByInterval() bool {
	intervalSeconds := utils.GetIntEnv("INSERT_BATCH_INTERVAL", 1)
	intervalNano := int64(intervalSeconds * 1000000000)

	shouldInsert := q.lastRun < time.Now().UnixNano()-intervalNano

	return shouldInsert
}

func (q *Queue) ContainsNickname(nickname string) bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	for _, person := range q.pendingInsertion {
		if person.Nickname == nickname {
			return true
		}
	}

	return false
}

func (q *Queue) Init() {
	go func() {
		for {
			q.shouldInsertByInterval()
			time.Sleep(1 * time.Second)
			size := q.PendingSize()
			batchSize := utils.GetIntEnv("INSERT_BATCH_SIZE", 10)
			if size >= batchSize || (size != 0 && q.shouldInsertByInterval()) {
				pending := q.PopFirstN(batchSize)
				err := q.personRepository.CreatePeople(pending)
				if err != nil {
					q.pendingInsertion = append(pending, q.pendingInsertion...)
					log.Println(err)
				} else {
					q.lastRun = time.Now().UnixNano()
				}

			}
		}
	}()
}
