package common

type Response struct {
    Success bool `json:"success"`
    Data    interface{} `json:"data"`
}

type RedirectResponse struct {
    Next string `json:"next"`
}
