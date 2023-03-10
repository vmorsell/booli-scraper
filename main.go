package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString() string {
	b := make([]byte, rand.Intn(10)*10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Apartment is the model for an apartment object.
type Apartment struct {
	ID             int
	Address        string
	Floor          float64
	Area           int
	Rooms          float64
	Price          int
	EstimatedValue int
	Fee            int
	ImageURLs      []string
}

func main() {
	f, err := os.Open("urls.txt")
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer f.Close()

	storage := NewFileStorage("./storage")

	s := bufio.NewScanner(f)
	for s.Scan() {
		adURL := s.Text()

		apt := Apartment{}

		c := colly.NewCollector(
			colly.AllowedDomains("www.booli.se", "bcdn.se"),
		)

		// ID.
		id, err := parseID(adURL)
		if err != nil {
			log.Fatalf("parse id: %v", err)
		}
		apt.ID = id

		// Address.
		c.OnHTML("h1.lzFZY._10w08", func(e *colly.HTMLElement) {
			apt.Address = strings.ReplaceAll(e.Text, "\u00ad", "")
		})

		// Price.
		c.OnHTML("h2.lzFZY._10w08", func(e *colly.HTMLElement) {
			if price, err := parsePrice(e.Text); err != nil {
				log.Printf("parse price: %v", err)
			} else {
				apt.Price = price
			}
		})

		// Area and rooms.
		c.OnHTML("div._2epd7._3XAuT._10w08 div._36W0F h4._1544W._10w08", func(e *colly.HTMLElement) {
			if area, err := parseArea(e.Text); err != nil {
				log.Printf("parse area: %v", err)
			} else {
				apt.Area = area
			}

			if rooms, err := parseRooms(e.Text); err != nil {
				log.Printf("parse rooms: %v", err)
			} else {
				apt.Rooms = rooms
			}
		})

		// Booli's estimated value.
		c.OnHTML("h2._1g-8A", func(e *colly.HTMLElement) {
			if v, err := parsePrice(e.Text); err != nil {
				log.Printf("parse estimated value: %v", err)
			} else {
				apt.EstimatedValue = v
			}
		})

		// Monthly fee and floor.
		c.OnHTML("div.DfWRI._1Pdm1._2zXIc.sVQc-", func(e *colly.HTMLElement) {
			switch e.ChildText("div._2soQI") {
			case "Avgift":
				if v, err := parsePrice(e.ChildText("div._18w8g")); err != nil {
					log.Printf("parse monthly fee: %v", err)
				} else {
					apt.Fee = v
				}
			case "V??ning":
				if v, err := parseFloor(e.ChildText("div._18w8g")); err != nil {
					log.Printf("parse floor: %v", err)
				} else {
					apt.Floor = v
				}
			}

		})

		// Get Images from the Apollo state.
		c.OnHTML("script", func(e *colly.HTMLElement) {
			if strings.HasPrefix(e.Text, "window.__APOLLO_STATE__") {
				re := regexp.MustCompile(`\"Image:([0-9]+)\":\{.*?\"width\":([0-9]+),\"height\":([0-9]+).*?\}`)
				matches := re.FindAllStringSubmatch(e.Text, -1)
				for _, m := range matches {
					id := m[1]
					width := m[2]
					height := m[3]
					url := fmt.Sprintf("https://bcdn.se/images/cache/%s_%sx%s.jpg", id, width, height)
					apt.ImageURLs = append(apt.ImageURLs, url)
				}
			}
		})

		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("User-Agent", randomString())
			fmt.Printf("Scraping %s...\n", r.URL.String())
		})

		c.OnScraped(func(_ *colly.Response) {
			if err := storage.Put(apt); err != nil {
				log.Fatalf("\nput: %v", err)
			}
			fmt.Println("Done.")
		})

		c.OnError(func(_ *colly.Response, err error) {
			log.Printf("\nError: %s\n", err)
		})

		if err := c.Visit(adURL); err != nil {
			log.Printf("visit: %s\n", err)
		}

	}
}

// parseID extracts the ad ID from the URL.
func parseID(url string) (int, error) {
	strID := url[strings.LastIndex(url, "/")+1:]
	id, err := strconv.Atoi(strID)
	if err != nil {
		return 0, fmt.Errorf("atoi: %w", err)
	}
	return id, nil
}

// parseArea extracts the area expressed as an integer.
func parseArea(s string) (int, error) {
	re := regexp.MustCompile(`([0-9]+) m??`)
	res := re.FindStringSubmatch(s)
	if len(res) == 0 {
		return 0, fmt.Errorf("can't find area information in input")
	}

	area, err := strconv.Atoi(res[1])
	if err != nil {
		return 0, fmt.Errorf("atoi: %w", err)
	}

	return area, nil
}

// parseRooms extracts the number of rooms expressed as a float.
func parseRooms(s string) (float64, error) {
	re := regexp.MustCompile(`([0-9]+)(???) rum`)
	res := re.FindStringSubmatch(s)
	if len(res) == 0 {
		return 0.0, fmt.Errorf("can't find number of rooms in input")
	}

	rooms, err := strconv.ParseFloat(res[1], 64)
	if err != nil {
		return 0.0, fmt.Errorf("parse float: %w", err)
	}

	if res[2] != "" {
		rooms += 0.5
	}

	return rooms, nil
}

// parseFloor extracts the floor expressed as a float.
func parseFloor(s string) (float64, error) {
	re := regexp.MustCompile(`([0-9]+)(???) tr`)
	res := re.FindStringSubmatch(s)
	if len(res) == 0 {
		return 0.0, fmt.Errorf("can't find floor in input")
	}

	rooms, err := strconv.ParseFloat(res[1], 64)
	if err != nil {
		return 0.0, fmt.Errorf("parse float: %w", err)
	}

	if res[2] != "" {
		rooms += 0.5
	}

	return rooms, nil
}

// parsePrice extracts the price from a string like '4 000 000 kr' expressed as an integer.
func parsePrice(s string) (int, error) {
	re := regexp.MustCompile(`([0-9][0-9 ]+) kr`)
	res := re.FindStringSubmatch(s)
	if len(res) == 0 {
		return 0, fmt.Errorf("can't find price in input")
	}

	v := res[1]
	v = strings.ReplaceAll(v, " ", "")
	vv, err := strconv.ParseInt(v, 0, 0)
	if err != nil {
		return 0, fmt.Errorf("parse int: %w", err)
	}

	return int(vv), nil
}
