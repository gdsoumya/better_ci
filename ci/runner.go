package ci

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/gdsoumya/better_ci/utils"

	"github.com/gdsoumya/better_ci/types"
)

func (c *Config) Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("CI RUNNING"))
}
func (c *Config) WebHook(w http.ResponseWriter, r *http.Request) {
	secret := r.Header.Get("X-Hub-Signature")
	data, _ := ioutil.ReadAll(r.Body)
	var jsonData map[string]*json.RawMessage
	err := json.Unmarshal(data, &jsonData)
	if !utils.VerifySig(secret, c.websec, data) {
		w.Write([]byte("OOPS Wrong Hook!"))
		log.Print("RECEIVED EVENT FROM WRONG HOOK, SHA1 SIG MISMATCH")
		return
	}
	if err != nil {
		log.Print("ERROR Handling Event : ", err.Error())
		w.Write([]byte("Error : " + err.Error()))
		return
	}
	if _, ok := jsonData["zen"]; ok {
		log.Printf("NEW HOOK ADDED %s", string(*jsonData["hook_id"]))
		w.Write([]byte("HOOK REGISTERED"))
		return
	} else if _, ok := jsonData["comment"]; ok {
		var commentData types.EventDetails
		err = json.Unmarshal(*jsonData["comment"], &commentData)
		if err != nil {
			log.Print("ERROR Handling Event : ", err.Error())
			w.Write([]byte("Error : " + err.Error()))
			return
		}
		var state string
		err = json.Unmarshal(*jsonData["action"], &state)
		if err != nil {
			log.Print("ERROR Handling Event : ", err.Error())
			w.Write([]byte("Error : " + err.Error()))
			return
		}
		if state == "created" {
			log.Print(commentData)
			pat := regexp.MustCompile(`[\/#]`)
			urlData := pat.Split(commentData.URL, -1)
			//log.Print(urlData[0], urlData[1], urlData[2], urlData[3], urlData[4], urlData[5], urlData[6])
			if urlData[5] != "pull" || commentData.Body != "/preview" {
				// log.Print(urlData[4], []byte(urlData[4]), []byte("pull"))
				w.Write([]byte("We don't do this here"))
				return
			}
			commentData.Username, commentData.Repo, commentData.PR = urlData[3], urlData[4], urlData[6]
			go c.Deploy(&commentData)
			w.Write([]byte("Deployment Initiated"))
		} else {
			w.Write([]byte("We don't do this here"))
		}
	} else {
		w.Write([]byte("We don't do this here"))
	}
}
