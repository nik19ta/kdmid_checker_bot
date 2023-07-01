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
	// Заменить все точки и пробелы на пустые строки и преобразовать все символы в верхний регистр
	cityName = strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(cityName, ".", ""), " ", ""))

	// Запрос к серверу
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://info.midpass.ru/api/deptbook/departments", nil)
	if err != nil {
		return "", err
	}
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

	// Поиск города и возврат DepartmentCode
	for _, department := range departments {
		if strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(department.City, ".", ""), " ", "")) == cityName {
			return department.DepartmentCode, nil
		}
	}

	return "", errors.New("city not found")
}
