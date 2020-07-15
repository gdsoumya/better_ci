package ci

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gdsoumya/better_ci/types"
	"github.com/gdsoumya/better_ci/utils"
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
			allowed := false
			for _, value := range c.permission {
				if strings.ToLower(value) == strings.ToLower(commentData.Permission) || strings.ToLower(value) == "any" {
					allowed = true
					break
				}
			}
			if !allowed {
				w.Write([]byte("We don't do this here"))
				//log.Print("ERROR : UNAUTHORIZED USER DEPLOY")
				return
			}
			pat := regexp.MustCompile(`[\/#]`)
			urlData := pat.Split(commentData.URL, -1)
			if urlData[5] != "pull" {
				// log.Print(urlData[4], []byte(urlData[4]), []byte("pull"))
				w.Write([]byte("We don't do this here"))
				return
			}
			cmtData := strings.TrimSpace(commentData.Body)
			cmtBody := strings.Split(cmtData, " ")
			if len(cmtBody) == 1 && cmtBody[0] == "/preview" {
				commentData.Time = 5
			} else if len(cmtBody) == 2 && cmtBody[0] == "/preview" {
				commentData.Time, err = strconv.Atoi(cmtBody[1])
				if err != nil {
					w.Write([]byte("We don't do this here"))
					log.Print("ERROR : WRONG MESSAGE FORMAT")
					return
				}
			} else {
				w.Write([]byte("We don't do this here"))
				log.Print("ERROR : WRONG MESSAGE FORMAT")
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
