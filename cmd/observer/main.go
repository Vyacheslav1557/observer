package main

import (
	"fmt"
	"github.com/Vyacheslav1557/observer"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/nats-io/nats.go"
	"log"
	"net/http"
)

func main() {
	var cfg observer.Config
	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		panic(fmt.Sprintf("error reading config: %s", err.Error()))
	}

	nc, err := nats.Connect(cfg.NatsUrl)
	if err != nil {
		log.Fatalf("Ошибка подключения к NATS: %v", err)
	}
	defer nc.Close()

	obs := observer.NewObserver(nc, cfg.JwtSecret)

	http.HandleFunc("/solutions", obs.ListSolutionsWS)
	log.Fatal(http.ListenAndServe(":13400", nil))
}
