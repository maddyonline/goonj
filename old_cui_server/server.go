package main

// OLD CODE
import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/httputil"
)

const MAIN_HTML = "../static_cui/cui/templates/cui.html"

var cui_html []byte
var tasks map[string]*Task

type MessageGetTask struct {
	Task                 string
	Ticket               string
	ProgLang             string
	HumanLang            string
	PreferServerProgLang bool
}

type Task struct {
	XMLName          xml.Name `xml:"response"`
	Status           string   `xml:"task_status" json: "task_status"`
	Description      string   `xml:"task_description"`
	Type             string   `xml:"task_type"`
	SolutionTemplate string   `xml:"solution_template"`
	CurrentSolution  string   `xml:"current_solution"`
	ExampleInput     string   `xml:"example_input"`
	ProgLangList     string   `xml:"prg_lang_list"`
	HumanLangList    string   `xml:"human_lang_list"`
	ProgLang         string   `xml:"prg_lang"`
	HumanLang        string   `xml:"human_lang"`
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	switch r.URL.Path {
	case "/":
		w.Write(cui_html)
	case "/c/_start/":
		w.Write([]byte("something"))
	case "/chk/clock/":
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Write(getClock())
	case "/c/_get_task/":
		//map[prefer_server_prg_lang:[false] ticket:[TICKET_ID] task:[task1] human_lang:[en] prg_lang:[c]]
		val := &MessageGetTask{
			Task:                 r.FormValue("task"),
			Ticket:               r.FormValue("ticket"),
			ProgLang:             r.FormValue("prg_lang"),
			HumanLang:            r.FormValue("human_lang"),
			PreferServerProgLang: r.FormValue("prefer_server_prg_lang") == "false",
		}
		log.Println(r.Form)
		j, _ := json.Marshal(val)
		fmt.Printf("%s\n", string(j))
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Write(getTask(val))
	case "/chk/save/":
		val := struct {
			Task     string
			Ticket   string
			ProgLang string
			Solution string
		}{
			Task:     r.FormValue("task"),
			Ticket:   r.FormValue("ticket"),
			ProgLang: r.FormValue("prg_lang"),
			Solution: r.FormValue("solution"),
		}
		log.Println("In /chk/save:", r.Form)
		tasks[val.Task].CurrentSolution = val.Solution
		tasks[val.Task].ProgLang = val.ProgLang

	case "/chk/verify/":
		r.ParseForm()
		log.Println(r.Form)
		type Status struct {
			OK      int    `xml:"ok"`
			Message string `xml:"message"`
		}
		type MainStatus struct {
			Compile Status `xml:"compile"`
			Example Status `xml:"example"`
		}
		resp := struct {
			XMLName xml.Name   `xml:"response"`
			Result  string     `xml:"result"`
			Extra   MainStatus `xml:"extra"`
		}{
			Result: "OK",
			Extra: MainStatus{
				Compile: Status{1, "The solution compiled flawlessly."},
				Example: Status{1, "OK"},
			},
		}
		xmlResp, err := xml.MarshalIndent(resp, " ", "    ")
		if err != nil {
			log.Fatal(err)
			xmlResp = []byte{}
		}
		fmt.Println(xmlResp)
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Write(xmlResp)
	}

}

func main() {
	var err error
	cui_html, err = ioutil.ReadFile(MAIN_HTML)
	if err != nil {
		log.Fatal(err)
	}

	tasks = map[string]*Task{}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handleHttp))
	mux.Handle("/static/", http.FileServer(http.Dir("../static_cui/cui/")))

	log.Fatal(http.ListenAndServe(":8082", mux))
}
