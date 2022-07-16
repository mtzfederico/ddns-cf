package main

import (
	"fmt"

	"github.com/Jeffail/gabs"
)

func test() {
	fmt.Println("test")
	jsonParsed, err := gabs.ParseJSON([]byte(`{"result":[{"id":"2776203cd06872e652b2b228eb38f10a","zone_id":"2247b5147ecdee1e3fb10e7bab455ef6","zone_name":"fedemtz66.tech","name":"h.fedemtz66.tech","type":"A","content":"189.210.48.34","proxiable":true,"proxied":false,"ttl":300,"locked":false,"meta":{"auto_added":false,"managed_by_apps":false,"managed_by_argo_tunnel":false,"source":"primary"},"created_on":"2022-05-17T16:06:45.177553Z","modified_on":"2022-05-17T16:06:45.177553Z"}],"success":true,"errors":[],"messages":[],"result_info":{"page":1,"per_page":100,"count":1,"total_count":1,"total_pages":1}}`))
	fmt.Println(("here"))
	if err != nil {
		fmt.Println("Error parsing json:", err)
	}
	fmt.Println("here1")

	// result := jsonParsed.S("result").Index(0)
	// ip, ok := result.Path("content").Data().(string)
	ip, ok := jsonParsed.S("result").Index(0).Path("content").Data().(string)

	if ok {
		fmt.Println("ok")
	} else {
		fmt.Println("not ok (like me)")
	}

	fmt.Println("ip:", ip)

	/*
		id, ok := jsonParsed.Path("result.0.zone_id").Data().(string)

		if ok {
			fmt.Println("ok")
		} else {
			fmt.Println("not ok (like me)")
		}

		fmt.Println("id:", id)
	*/
}

/*
	requestBody, _ := json.Marshal(map[string]string{
		"type":    recordType,
		"name":    name,
		"content": IP,
		"ttl":     ttl,
		"proxied": isProxied,
	})
*/
