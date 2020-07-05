package old

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/v32/github"
	"github.com/gorilla/mux"
)

type commentDetails struct {
	Permission string `json:"author_association"`
	Body       string `json:"body"`
	URL        string `json:"html_url"`
	ID         int64  `json:"id"`
}

// var revProxy map[string]string = make(map[string]string)
// var ports []string = []string{}
var lock sync.Mutex

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/payload", webhook)
	router.HandleFunc("/", hello)
	log.Print("Server Started at :8080")
	http.ListenAndServe(":8080", router)
	// router1 := mux.NewRouter()
	// router1.HandleFunc("/", hello)
	// http.ListenAndServe(":8000", router1)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("CI RUNNING"))
}

// func reverseProxy(res http.ResponseWriter, req *http.Request) {
// 	vars := mux.Vars(r)
// 	path := vars["path"]
// 	if value, ok := revProxy[path]; !ok {
// 		res.WriteHeader(http.StatusNotFound)
// 		res.Write([]byte("404 Not Found"))
// 		return
// 	}
// 	url, _ := url.Parse("http://localhost:8000/")
// 	proxy := httputil.NewSingleHostReverseProxy(url)

// 	// Update the headers to allow for SSL redirection
// 	req.URL.Host = url.Host
// 	req.URL.Scheme = url.Scheme
// 	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
// 	req.Host = url.Host

// 	// Note that ServeHttp is non blocking and uses a go routine under the hood
// 	proxy.ServeHTTP(res, req)
// }
func webhook(w http.ResponseWriter, r *http.Request) {

	data, _ := ioutil.ReadAll(r.Body)
	// log.Print(string(data))
	var jsonData map[string]*json.RawMessage
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		log.Print("\n\n!! ERRORR !!\n\n", err.Error())
		w.Write([]byte("Error"))
	}
	if _, ok := jsonData["zen"]; ok {
		log.Printf("NEW HOOK ADDED %s", string(*jsonData["hook_id"]))
		w.Write([]byte("HOOK REGISTERED"))
	} else if _, ok := jsonData["comment"]; ok {
		var commentData commentDetails
		err = json.Unmarshal(*jsonData["comment"], &commentData)
		if err != nil {
			log.Print("\n\n!! ERRORR !!\n\n", err.Error())
			w.Write([]byte("Error"))
		}
		var state string
		_ = json.Unmarshal(*jsonData["action"], &state)
		if state == "created" {
			log.Print(commentData)
			pat := regexp.MustCompile(`[\/#]`)
			urlData := pat.Split(commentData.URL, -1)
			log.Print(urlData)
			if urlData[5] != "pull" || commentData.Body != "/khulja-simsim" {
				// log.Print(urlData[4], []byte(urlData[4]), []byte("pull"))
				w.Write([]byte("We don't do this here"))
				return
			}
			username, repo, pr := urlData[3], urlData[4], urlData[6]
			deploy(username, repo, pr, commentData)
			// githubAction(commentData, username, repo)
			// log.Print("\n\nPR DETAILS : ", username, repo, pr, "\n\n")
		}
		w.Write([]byte("COMMENT EVENT HEARD"))
	} else {
		w.Write([]byte("We don't do this here"))
	}
}

func deploy(username, repo, pr string, cmt commentDetails) {
	dir := username + "_" + repo + "_pr" + pr
	_, err := os.Stat(dir)
	if !os.IsNotExist(err) {
		githubAction(cmt, username, repo, "** WAIT FOR PREVIOUS BUILD TO EXPIRE **")
		return
	}
	errDir := os.MkdirAll(dir, 0755)
	if errDir != nil {
		githubAction(cmt, username, repo, "** ERROR IN CI **")
		return
	}
	cmd := exec.Command("git", "clone", "https://github.com/"+username+"/"+repo+".git", ".")
	cmd.Dir = dir
	// run command
	if otuput, err := cmd.Output(); err != nil {
		log.Println("Error:", err)
	} else {
		log.Printf("Otuput: %s\n", otuput)
	}
	// cmd = exec.Command("pwd")
	// cmd.Dir = dir
	// if otuput, err := cmd.Output(); err != nil {
	// 	log.Println("Error:", err)
	// } else {
	// 	log.Printf("Otuput: %s\n", otuput)
	// }
	cmd = exec.Command("git", "fetch", "origin", "refs/pull/"+pr+"/head:pr_"+pr)
	cmd.Dir = dir
	// run command
	if otuput, err := cmd.Output(); err != nil {
		log.Println("Error:", err)
	} else {
		log.Printf("Otuput: %s\n", otuput)
	}
	cmd = exec.Command("git", "checkout", "pr_"+pr)
	cmd.Dir = dir
	// run command
	if otuput, err := cmd.Output(); err != nil {
		log.Println("Error:", err)
	} else {
		log.Printf("Otuput: %s\n", otuput)
	}
	urls := tempParser(dir + "/" + "docker-compose.yml")
	cmtData := ""
	for _, value := range urls {
		cmtData += "<base_url>:" + value + "<br>"
	}
	log.Print("STARTING DOCKER FOR : ", dir)
	cmd = exec.Command("docker-compose", "up", "-d")
	cmd.Dir = dir
	// run command
	if otuput, err := cmd.Output(); err != nil {
		log.Println("Error:", err)
	} else {
		log.Printf("Otuput: %s\n", otuput)
	}
	githubAction(cmt, username, repo, cmtData)
	time.Sleep(5 * time.Minute)
	log.Print("STOPPING DOCKER FOR : ", dir)
	// lock.Lock()
	// for _, value := range urls {
	// 	delete(revProxy, value)
	// }
	// lock.Unlock()
	cmd = exec.Command("docker-compose", "down", "-v") //ADD --rmi all
	cmd.Dir = dir
	// run command
	if otuput, err := cmd.Output(); err != nil {
		log.Println("Error:", err)
	} else {
		log.Printf("Otuput: %s\n", otuput)
	}
	err = os.RemoveAll(dir)
	if err != nil {
		log.Fatal(err)
	}
	cmtData += "LINK EXPIRED<br>"
	githubAction(cmt, username, repo, cmtData)
	// time.Sleep(2 * time.Minute)
	// log.Print(dir, " WOKE")
}
func githubAction(cmt commentDetails, username, repo, cmtData string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "token"},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	// log.Print(client.Organizations.List(ctx, "gdsoumya", nil))
	var data github.IssueComment
	cmt.Body += "\n" + cmtData
	data.Body = &cmt.Body
	// log.Print("\n\nHEREE\n\n")
	_, _, err := client.Issues.EditComment(ctx, username, repo, cmt.ID, &data)
	if err != nil {
		log.Print("\n\nERORR : ", err.Error())
	}
	log.Print("COMMENT EDITED")
}
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func tempParser(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	urls := []string{}
	lines := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "<PORT>") {
			// temp := genRandom(16)
			// lock.Lock()
			tempP, _ := getFreePort()
			temp := strconv.Itoa(tempP)
			urls = append(urls, temp)
			// revProxy[temp] = strconv.Itoa(tempP)
			// lock.Unlock()
			line = strings.Replace(line, "<PORT>", temp, -1)
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	err = printLines(path, lines)
	if err != nil {
		log.Fatal(err)
	}
	return urls
}

// func genRandom(length int) string {
// 	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
// 	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
// 	b := make([]byte, length)
// 	for i := range b {
// 		b[i] = charset[seededRand.Intn(len(charset))]
// 	}
// 	return string(b)
// }
func printLines(filePath string, values []string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, value := range values {
		fmt.Fprintln(f, value) // print values to f, one per line
	}
	return nil
}
