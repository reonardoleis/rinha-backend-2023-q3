package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/reonardoleis/rinha-backend-2023/models"
	"github.com/reonardoleis/rinha-backend-2023/utils"
)

type Cache struct {
	client *redis.Client
}

var singleton *Cache

func Instance() (*Cache, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if singleton != nil {
		return singleton, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	singleton = &Cache{
		client,
	}

	return singleton, nil
}

func (c *Cache) SetPerson(key string, person *models.Person) error {
	json, err := person.ToJSON()
	if err != nil {
		log.Println(err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = c.client.Set(ctx, key, json, utils.GetCacheDurationEnv()*time.Second).Err()
	if err != nil {
		log.Println(err)
		return err
	}

	err = c.client.Set(ctx, person.Nickname, key, utils.GetCacheDurationEnv()*time.Second).Err()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (c *Cache) GetPersonByID(key string) (*models.Person, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}

		log.Println(err)
		return nil, false, err
	}

	person := &models.Person{}

	err = person.FromJSON([]byte(val))
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	return person, true, nil
}

func (c *Cache) GetPersonByNickname(nickname string) (*models.Person, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	userID, err := c.client.Get(ctx, nickname).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}

		log.Println(err)
		return nil, false, err
	}

	val, err := c.client.Get(ctx, userID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}

		log.Println(err)
		return nil, false, err
	}

	person := &models.Person{}

	err = person.FromJSON([]byte(val))
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	return person, true, nil
}

func (c *Cache) PersonExists(nickname string) (bool, error) {
	id, err := c.client.Get(context.Background(), nickname).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}

		log.Println(err)
	}

	_, err = c.client.Exists(context.Background(), id).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
	}

	return true, nil
}

func (c *Cache) SetTermSearch(term string, people []*models.Person) error {
	json, err := json.Marshal(people)
	if err != nil {
		log.Println(err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = c.client.Set(ctx, fmt.Sprintf("TERM_%s", term), json, utils.GetCacheDurationEnv()*time.Second).Err()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (c *Cache) GetTermSearch(term string) ([]*models.Person, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	val, err := c.client.Get(ctx, fmt.Sprintf("TERM_%s", term)).Result()
	if err != nil {
		if err == redis.Nil {
			return []*models.Person{}, false, nil
		}

		log.Println(err)
		return []*models.Person{}, false, err
	}
	people := []*models.Person{}

	err = json.Unmarshal([]byte(val), &people)
	if err != nil {
		log.Println(err)
		return nil, false, err
	}

	return people, true, nil
}

func (c *Cache) CleanAllTermSearch() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	iter := c.client.Scan(ctx, 0, "TERM_*", 0).Iterator()
	for iter.Next(ctx) {
		err := c.client.Del(ctx, iter.Val()).Err()
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (c *Cache) Client() *redis.Client {
	return c.client
}
