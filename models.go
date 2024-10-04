package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type RepositoryEntity struct {
	Id            float64                `json:"id"`
	Name          string                 `json:"name"`
	Full_name     string                 `json:"full_name"`
	Languages_url string                 `json:"languages_url"`
	Languages     map[string]interface{} `json:"languages"`
	Description   string                 `json:"description"`
	Readme_url    string                 `json:"readme_url"`
	Readme        string                 `json:"readme"`
	Updated_at    time.Time              `json:"updated_at"`
}

type Repositories struct {
	m  sync.Mutex
	fn string
}

type Url struct {
	Id   float64 `json:"id"`
	Url  string  `json:"url"`
	Type int     `json:"type"`
}

type Urls struct {
	m  sync.Mutex
	fn string
}

func (r *Repositories) init(fn string) {
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		file, err := os.Create(fn)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		r.fn = fn
	} else {
		r.fn = fn
	}
}

func (r *Repositories) open() *os.File {
	file, err := os.Open(r.fn)
	if err != nil {
		panic(err)
	}
	return file
}

func (r *Repositories) readJson() []RepositoryEntity {
	file := r.open()
	defer file.Close()
	var repos []map[string]interface{}
	err := json.NewDecoder(file).Decode(&repos)
	if err != nil {
		fmt.Println("Error:", err)
		return []RepositoryEntity{}
	}
	var re []RepositoryEntity
	for _, v := range repos {
		updated_at, err := time.Parse(time.RFC3339, v["updated_at"].(string))
		if err != nil {
			updated_at = time.Now()
		}
		languages := make(map[string]interface{})
		if v["languages"] != nil {
			languages = v["languages"].(map[string]interface{})
		}
		re = append(re, RepositoryEntity{
			Id:            v["id"].(float64),
			Name:          v["name"].(string),
			Full_name:     v["full_name"].(string),
			Description:   v["description"].(string),
			Languages_url: v["languages_url"].(string),
			Languages:     languages,
			Readme_url:    v["readme_url"].(string),
			Updated_at:    updated_at,
			Readme:        v["readme"].(string),
		})
	}
	return re
}

func (r *Repositories) readJsonSafely() []RepositoryEntity {
	r.m.Lock()
	defer r.m.Unlock()
	return r.readJson()
}

func (r *Repositories) writeJson(repos []RepositoryEntity) {
	if err := os.Truncate(r.fn, 0); err != nil {
		panic(err)
	}
	file, err := os.OpenFile(r.fn, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(repos)
	if err != nil {
		panic(err)
	}
}

func (r *Repositories) lock() {
	r.m.Lock()
}

func (r *Repositories) unlock() {
	r.m.Unlock()
}

func (u *Urls) init(fn string) {
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		file, err := os.Create(fn)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		u.fn = fn
	} else {
		u.fn = fn
	}
}

func (u *Urls) open() *os.File {
	file, err := os.Open(u.fn)
	if err != nil {
		panic(err)
	}
	return file
}

func (u *Urls) readJson() []Url {
	file := u.open()
	defer file.Close()
	var urls []Url
	err := json.NewDecoder(file).Decode(&urls)
	if err != nil {
		fmt.Println("Error:", err)
		return []Url{}
	}
	return urls
}

func (u *Urls) writeJson(urls []Url) {
	if err := os.Truncate(u.fn, 0); err != nil {
		panic(err)
	}
	file, err := os.OpenFile(u.fn, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(urls)
	if err != nil {
		panic(err)
	}
}

func (u *Urls) lock() {
	u.m.Lock()
}

func (u *Urls) unlock() {
	u.m.Unlock()
}

func (u *Urls) addUrls(urls []Url) {
	ur := u.readJson()
	for _, v := range urls {
		ur = append(ur, v)
	}
	u.writeJson(ur)
}

func (u *Urls) takeUrls(n int) []Url {
	ur := u.readJson()
	if len(ur) < n {
		n = len(ur)
	}
	urls := ur[:n]
	ur = ur[n:]
	u.writeJson(ur)
	return urls
}

func nameExists(r []RepositoryEntity, name string) bool {
	for _, v := range r {
		if v.Name == name {
			return true
		}
	}
	return false
}

func olderThan(u []RepositoryEntity, id float64, t time.Time) bool {
	for _, v := range u {
		if v.Id == id && v.Updated_at.Before(t) {
			return true
		}
	}
	return false
}

func setUpdatedAt(u []RepositoryEntity, id float64, t time.Time) {
	for i, v := range u {
		if v.Id == id {
			u[i].Updated_at = t
		}
	}
}

func setRepository(u []RepositoryEntity, id float64, r RepositoryEntity) {
	for i, v := range u {
		if v.Id == id {
			u[i].Name = r.Name
			u[i].Full_name = r.Full_name
			u[i].Description = r.Description
		}
	}
}

func languageExists(u []Url, id float64) bool {
	for _, v := range u {
		if v.Id == id && v.Type == Language {
			return true
		}
	}
	return false
}

func readmeExists(u []Url, id float64) bool {
	for _, v := range u {
		if v.Id == id && v.Type == Readme {
			return true
		}
	}
	return false
}

type Repository struct {
	Id          float64                `json:"id"`
	Name        string                 `json:"name"`
	Full_name   string                 `json:"full_name"`
	Description string                 `json:"description"`
	Languages   map[string]interface{} `json:"languages"`
	Readme      string                 `json:"readme"`
}

func toRepositories(r []RepositoryEntity) []Repository {
	var repos []Repository
	for _, v := range r {
		repos = append(repos, Repository{
			Id:          v.Id,
			Name:        v.Name,
			Full_name:   v.Full_name,
			Description: v.Description,
			Languages:   v.Languages,
			Readme:      v.Readme,
		})
	}
	return repos
}
