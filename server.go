package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var CONNECTED_NODE_ADDRESS = "http://127.0.0.1:8000"
var posts []map[string]interface{}

func fetchPosts() {
	getChainAddress := fmt.Sprintf("%s/chain", CONNECTED_NODE_ADDRESS)
	response, err := http.Get(getChainAddress)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		var chain map[string]interface{}
		content := []map[string]interface{}{}
		err = json.NewDecoder(response.Body).Decode(&chain)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, block := range chain["chain"].([]interface{}) {
			for _, tx := range block.(map[string]interface{})["transactions"].([]interface{}) {
				txMap := tx.(map[string]interface{})
				txMap["index"] = block.(map[string]interface{})["index"]
				txMap["hash"] = block.(map[string]interface{})["previous_hash"]
				content = append(content, txMap)
			}
		}

		posts = content
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fetchPosts()
	renderTemplate(w, "index.html", map[string]interface{}{
		"title":          "YourNet: Decentralized content sharing",
		"posts":          posts,
		"nodeAddress":    CONNECTED_NODE_ADDRESS,
		"readableTime":   timestampToString,
	})
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	postContent := r.FormValue("content")
	author := r.FormValue("author")

	postObject := map[string]interface{}{
		"author":  author,
		"content": postContent,
	}

	newTxAddress := fmt.Sprintf("%s/new_transaction", CONNECTED_NODE_ADDRESS)

	reqBody, err := json.Marshal(postObject)
	if err != nil {
		fmt.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, newTxAddress, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func timestampToString(epochTime int64) string {
	return time.Unix(epochTime, 0).Format("15:04")
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmpl = fmt.Sprintf("templates/%s", tmpl)
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/submit", submitHandler)
	http.ListenAndServe(":8080", nil)
}
