package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func scheduler(r *Repositories, urls *Urls) {
	go scheduleRepositories(r, urls)
	go scheduleUrls(r, urls)
}

func scheduleRepositories(r *Repositories, urls *Urls) {
	for {
		u, err := getRepositories(r)
		if err == nil {
			for _, v := range u {
				*urls = append(*urls, v)
			}
		}
		time.Sleep(5 * time.Minute)
	}
}

func scheduleUrls(r *Repositories, urls *Urls) {
	lc := make(chan *Urls)
	rc := make(chan *Urls)
	go addLanguage(r, lc)
	go addReadme(r, rc)
	for {
		u := takeUrls(urls, 40)
		for _, v := range u {
			if v.Type == Language {
				lc <- &Urls{v}
			}
			if v.Type == Readme {
				rc <- &Urls{v}
			}
		}
		time.Sleep(2 * time.Minute)
	}
}

func addLanguage(r *Repositories, c chan *Urls) {
	for {
		u := <-c
		for _, v := range *u {
			languages, err := getLanguages(v.Url)
			if err != nil {
				continue
			}
			for i, vv := range *r {
				if vv.Id == v.Id {
					(*r)[i].Languages = languages
				}
			}
		}
	}
}

func addReadme(r *Repositories, c chan *Urls) {
	for {
		u := <-c
		for _, v := range *u {
			readme, err := getReadme(v.Url)
			if err != nil {
				continue
			}
			for i, vv := range *r {
				if vv.Id == v.Id {
					(*r)[i].Readme = readme
				}
			}
		}
	}
}

func unmarshalRepository(data []byte) (Repositories, error) {
	var r []Repository
	var ud []map[string]interface{}
	var err = json.Unmarshal(data, &ud)
	if err != nil {
		return nil, err
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
		r = append(r, Repository{
			Id:            v["id"].(float64),
			Name:          v["name"].(string),
			Full_name:     v["full_name"].(string),
			Description:   description.(string),
			languages_url: v["languages_url"].(string),
			readme_url:    fmt.Sprintf("%s/contents/README.md", v["url"]),
			updated_at:    updated_at,
		})
	}
	return r, nil
}

func removeRepository(r *Repositories, name string) {
	for i, v := range *r {
		if v.Name == name {
			*r = append((*r)[:i], (*r)[i+1:]...)
		}
	}
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

	for _, v := range repos {
		if !r.nameExists(v.Name) {
			*r = append(*r, v)
			newUrls = append(newUrls, Url{Id: v.Id, Url: v.languages_url, Type: Language})
			newUrls = append(newUrls, Url{Id: v.Id, Url: v.readme_url, Type: Readme})
		} else {
			if r.olderThan(v.Id, v.updated_at) {
				r.setUpdatedAt(v.Id, v.updated_at)
				newUrls = append(newUrls, Url{Id: v.Id, Url: v.languages_url, Type: Language})
				newUrls = append(newUrls, Url{Id: v.Id, Url: v.readme_url, Type: Readme})
			}
			(*r).setRepository(v.Id, v)
		}
	}

	for _, v := range *r {
		if !repos.nameExists(v.Name) {
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

func takeUrls(items *Urls, n int) Urls {
	if n > len(*items) {
		u := (*items)
		*items = (*items)[:0]
		return u
	}
	u := (*items)[:n]
	*items = (*items)[n:]
	return u
}
