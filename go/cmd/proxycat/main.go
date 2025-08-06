package main

import (
	"html/template"
	"log"
	"net/http"

	"proxycat/proxy"
)

var (
	tpl     = template.Must(template.ParseFiles("web/index.html"))
	manager = proxy.NewManager(proxy.HTTPCheck)
)

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/retry", handleRetry)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/deleteAll", handleDeleteAll)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Active []proxy.Proxy
		Failed []proxy.Proxy
	}{
		Active: manager.Active(),
		Failed: manager.Failed(),
	}
	tpl.Execute(w, data)
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err == nil {
		addr := r.Form.Get("address")
		if addr != "" {
			manager.Add(addr)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleRetry(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err == nil {
		id := r.Form.Get("id")
		manager.Retry(id)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err == nil {
		id := r.Form.Get("id")
		manager.Delete(id)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleDeleteAll(w http.ResponseWriter, r *http.Request) {
	manager.DeleteAllFailed()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
