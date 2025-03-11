package models

type City struct {
	ID             int     `json:"id"`
	DepartmentCode string  `json:"department_code"`
	InseeCode      string  `json:"insee_code"`
	ZipCode        string  `json:"zip_code"`
	Name           string  `json:"name"`
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
}
