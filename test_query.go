package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func main() {
	q1 := `sum_over_time(avg(shundao_sensor{name="total_irradiance_wm2", site_name="SHUNDAO_1"})[1d:5m])`
	execute(q1)

	q2 := `avg(shundao_sensor{name="total_irradiance_wm2", site_name="SHUNDAO_1"})`
	execute(q2)
}

func execute(query string) {
	u := "http://localhost:8428/api/v1/query?query=" + url.QueryEscape(query)
	resp, err := http.Get(u)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var res map[string]interface{}
	json.Unmarshal(body, &res)
	
	fmt.Printf("Query: %s\n", query)
	data, _ := json.MarshalIndent(res["data"], "", "  ")
	fmt.Println(string(data))
	fmt.Println("- - - - -")
}
