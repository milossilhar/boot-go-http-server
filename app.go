package main

import "net/http"

func (ac *apiConfig) registerAPP() http.Handler { 
	return ac.middlewareIncreaseHits(http.FileServer(http.Dir("./")))
}