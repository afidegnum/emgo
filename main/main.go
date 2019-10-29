package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/afidegnum/emgo/smodels"
	"github.com/alehano/reverse"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"github.com/thedevsaddam/renderer"
	"github.com/xo/dburl"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

var err error

const (
	hostName string = "localhost:27017"
	dbName   string = "demo_todo"
	port     string = ":9000"
)

var root, _ = os.Getwd()

var rnd *renderer.Render

type Paths struct {
	Indx, Pages, Signin, Signon string
	Res                         []*smodels.Page
}

var froutes Paths

//var froutes = Paths{indx: reverse.Rev("index"),
//	pages: reverse.Rev("pages"), signin: reverse.Rev("index"),
//	signon: reverse.Rev("index"), res: []*smodels.Page{}}

//index
//pages
//register
//login
//activate
//recover
//2fa --2BA

//var flagVerbose = flag.Bool("v", false, "verbose")

//var FlagURL = flag.String("url", "postgres://postgres:@127.0.0.1/sweb", "url") // Page represents a row from 'sweb.pages'.

func init() {
	rnd = renderer.New(
		renderer.Options{
			ParseGlobPattern: "res/html/*.html",
		},
	)
}

func FetchPages(db smodels.XODB) ([]*smodels.Page, error) {
	const sqlstr = `SELECT ` +
		`tag, body, slug, title, id, FROM public.pages`

	q, err := db.Query(sqlstr)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	var res []*smodels.Page
	for q.Next() {
		p := smodels.Page{}

		// scan
		err = q.Scan(&p.Body, &p.Slug, &p.Title, &p.ID)
		if err != nil {
			return nil, err
		}

		res = append(res, &p)

	}

	return res, nil
}

//
//walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
//	route = strings.Replace(route, "/*/", "/", -1)
//	fmt.Printf("%s \n", route)
//	return nil
//}
//
//func Walker(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
//	route = strings.Replace(route, "/*/", "/", -1)
//	fmt.Printf("%s \n", route)
//	return nil
//}
//

//func RenderLinks(w io.Writer, r *chi.Router) error {
//	return chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
//		route = strings.Replace(route, "/*/", "/", -1)
//		_, err := fmt.Fprintf(w, `<a href="%s">%s</a>`, route)
//		return err
//	})
//}

func MainHandler(w http.ResponseWriter, r *http.Request) {

	//DB OPERATIONS START

	flag.Parse()
	//if *smodels.FlagVerbose {
	//	smodels.XOLog = func(s string, p ...interface{}) {
	//		fmt.Printf("-------------------------------------\nQUERY: %s\n  VAL: %v\n", s, p)
	//	}
	//}

	// open database
	db, err := dburl.Open(*smodels.FlagURL)
	if err != nil {
		log.Fatal(err)
	}

	res, err := smodels.FetchAllPage(db)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(reverse.Rev("index"))

	err = rnd.HTML(w, http.StatusOK, "indexPage", res)
	if err != nil {
		log.Fatal(err)
	}
}

func FrontHandler(w http.ResponseWriter, r *http.Request) {

	//DB OPERATIONS START

	//flag.Parse()
	//if *smodels.FlagVerbose {
	//	smodels.XOLog = func(s string, p ...interface{}) {
	//		fmt.Printf("-------------------------------------\nQUERY: %s\n  VAL: %v\n", s, p)
	//	}
	//}

	// open database
	db, err := dburl.Open(*smodels.FlagURL)
	if err != nil {
		log.Fatal(err)
	}

	res, err := FetchPages(db)

	if err != nil {
		log.Fatal(err)
	}

	froutes.Res = append(froutes.Res, res...)

	rnd.HTML(w, http.StatusOK, "indexPage", froutes)

}

// FileServer conveniently sets up a http.FileServer handler to serv hence;e
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err) //respond with error page or
	}
}

func main() {

	froutes = Paths{Indx: reverse.Rev("index"),
		Pages: reverse.Rev("pages"), Signin: reverse.Rev("index"),
		Signon: reverse.Rev("index"), Res: []*smodels.Page{}}

	cor := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cor.Handler)

	//r.Get(reverse.Add("index", "/"), MainHandler)
	//r.Get(reverse.Add("pages", "/pages"), smodels.PageHandler)
	//r.Get(reverse.Add("getpages", "/getpg"), smodels.GetPage)
	r.Get("/", MainHandler)
	r.Get("/pages", smodels.PageHandler)
	r.Post("/jpages", smodels.NewPage)
	r.Mount("/page", smodels.PageRt())
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "res")
	FileServer(r, "/", http.Dir(filesDir))

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Listening on port ", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	log.Println("Server gracefully stopped!")

}
