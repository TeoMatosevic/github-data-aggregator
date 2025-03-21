package main

import (
	"encoding/json"
	"log"
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
	m sync.Mutex
}

type Url struct {
	Id     string  `json:"id"`
	RepoId float64 `json:"repoId"`
	Url    string  `json:"url"`
	Type   int     `json:"type"`
}

type Urls struct {
	m sync.Mutex
}

type Organization struct {
	Id          float64 `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Readme_url  string  `json:"readme_url"`
	Readme      string  `json:"readme"`
}

type Organizations struct {
	m sync.Mutex
}

func (o *Organizations) getCounter() int {
	o.m.Lock()
	defer o.m.Unlock()

	var count int
	err := db.QueryRow("SELECT count FROM counters WHERE type='urls'").Scan(&count)
	if err != nil {
		panic(err)
	}

	return count
}

func (r *Organizations) incrementCounter() {
	r.m.Lock()
	defer r.m.Unlock()

	_, err := db.Exec("UPDATE counters SET count=count+1 WHERE type='urls'")
	if err != nil {
		panic(err)
	}
}

func (o *Repositories) read() []RepositoryEntity {
	o.m.Lock()
	defer o.m.Unlock()

	var repos []map[string]interface{}
	rows, err := db.Query("SELECT * FROM repositories")
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		var id float64
		var name, full_name, languages_url, languages, description, readme_url, readme string
		var updated_at time.Time
		err = rows.Scan(&id, &name, &full_name, &languages_url, &languages, &description, &readme_url, &readme, &updated_at)
		if err != nil {
			panic(err)
		}
		repos = append(repos, map[string]interface{}{
			"id":            id,
			"name":          name,
			"full_name":     full_name,
			"languages_url": languages_url,
			"languages":     languages,
			"description":   description,
			"readme_url":    readme_url,
			"readme":        readme,
			"updated_at":    updated_at.Format(time.RFC3339),
		})
	}

	var re []RepositoryEntity
	for _, v := range repos {
		updated_at, err := time.Parse(time.RFC3339, v["updated_at"].(string))
		if err != nil {
			updated_at = time.Now()
		}
		languages := make(map[string]interface{})
		if v["languages"] != nil {
			languages = make(map[string]interface{})
			err = json.Unmarshal([]byte(v["languages"].(string)), &languages)
			if err != nil {
				panic(err)
			}
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

func (r *Repositories) write(repo RepositoryEntity) {
	r.m.Lock()
	defer r.m.Unlock()

	languages, err := json.Marshal(repo.Languages)

	_, err = db.Exec("INSERT INTO repositories (id, name, full_name, languages_url, languages, description, readme_url, readme, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		repo.Id, repo.Name, repo.Full_name, repo.Languages_url, string(languages), repo.Description, repo.Readme_url, repo.Readme, repo.Updated_at,
	)
	if err != nil {
		panic(err)
	}

	log.Println("[INFO] Added repository", repo.Full_name)
}

func (r *Repositories) update(repo RepositoryEntity) {
	r.m.Lock()
	defer r.m.Unlock()

	languages, err := json.Marshal(repo.Languages)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`
		UPDATE repositories SET name=$1, full_name=$2, languages_url=$3, languages=$4, description=$5, readme_url=$6, readme=$7, updated_at=$8 WHERE id=$9
		`, repo.Name, repo.Full_name, repo.Languages_url, string(languages), repo.Description, repo.Readme_url, repo.Readme, repo.Updated_at, repo.Id,
	)
	if err != nil {
		panic(err)
	}

	log.Println("[INFO] Updated repository", repo.Full_name)
}

func (r *Repositories) remove(name string) {
	r.m.Lock()
	defer r.m.Unlock()

	_, err := db.Exec("DELETE FROM repositories WHERE name=$1", name)
	if err != nil {
		panic(err)
	}

	log.Println("[INFO] Removed repository", name)
}

func (u *Urls) read() []Url {
	u.m.Lock()
	defer u.m.Unlock()

	var urls []map[string]interface{}
	rows, err := db.Query("SELECT * FROM urls")
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		var id string
		var repoId float64
		var url string
		var t int
		err = rows.Scan(&id, &repoId, &url, &t)
		if err != nil {
			panic(err)
		}
		urls = append(urls, map[string]interface{}{
			"id":     id,
			"repoId": repoId,
			"url":    url,
			"type":   t,
		})
	}

	var ur []Url
	for _, v := range urls {
		ur = append(ur, Url{
			Id:     v["id"].(string),
			RepoId: v["repoId"].(float64),
			Url:    v["url"].(string),
			Type:   v["type"].(int),
		})
	}
	return ur
}

func (u *Urls) write(urls Url) {
	u.m.Lock()
	defer u.m.Unlock()

	_, err := db.Exec("INSERT INTO urls (id, repo_id, url, type) VALUES ($1, $2, $3, $4)", urls.Id, urls.RepoId, urls.Url, urls.Type)
	if err != nil {
		panic(err)
	}

	log.Println("[INFO] Added url", urls.Url)
}

func (u *Urls) take(n int) []Url {
	ur := u.read()
	if len(ur) < n {
		n = len(ur)
	}

	rows, err := db.Query(`
	DELETE FROM urls WHERE id IN (
		SELECT id FROM urls ORDER BY id LIMIT $1
	)
	RETURNING id, repo_id, url, type
	`, n)

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	var urls []map[string]interface{}
	for rows.Next() {
		var id string
		var repoId float64
		var url string
		var t int
		err = rows.Scan(&id, &repoId, &url, &t)
		if err != nil {
			panic(err)
		}
		urls = append(urls, map[string]interface{}{
			"id":     id,
			"repoId": repoId,
			"url":    url,
			"type":   t,
		})
	}

	var uro []Url
	for _, v := range urls {
		uro = append(uro, Url{
			Id:     v["id"].(string),
			RepoId: v["repoId"].(float64),
			Url:    v["url"].(string),
			Type:   v["type"].(int),
		})

		log.Println("[INFO] Took url", v["url"])
	}

	return uro
}

func (u *Urls) addUrls(urls []Url) {
	ur := u.read()
	for _, v := range urls {
		if !languageExists(ur, v.RepoId) && !readmeExists(ur, v.RepoId) {
			u.write(v)
		}
	}
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
		if v.RepoId == id && v.Type == Language {
			return true
		}
	}
	return false
}

func readmeExists(u []Url, id float64) bool {
	for _, v := range u {
		if v.RepoId == id && v.Type == Readme {
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
	Last_update time.Time              `json:"last_update"`
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
			Last_update: v.Updated_at,
		})
	}
	return repos
}

func (o *Organizations) nameExists(orgs []Organization, name string) bool {
	for _, v := range orgs {
		if v.Name == name {
			return true
		}
	}
	return false
}

func (o *Organizations) read() []Organization {
	o.m.Lock()
	defer o.m.Unlock()

	var orgs []map[string]interface{}
	rows, err := db.Query("SELECT * FROM organizations")
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		var id float64
		var name, description, readme_url, readme string
		err = rows.Scan(&id, &name, &description, &readme_url, &readme)
		if err != nil {
			panic(err)
		}
		orgs = append(orgs, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"readme_url":  readme_url,
			"readme":      readme,
		})
	}

	var or []Organization
	for _, v := range orgs {
		or = append(or, Organization{
			Id:          v["id"].(float64),
			Name:        v["name"].(string),
			Description: v["description"].(string),
			Readme_url:  v["readme_url"].(string),
			Readme:      v["readme"].(string),
		})
	}
	return or
}

func (o *Organizations) write(org Organization) {
	o.m.Lock()
	defer o.m.Unlock()

	_, err := db.Exec("INSERT INTO organizations (id, name, description, readme_url, readme) VALUES ($1, $2, $3, $4, $5)", org.Id, org.Name, org.Description, org.Readme_url, org.Readme)
	if err != nil {
		panic(err)
	}

	log.Println("[INFO] Added organization", org.Name)
}

func (o *Organizations) set(org Organization) {
	o.m.Lock()
	defer o.m.Unlock()

	_, err := db.Exec("UPDATE organizations SET name=$1, description=$2, readme_url=$3, readme=$4 WHERE id=$5", org.Name, org.Description, org.Readme_url, org.Readme, org.Id)
	if err != nil {
		panic(err)
	}

	log.Println("[INFO] Updated organization", org.Name)
}

func (o *Organizations) remove(name string) {
	o.m.Lock()
	defer o.m.Unlock()

	_, err := db.Exec("DELETE FROM organizations WHERE name=$1", name)
	if err != nil {
		panic(err)
	}

	log.Println("[INFO] Removed organization", name)
}
