package main

import (
	"time"
)

type Repository struct {
	Id            float64 `json:"id"`
	Name          string  `json:"name"`
	Full_name     string  `json:"full_name"`
	languages_url string
	Languages     map[string]interface{} `json:"languages"`
	Description   string                 `json:"description"`
	readme_url    string
	Readme        string `json:"readme"`
	updated_at    time.Time
}

type Repositories []Repository

type Url struct {
	Id   float64 `json:"id"`
	Url  string  `json:"url"`
	Type int     `json:"type"`
}

type Urls []Url

func (r Repositories) nameExists(name string) bool {
	for _, v := range r {
		if v.Name == name {
			return true
		}
	}
	return false
}

func (u Repositories) olderThan(id float64, t time.Time) bool {
	for _, v := range u {
		if v.Id == id && v.updated_at.Before(t) {
			return true
		}
	}
	return false
}

func (u Repositories) setUpdatedAt(id float64, t time.Time) {
	for i, v := range u {
		if v.Id == id {
			u[i].updated_at = t
		}
	}
}

func (u Repositories) setRepository(id float64, r Repository) {
	for i, v := range u {
		if v.Id == id {
			u[i].Name = r.Name
			u[i].Full_name = r.Full_name
			u[i].Description = r.Description
		}
	}
}

func (u Urls) languageExists(id float64) bool {
	for _, v := range u {
		if v.Id == id && v.Type == Language {
			return true
		}
	}
	return false
}

func (u Urls) readmeExists(id float64) bool {
	for _, v := range u {
		if v.Id == id && v.Type == Readme {
			return true
		}
	}
	return false
}
