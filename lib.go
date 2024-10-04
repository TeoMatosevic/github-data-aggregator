package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func createDir(path string, dir string) {
	if _, err := os.Stat(path + dir); os.IsNotExist(err) {
		os.Mkdir(path+dir, 0755)
	}
}

func scheduler(r *Repositories, urls *Urls) {
	go scheduleRepositories(r, urls)
	go scheduleUrls(r, urls)
}

func scheduleRepositories(r *Repositories, urls *Urls) {
	for {
		u, err := getRepositories(r)
		if err == nil {
			urls.lock()
			urls.addUrls(u)
			urls.unlock()
		}
		time.Sleep(12 * time.Hour)
	}
}

func scheduleUrls(r *Repositories, urls *Urls) {
	lc := make(chan *[]Url)
	rc := make(chan *[]Url)
	go addLanguage(r, lc)
	go addReadme(r, rc)
	for {
		urls.lock()
		u := urls.takeUrls(10)
		for _, v := range u {
			if v.Type == Language {
				lc <- &[]Url{v}
			}
			if v.Type == Readme {
				rc <- &[]Url{v}
			}
		}
		urls.unlock()
		time.Sleep(30 * time.Minute)
	}
}

func addLanguage(r *Repositories, c chan *[]Url) {
	for {
		u := <-c
		for _, v := range *u {
			r.lock()
			rc := r.readJson()
			languages, err := getLanguages(v.Url)
			if err != nil {
				r.unlock()
				continue
			}
			for i, vv := range rc {
				if vv.Id == v.Id {
					rc[i].Languages = languages
				}
			}
			r.writeJson(rc)
			r.unlock()
		}
	}
}

func addReadme(r *Repositories, c chan *[]Url) {
	for {
		u := <-c
		for _, v := range *u {
			r.lock()
			rc := r.readJson()
			readme, err := getReadme(v.Url)
			if err != nil {
				r.unlock()
				continue
			}
			for i, vv := range rc {
				if vv.Id == v.Id {
					rc[i].Readme = readme
				}
			}
			r.writeJson(rc)
			r.unlock()
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
			Readme_url:    readMe,
			Updated_at:    updated_at,
			Readme:        "",
		})
	}
	return r, nil
}

func removeRepository(r *Repositories, name string) {
	r.lock()
	rc := r.readJson()
	for i, v := range rc {
		if v.Name == name {
			rc = append((rc)[:i], (rc)[i+1:]...)
		}
	}
	r.writeJson(rc)
	r.unlock()
}

func getRepositories(r *Repositories) ([]Url, error) {
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

	r.lock()
	cr := r.readJson()

	for _, v := range repos {
		if !nameExists(cr, v.Name) {
			cr = append(cr, v)
			newUrls = append(newUrls, Url{Id: v.Id, Url: v.Languages_url, Type: Language})
			newUrls = append(newUrls, Url{Id: v.Id, Url: v.Readme_url, Type: Readme})
		} else {
			if olderThan(cr, v.Id, v.Updated_at) {
				setUpdatedAt(cr, v.Id, v.Updated_at)
				newUrls = append(newUrls, Url{Id: v.Id, Url: v.Languages_url, Type: Language})
				newUrls = append(newUrls, Url{Id: v.Id, Url: v.Readme_url, Type: Readme})
			}
			setRepository(cr, v.Id, v)
		}
	}

	for _, v := range cr {
		if !nameExists(repos, v.Name) {
			removeRepository(r, v.Name)
		}
	}

	r.writeJson(cr)
	r.unlock()

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
