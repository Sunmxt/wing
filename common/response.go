package common

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type RedirectResponse struct {
	Next string `json:"next"`
}
