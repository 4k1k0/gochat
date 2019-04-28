package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"../trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
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

type Credentials struct {
	id        string
	secretKey string
}

func SetUp() (Credentials, error) {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	conf := Credentials{}
	err := decoder.Decode(&conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

func main() {

	conf, err := SetUp()
	if err != nil {
		log.Fatal("Error loading config file.")
	}

	// Agrega flags para cli
	var addr = flag.String("addr", ":8080", "The addr of the application")
	flag.Parse()

	url := "http://localhost" + *addr + "/auth/callback/github"

	// Auth

	gomniauth.SetSecurityKey("af7hZXjbEmSVp7OyTs11M8Ij6ITyJjTrqPomRmM53KTpGbigGVjcZqUvISuDeeV0")
	gomniauth.WithProviders(
		github.New(conf.id, conf.secretKey, url),
	)

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
