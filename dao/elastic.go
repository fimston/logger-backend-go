package dao

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type Elastic struct {
	conn string
}

type RequestType string

const (
	RequestPost RequestType = "POST"
	RequestPut  RequestType = "PUT"
)

func NewElastic(conn string) *Elastic {
	return &Elastic{conn: conn}
}

func (self *Elastic) CurrentIndex() string {
	return "index1"
}

func (self *Elastic) request(method RequestType, cmd string, data []byte) (string, error) {
	client := &http.Client{}
	r, _ := http.NewRequest(string(method), self.url(cmd), bytes.NewBuffer(data))
	r.Header.Add("Content-Type", "application/text")
	r.Header.Add("Content-Length", strconv.Itoa(len(data)))

	resp, err := client.Do(r)

	if err != nil {
		log.Println(err)
		return "", err
	}
	res, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err)
		return "", err
	}
	resp.Body.Close()
	return string(res), nil
}

func (self *Elastic) url(postfix string) string {
	return fmt.Sprintf("%s/%s", self.conn, postfix)
}

func (self *Elastic) CreateAccount(acc *AccountInfo) error {
	json := fmt.Sprintf(`
		{
			"actions":	[
				{
					"add": {
						"index":"%s",
						"alias":"%s",
						"filter": {
							"term":{
								"__user_id": %d
							}
						}
					}
				}
			]
		}`,
		self.CurrentIndex(),
		acc.IndexAlias(),
		acc.userId,
	)
	result, err := self.request(RequestPost, "_aliases", []byte(json))
	log.Println(result)
	return err
}
