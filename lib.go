package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDatabase() {
	var err error
	db, err = sql.Open("sqlite3", "./repos.db")
	if err != nil {
		panic(err)
	}

	sqlStmt := `
	create table if not exists repositories (
		id integer not null primary key,
		name text,
		full_name text,
		languages_url text,
		languages text,
		description text,
		readme_url text, 
		readme text, 
		updated_at datetime
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	sqlStmt = `
	create table if not exists urls (id string not null primary key, repo_id integer, url text, type integer);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}
}

func getRepositories(r *Repositories, urls *Urls) ([]Url, error) {
	u, err := getRepoUrls(r)
	if err != nil {
		return nil, err
	}

	urls.addUrls(u)
	return u, nil
}

func sendRequests(r *Repositories, urls *Urls) []Url {
	numberOfUrls := 10

	u := urls.take(numberOfUrls)

	for _, v := range u {
		switch v.Type {
		case Language:
			addLanguage(r, v)
		case Readme:
			addReadme(r, v)
		}
	}

	return u
}

func addLanguage(r *Repositories, url Url) {
	rc := r.read()
	languages, err := getLanguages(url.Url)
	if err != nil {
		return
	}
	for i, vv := range rc {
		if vv.Id == url.RepoId {
			rc[i].Languages = languages
			log.Println("[INFO] Updated languages for", rc[i].Full_name)
			r.update(rc[i])
		}
	}
}

func addReadme(r *Repositories, url Url) {
	rc := r.read()
	readme, err := getReadme(url.Url)
	if err != nil {
		return
	}
	for i, vv := range rc {
		if vv.Id == url.RepoId {
			rc[i].Readme = readme
			log.Println("[INFO] Updated readme for", rc[i].Full_name)
			r.update(rc[i])
		}
	}
}

func unmarshalRepository(data []byte) ([]RepositoryEntity, error) {
	var r []RepositoryEntity
	var ud []map[string]interface{}
	var err = json.Unmarshal(data, &ud)
	if err != nil {
		empty := []RepositoryEntity{}
		return empty, nil
	}
	for _, v := range ud {
		updated_at, err := time.Parse(time.RFC3339, v["updated_at"].(string))
		if err != nil {
			updated_at = time.Now()
		}
		description := v["description"]
		if description == nil {
			description = ""
		}
		readMe := fmt.Sprintf("%s/contents/README.md", v["url"])
		r = append(r, RepositoryEntity{
			Id:            v["id"].(float64),
			Name:          v["name"].(string),
			Full_name:     v["full_name"].(string),
			Description:   description.(string),
			Languages_url: v["languages_url"].(string),
			Languages:     nil,
			Readme_url:    readMe,
			Updated_at:    updated_at,
			Readme:        "",
		})
	}
	return r, nil
}

func removeRepository(r *Repositories, name string) {
	r.remove(name)
}

func getRepoUrls(r *Repositories) ([]Url, error) {
	var newUrls []Url
	response, err := http.Get(reposApiURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	res, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	repos, err := unmarshalRepository(res)
	if err != nil {
		return nil, err
	}

	cr := r.read()

	for _, v := range repos {
		uuid_1 := uuid.New().String()
		uuid_2 := uuid.New().String()
		if !nameExists(cr, v.Name) {
			cr = append(cr, v)
			newUrls = append(newUrls, Url{Id: uuid_1, RepoId: v.Id, Url: v.Languages_url, Type: Language})
			newUrls = append(newUrls, Url{Id: uuid_2, RepoId: v.Id, Url: v.Readme_url, Type: Readme})
			r.write(v)
		} else {
			if olderThan(cr, v.Id, v.Updated_at) {
				setUpdatedAt(cr, v.Id, v.Updated_at)
				newUrls = append(newUrls, Url{Id: uuid_1, RepoId: v.Id, Url: v.Languages_url, Type: Language})
				newUrls = append(newUrls, Url{Id: uuid_2, RepoId: v.Id, Url: v.Readme_url, Type: Readme})
				r.update(v)
			}
			setRepository(cr, v.Id, v)
		}
	}

	for _, v := range cr {
		if !nameExists(repos, v.Name) {
			removeRepository(r, v.Name)
		}
	}

	return newUrls, nil
}

func getLanguages(u string) (map[string]interface{}, error) {
	response, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	res, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var l map[string]interface{}
	err = json.Unmarshal(res, &l)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func getReadme(u string) (string, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.raw+json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", nil
	}
	return string(body), nil
}
