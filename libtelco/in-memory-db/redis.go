/*
Package redis содержит в себе необходимое API для работы с in-memory базой данных Redis с использованием библиотеки
*/
package redis

import (
	cp "SchoolServer/libtelco/config-parser"
	"fmt"

	redis "gopkg.in/redis.v5"
)

// Database struct представляет абстрактную структуру базы данных.
type Database struct {
	schoolServerDB *redis.Client
}

// NewDatabase создает клиента базы данных.
func NewDatabase(config *cp.Config) (*Database, error) {
	sdb := redis.NewClient(&redis.Options{
		Addr: config.Redis.Host + ":" + config.Redis.Port,
		//Password: config.Redis.Password,
		DB: config.Redis.DBname,
	})
	if sdb == nil {
		return nil, fmt.Errorf("Unable to connect to Redis")
	}
	return &Database{
		sdb,
	}, nil
}

// FlushAll стирает ВСЕ
func (db *Database) FlushAll() error {
	return db.schoolServerDB.FlushAll().Err()
}

// Работа с cookie.

// AddCookie добавляет новый Cookie в Redis.
func (db *Database) AddCookie(key, value string) error {
	return db.schoolServerDB.Set(key, value, 0).Err()
}

// ExistsCookie проверяет, существует ли данный Cookie в БД.
func (db *Database) ExistsCookie(key string) (bool, error) {
	return db.schoolServerDB.Exists(key).Result()
}

// UpdateCookie обновляет информацию о Cookie, если он уже существует.
func (db *Database) UpdateCookie(key, value string) (bool, error) {
	var err error
	flag, err := db.ExistsCookie(key)
	if err != nil {
		return false, err
	}
	if !flag {
		return false, nil
	}
	err = db.AddCookie(key, value)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteCookie удаляет Cookie.
func (db *Database) DeleteCookie(key string) (bool, error) {
	flag, err := db.schoolServerDB.Del(key).Result()
	if flag == 1 {
		return true, err
	}
	return false, err
}

// GetCookie возвращает Cookie.
func (db *Database) GetCookie(key string) (string, error) {
	return db.schoolServerDB.Get(key).Result()
}

// Работа с файлами.

// AddFileDate добавляет новую дату в Redis.
func (db *Database) AddFileDate(key, value string) error {
	return db.schoolServerDB.Set(key, value, 0).Err()
}

// ExistsFileDate проверяет, существует ли данная дата в БД.
func (db *Database) ExistsFileDate(key string) (bool, error) {
	return db.schoolServerDB.Exists(key).Result()
}

// UpdateFileDate обновляет информацию о дате, если она уже сущетсвует.
func (db *Database) UpdateFileDate(key, value string) (bool, error) {
	var err error
	flag, err := db.ExistsFileDate(key)
	if err != nil {
		return false, err
	}
	if !flag {
		return false, nil
	}
	err = db.AddFileDate(key, value)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteFileDate удаляет дату.
func (db *Database) DeleteFileDate(key, value string) (bool, error) {
	flag, err := db.schoolServerDB.Del(key).Result()
	if flag == 1 {
		return true, err
	}
	return false, err
}

// GetFileDate возвращает дату.
func (db *Database) GetFileDate(key string) (string, error) {
	return db.schoolServerDB.Get(key).Result()
}

// Close закрывает клиента.
func (db *Database) Close() error {
	return db.schoolServerDB.Close()
}
