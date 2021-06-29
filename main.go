package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Urls   Urls
	Server Server
}

type Urls struct {
	ChackNoris string
	Insult     string
	BadJoke    string
}

type Server struct {
	Host string
	Port string
}

type JokeResponse struct {
	Joke string `json:"joke"`
	Code int64  `json:"code"`
}

type PasswordResponse struct {
	Password string `json:"password"`
	Code     uint16 `json:"code"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  uint16 `json:"code"`
}

func main() {
	//get config data
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	config := new(Config)
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println(err)
	}

	http.HandleFunc("/badJoke", func(w http.ResponseWriter, r *http.Request) {
		joke, err := getBadJoke(config.Urls.BadJoke)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			log.Println(err)
			json.NewEncoder(w).Encode(ErrorResponse{err.Error(), 400})
		} else {
			json.NewEncoder(w).Encode(JokeResponse{joke, 200})
		}
		fmt.Fprintf(w, "")
	})

	http.HandleFunc("/chackNorisJoke", func(w http.ResponseWriter, r *http.Request) {
		joke, err := getChackNorisJoke(config.Urls.ChackNoris)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			log.Println(err)
			json.NewEncoder(w).Encode(ErrorResponse{err.Error(), 400})
		} else {
			json.NewEncoder(w).Encode(JokeResponse{joke, 200})
		}
		fmt.Fprintf(w, "")
	})

	http.HandleFunc("/generatePass", func(w http.ResponseWriter, r *http.Request) {
		password_num, ok := r.URL.Query()["password_num"]

		if !ok || len(password_num[0]) < 1 {
			log.Println("Url Param 'password_num' is missing")
			password_num = []string{"12"}
		}
		if len(password_num[0]) > 3 {
			fmt.Fprintf(w, "Max value of 'password_num' = 32")
			return
		}
		num, err := strconv.Atoi(password_num[0])
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "Не корректный тип введите число от 1 до 999")
		}
		password := PasswordGenerator(num)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PasswordResponse{password, 200})
		fmt.Fprintf(w, "")
	})

	http.HandleFunc("/EvilInsult", func(w http.ResponseWriter, r *http.Request) {
		lang, ok := r.URL.Query()["lang"]
		if !ok {
			lang = []string{"en"}
		}
		joke, err := getEvilInsult(strings.ToLower(lang[0]), config.Urls.Insult)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			log.Println(err)
			json.NewEncoder(w).Encode(ErrorResponse{err.Error(), 400})
		} else {
			json.NewEncoder(w).Encode(JokeResponse{joke, 200})
		}

		fmt.Fprintf(w, "")
	})

	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port),
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(server.ListenAndServe())

}

func PasswordGenerator(n int) string {
	var letterRunes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var numberRunes = "1234567890"
	var symbolRunes = "!@#$%^&*()"
	halh := n / 3
	halhLetter := halh
	halhnumber := halh * 2
	halhsymbol := halh * 3

	b := make([]string, n)
	for i := range b {
		if i >= 0 && i < halhLetter {
			b[i] = string(letterRunes[rand.Intn(len(letterRunes))])
		} else if i >= halhLetter && i < halhnumber {
			b[i] = string(numberRunes[rand.Intn(len(numberRunes))])

		} else if i >= halhnumber && i < halhsymbol {
			b[i] = string(symbolRunes[rand.Intn(len(symbolRunes))])

		}
	}
	result := strings.Join(b, "")
	return result
}

func doGETRequest(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("ERR", err)
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("ERR", err)
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("ERRRROR", err)
		return nil, err
	}
	return body, nil
}

type BadJoke struct {
	Id     string
	Joke   string
	Status int64
}

func getBadJoke(url string) (string, error) {
	body, err := doGETRequest(url)
	if err != nil {
		return "", err
	}

	var message BadJoke
	err = json.Unmarshal(body, &message)
	if err != nil {
		return "", err
	}
	if message.Status != 200 {
		return "", errors.New(fmt.Sprintln("Status != 200: body = ", message))
	}
	return message.Joke, nil
}

type ChackNoris struct {
	Type  string
	Value ChackNorisJoke
}
type ChackNorisJoke struct {
	Id   int64
	Joke string
}

func getChackNorisJoke(url string) (string, error) {
	body, err := doGETRequest(url)
	if err != nil {
		return "", err
	}
	var message ChackNoris
	err = json.Unmarshal(body, &message)
	if err != nil {
		return "", err
	}
	if message.Type != "success" {
		return "", errors.New(fmt.Sprintln(message))
	}
	return message.Value.Joke, nil
}

type EvilInsult struct {
	Number   string
	Language string
	Insult   string
	Created  string
	Shown    string
	Createby string
	Active   string
	Comment  string
}

func getEvilInsult(lang string, url string) (string, error) {
	languages := [2]string{"ru", "en"}
	lang_exists := false
	for _, language := range languages {
		if language == lang {
			lang_exists = true
		}
	}
	if lang_exists == false {
		return "", errors.New(fmt.Sprintf("'%s' Language is not supported. Use %s", lang, languages))
	}
	body, err := doGETRequest(fmt.Sprintf("%s?lang=%s&type=json", url, lang))
	if err != nil {
		return "", err
	}
	var message EvilInsult
	err = json.Unmarshal(body, &message)
	if err != nil {
		return "", err
	}

	return message.Insult, nil
}
