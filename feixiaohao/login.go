package feixiaohao

import (
	"fmt"
	"time"
	"errors"
	"strings"
	"net/http"
	"encoding/json"

	"github.com/parnurzeal/gorequest"
	"github.com/PuerkitoBio/goquery"
)

type ResponseOpt struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Content string `json:"content"`
}

type CoinFilter struct {
	CoinType   []string
	High       float32
	Low        float32
	Amplitude  float32
	TimePeriod int64
}

type CoinPriceMeta struct {
	Platform string
	Price    string
	Percent  string
	CoinType string
}

func Login(user UserLoginMeta) ([]*http.Cookie, error) {

	client := gorequest.New()
	formstring, errs := json.Marshal(user)
	if errs != nil {
		return nil, errors.New(fmt.Sprintf("parse user content error: %s", user))
	}
	// create cookie jar
	response, body, err := client.Post("https://api.feixiaohao.com/user/login").
		Type("multipart").
		SendString(string(formstring)).
		Timeout(10 * time.Second).
		End()
	if err != nil || response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("login http error: %s", err))
	}

	var result ResponseOpt
	json.Unmarshal([]byte(body), &result)

	if strings.Compare(result.Status, "success") != 0 {
		return nil, errors.New(fmt.Sprintf("login fails: %s", result.Content))
	}

	return (*http.Response)(response).Cookies(), nil
}

func GetUserTicket(cookies []*http.Cookie, filter CoinFilter) ([]CoinPriceMeta, error) {

	metas := make([]CoinPriceMeta, 0, len(filter.CoinType))

	client := gorequest.New()

	response, _, errs := client.Get("https://www.feixiaohao.com/userticker/").
		AddCookies(cookies).
		Timeout(15 * time.Second).
		End()
	if response.StatusCode != 200 || errs != nil {
		return nil, errors.New(fmt.Sprintf("get user ticket error: %s", errs))
	}
	query, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("parse html error: %s", err))
	}
	query.Find(".new-table.new-table-custom#table tbody>tr").Each(func(i int, selection *goquery.Selection) {

		var text string
		if StringListContains(filter.CoinType, selection.Find("td").Eq(1).Text()) {
			var newNode = CoinPriceMeta{}
			for _, value := range []int{1, 2, 3, 6} {
				text = strings.TrimSpace(selection.Find("td").Eq(value).Text())
				if value == 1 {
					newNode.CoinType = text
				} else if value == 2 {
					newNode.Platform = text
				} else if value == 3 {
					newNode.Price = text
				} else if value == 6 {
					newNode.Percent = text
				}
			}
			metas = append(metas, newNode)
		}

	})

	return metas, nil
}

// StringListContains check if elements in array
func StringListContains(array []string, element string) bool {

	for _, value := range array {
		if strings.Contains(element, value) {
			return true
		}
	}
	return false
}
