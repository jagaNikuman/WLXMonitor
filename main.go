package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/yaml.v2"
)

type Config struct {
	DBUrl  string `yaml:"dburl"`
	DBName string `yaml:"dbname"`

	Username string `yaml:"username"`
	Password string `yaml:"password"`
	WLX313   WLX313 `yaml:"wlx313"`
}
type WLX313 struct {
	TargetURL string   `yaml:"targetURL"`
	UserAgent string   `yaml:"userAgent"`
	Selectors Selector `yaml:"selectors"`
}
type Selector struct {
	Client2G string `yaml:"2G"`
	Client5G string `yaml:"5G"`
	Temp     string `yaml:"temp"`
	Mem      string `yaml:"mem"`
	CPU      string `yaml:"cpu"`
}
type Values struct {
	Clients2G int
	Clients5G int
	Temp      int
	Mem       int
	CPU       int
}

func loadConfig(filename string, config *Config) {
	buffer, _ := ioutil.ReadFile(filename)
	yaml.Unmarshal(buffer, &config)
}

func getHTML(config *Config) *http.Response {
	cookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: cookieJar,
	}
	req, _ := http.NewRequest("GET", config.WLX313.TargetURL, nil)
	req.Header.Set("User-Agent", config.WLX313.UserAgent)
	req.SetBasicAuth(config.Username, config.Password)
	res, err := httpClient.Do(req)
	if err != nil {
		println(err)
	}

	return res
}

func insertData(config *Config, device string, tag string, value int) {
	url := fmt.Sprintf("%swrite?db=%s", config.DBUrl, config.DBName)
	sendMessage := fmt.Sprintf("%s,%s value=%d", device, tag, value)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(sendMessage)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
	defer resp.Body.Close()
}

func insertAllData(config *Config, device string, values *Values) {
	//value
	v := reflect.Indirect(reflect.ValueOf(values))
	//type
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		//value of i
		f := v.Field(i)
		value := f.Interface()
		insertData(config, device, "type="+strings.ToLower(t.Field(i).Name), value.(int))
	}
}

func getValues(selectors *Selector, doc *goquery.Document, values *Values) {
	s := doc.Find(selectors.Client2G)
	values.Clients2G, _ = strconv.Atoi(strings.Trim(strings.TrimSpace(s.Text()), "台"))

	s = doc.Find(selectors.Client5G)
	values.Clients5G, _ = strconv.Atoi(strings.Trim(strings.TrimSpace(s.Text()), "台"))

	s = doc.Find(selectors.CPU)
	values.CPU, _ = strconv.Atoi(strings.Trim(strings.TrimSpace(s.Text()), "%"))

	s = doc.Find(selectors.Mem)
	values.Mem, _ = strconv.Atoi(strings.Trim(strings.TrimSpace(s.Text()), "%"))

	s = doc.Find(selectors.Temp)
	values.Temp, _ = strconv.Atoi(strings.Trim(strings.TrimSpace(s.Text()), "℃"))
}

func getRTX313Values(selectors *Selector, doc *goquery.Document, values *Values) {
	s := doc.Find(selectors.Client2G)
	fmt.Println(s.Text())
	values.Clients2G, _ = strconv.Atoi(strings.Trim(strings.TrimSpace(s.Text()), "台"))
}

func main() {
	var config Config
	loadConfig("./config.yaml", &config)
	var values Values

	var config2 Config
	loadConfig("./config2.yaml", &config2)
	var values2 Values

	for {

		res := getHTML(&config)
		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			println(err)
		}
		getValues(&config.WLX313.Selectors, doc, &values)
		fmt.Println(values)
		insertAllData(&config, "wlx313_1", &values)
		fmt.Println("WLX313_1 Inserted Datas.")

		res = getHTML(&config2)
		doc, _ = goquery.NewDocumentFromResponse(res)
		getValues(&config2.WLX313.Selectors, doc, &values2)
		fmt.Println(values2)
		insertAllData(&config2, "wlx313_2", &values2)
		fmt.Println("WLX313_1 Inserted Datas. Waiting 3 seconds....")
		time.Sleep(3 * time.Second)

	}
}
