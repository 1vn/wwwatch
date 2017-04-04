package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/0xAX/notificator"
	"github.com/PuerkitoBio/goquery"
)

type Config struct {
	WatchList []string `json:"watchList"`
}

func gracefulExit(err error) {
	log.Println(err)
	os.Exit(0)
}

var notify *notificator.Notificator

func main() {
	rand.Seed(42)

	var interval int
	flag.IntVar(&interval, "interval", 60, "number of seconds between each poll")
	log.Println("polling interval: ", interval)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	cache := make(map[string][]string)

	configFile, err := os.Open(fmt.Sprintf("%s/%s", dir, "config.json"))
	if err != nil {
		gracefulExit(err)
	}

	var conf Config

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&conf); err != nil {
		gracefulExit(err)
	}

	for {
		log.Println("Wwwatching...")
		for _, site := range conf.WatchList {
			doc, err := goquery.NewDocument(site)
			if err != nil {
				gracefulExit(err)
			}

			curImgs := make([]string, 0)

			doc.Find("img").Each(func(i int, s *goquery.Selection) {
				src, ok := s.Attr("src")
				if ok {
					curImgs = append(curImgs, path.Base(src))
				}
			})
			if err != nil {
				gracefulExit(err)
			}

			changed := false
			if cur, ok := cache[site]; ok {
				if len(cache[site]) != len(curImgs) {
					changed = true
				} else {
					for i, img := range curImgs {
						if img != cur[i] {
							changed = true
						}
					}
				}
			}

			if changed {
				notify = notificator.New(notificator.Options{
					DefaultIcon: "icon/default.png",
					AppName:     "wwwatch",
				})
				notify.Push("Site Changed", site, "/home/user/icon.png", notificator.UR_CRITICAL)
				log.Println("[CHANGED]", site)
			} else {
				log.Println("[NO CHANGE]", site)
			}

			cache[site] = curImgs
		}

		sleepFor := time.Second * time.Duration(interval+rand.Intn(500))
		log.Printf("waiting %v before next poll", sleepFor)
		time.Sleep(sleepFor)
	}
}
