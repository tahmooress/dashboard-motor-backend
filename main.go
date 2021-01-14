package main

import (
	"database/sql"
	"log"
	"net/http"

	"dashboard.motor/customedb"
	h "dashboard.motor/handlers"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	r := mux.NewRouter()
	method := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
	origin := handlers.AllowedOrigins([]string{"*"})
	headers := handlers.AllowedHeaders([]string{"X-Request-With", "Content-Type", "Authorization"})
	db, err := sql.Open("postgres", customedb.SetDbInfo())
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	env := &customedb.Env{DB: db}
	r.Handle("/login", h.LoginHandler(env)).Methods("POST")
	r.Handle("/create-user", h.CreateUser(env)).Methods("POST")
	r.Handle("/buy-factor", h.AuthMiddleWare(h.HandleBuy(env))).Methods("POST")
	r.Handle("/sell-factor", h.AuthMiddleWare(h.HandleSell(env))).Methods("POST")
	// r.Handle("/entry-list", h.AuthMiddleWare(h.HandleList(env))).Methods("POST")
	r.Handle("/stock-lookup", h.AuthMiddleWare(h.StockHandle(env))).Methods("POST")
	r.Handle("/sales-history", h.AuthMiddleWare(h.HandleSaleHistory(env))).Methods("POST")
	r.Handle("/update-recive", h.AuthMiddleWare(h.UpdateReceive(env))).Methods("PUT")
	r.Handle("/update-payable", h.AuthMiddleWare(h.UpdatePayable(env))).Methods("PUT")
	// r.Handle("/swap-inventory", h.AuthMiddleWare(h.HandleSwap(env))).Methods("PUT")
	r.Handle("/unrec-list", h.AuthMiddleWare(h.HandleUnpayedRec(env))).Methods("POST")
	r.Handle("/unpay-list", h.AuthMiddleWare(h.HandleUnpayedPay(env))).Methods("POST")
	r.Handle("/partly-rec", h.AuthMiddleWare(h.PartlyUpdateReceives(env))).Methods("POST")
	r.Handle("/partly-pay", h.AuthMiddleWare(h.PartlyUpdatePays(env))).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(headers, method, origin)(r)))
}
