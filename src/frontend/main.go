package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

const (
	port            = "8080"
	defaultCurrency = "USD"
	cookieMaxAge    = 60 * 60 * 48

	cookieSessionID = "session-id"
	cookieCurrency  = "currency"
)

var (
	whitelistedCurrencies = map[string]bool{
		"USD": true, "EUR": true, "CAD": true, "JPY": true}
)

type ctxKeySessionID struct{}

type frontendServer struct {
	productCatalogSvcAddr string
	productCatalogSvcConn *grpc.ClientConn

	currencySvcAddr string
	currencySvcConn *grpc.ClientConn

	cartSvcAddr string
	cartSvcConn *grpc.ClientConn
}

func main() {
	ctx := context.Background()

	srvPort := port
	if os.Getenv("PORT") != "" {
		srvPort = os.Getenv("PORT")
	}
	svc := new(frontendServer)
	mustMapEnv(&svc.productCatalogSvcAddr, "PRODUCT_CATALOG_SERVICE_ADDR")
	mustMapEnv(&svc.currencySvcAddr, "CURRENCY_SERVICE_ADDR")
	// mustMapEnv(&svc.cartSvcAddr, "CART_SERVICE_ADDR")

	var err error
	svc.currencySvcConn, err = grpc.DialContext(ctx, svc.currencySvcAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect currency service: %+v", err)
	}
	svc.productCatalogSvcConn, err = grpc.DialContext(ctx, svc.productCatalogSvcAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect productcatalog service: %+v", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", refreshCookies(
		ensureSessionID(
			svc.homeHandler))).
		Methods(http.MethodGet)
	r.HandleFunc("/logout", svc.logoutHandler).
		Methods(http.MethodGet)
	r.HandleFunc("/setCurrency", refreshCookies(
		ensureSessionID(
			svc.setCurrencyHandler))).
		Methods(http.MethodPost)
	log.Printf("starting server on :" + srvPort)
	log.Fatal(http.ListenAndServe("localhost:"+srvPort, r))
}

func mustMapEnv(target *string, envKey string) {
	v := os.Getenv(envKey)
	if v == "" {
		panic(fmt.Sprintf("environment variable %q not set", envKey))
	}
	*target = v
}