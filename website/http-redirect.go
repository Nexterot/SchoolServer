// http-redirect
package main

import (
	"net/http"
)

func redirecter(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://netschool.app/", 301)
}

func main() {
	http.HandleFunc("/", redirecter)
	http.ListenAndServe("0.0.0.0:80", nil)
}
