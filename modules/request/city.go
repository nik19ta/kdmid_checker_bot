package request

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type Department struct {
	DepartmentCode string `json:"DepartmentCode"`
	City           string `json:"City"`
}

func GetCityIdByName(cityName string) (string, error) {
	// * Replace all dots and spaces with empty lines and convert all characters to uppercase.
	cityName = strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(cityName, ".", ""), " ", ""))

	// * Request to the server to retrieve cities.
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://info.midpass.ru/api/deptbook/departments", nil)
	if err != nil {
		return "", err
	}

	// * We set the user agent to prevent the API from blocking the request.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	departments := []Department{}
	if err = json.Unmarshal(body, &departments); err != nil {
		return "", err
	}

	// * Searching for a city and returning the DepartmentCode.
	for _, department := range departments {
		if strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(department.City, ".", ""), " ", "")) == cityName {
			return department.DepartmentCode, nil
		}
	}

	// * If no city is found, we return an error.
	return "", errors.New("city not found")
}
