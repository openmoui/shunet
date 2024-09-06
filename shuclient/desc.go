package shuclient

type PageInfo struct {
	PasswordEncrypt   string `json:"passwordEncrypt"`
	PublicKeyExponent string `json:"publicKeyExponent"`
	PublicKeyModulus  string `json:"publicKeyModulus"`
}

type LoginResponse struct {
	UserIndex string `json:"userIndex"`
	GeneralResponse
	ForwardURL        string `json:"forwordurl"`
	KeepAliveInterval int    `json:"keepaliveInterval"`
	CasFailErrString  string `json:"casFailErrString"`
	ValidCodeURL      string `json:"validCodeUrl"`
}

type GeneralResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}
