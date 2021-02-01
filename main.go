package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// Author struct
type Author struct {
	Name string `xml:"name"`
	URI  string `xml:"uri"`
}

// Link struct
type Link struct {
	HREF string `xml:"href,attr"`
}

// Entry struct
type Entry struct {
	ID        string `xml:"id"`
	VideoID   string `xml:"videoId"`
	ChannelID string `xml:"channelId"`
	Title     string `xml:"title"`
	Link      Link   `xml:"link"`
	Author    Author `xml:"author"`
	Published string `xml:"published"`
	Updated   string `xml:"Updated"`
}

// Feed struct
type Feed struct {
	Updated string `xml:"updated"`
	Title   string `xml:"title"`
	Entry   Entry  `xml:"entry"`
}

// Image struct
type Image struct {
	URL string `json:"url"`
}

// Embed struct
type Embed struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Image Image  `json:"image"`
}

// Message struct
type Message struct {
	Content string `json:"content"`
	Embed   Embed  `json:"embed"`
}

func main() {
	serve()
}

// Generates Message from webhook data
func generateJSON(feed *Feed) (v Message) {
	v = Message{
		Content: fmt.Sprintf("%s Just Posted a Video!", feed.Entry.Author.Name),
		Embed: Embed{
			Title: feed.Entry.Title,
			URL:   feed.Entry.Link.HREF,
			Image: Image{
				URL: os.Getenv("IMAGE_URL"),
			},
		},
	}

	// fmt.Println(v)
	return
}

// Reponse handler for Youtube webhook
func youtubeHandle(w http.ResponseWriter, req *http.Request) {

	// Parse XML from webhook
	xmlBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	v := Feed{}

	err = xml.Unmarshal(xmlBody, &v)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// fmt.Printf("%s", xmlBody)

	msg := generateJSON(&v)

	jsonBody, jsonErr := json.Marshal(msg)
	if jsonErr != nil {
		fmt.Println(jsonErr.Error())
	}

	// fmt.Printf("%s", jsonBody)
	url := fmt.Sprintf("https://discord.com/api/v8/channels/%s/messages", os.Getenv("CHANNEL_ID"))
	botToken := "Bot " + os.Getenv("TOKEN")
	req, clientErr := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if clientErr != nil {
		log.Println(err)
		return
	}

	req.Header.Add("Authorization", botToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, postErr := client.Do(req) // Send JSON to Discord API
	if postErr != nil {
		log.Println(postErr)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s", body)
}

// Setup and start the service
func serve() {

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("$PORT must be set")
		return
	}

	handle := func(w http.ResponseWriter, req *http.Request) {

		if req.Method == "GET" { // handle challenge request

			params := req.URL.Query()
			challenge := params["hub.challenge"]

			input := strings.Join(challenge, "")
			fmt.Printf("Input: %s\n", input)

			io.WriteString(w, input)

		} else if req.Method == "POST" { // handle YouTube post update
			youtubeHandle(w, req)
		}

	}

	http.HandleFunc("/", handle)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
