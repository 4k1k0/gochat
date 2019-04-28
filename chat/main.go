package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"../trace"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {

	// Agrega flags para cli
	var addr = flag.String("addr", ":8080", "The addr of the application")
	flag.Parse()

	// El room r son los websockets,
	// por lo que no se carga un template html en esa ruta
	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	// Para poder utilizar assets en los templates
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("templates/assets/"))))

	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	// El room se ejecuta en una gorutina
	// por lo que el server se ejecuta en la
	// gorutina principal

	go r.run()

	log.Println("Starting the web server on", *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("Listen and serve", err)
	}
}
