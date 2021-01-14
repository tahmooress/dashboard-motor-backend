package utility

//User is model for active users
type User struct {
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	Accessibility []string `json:"accessibility"`
}

//LoginResponse is a type for login users
type LoginResponse struct {
	Err           string   `json:"err"`
	Result        string   `json:"result"`
	Accessibility []string `json:"accessibility"`
}

//Response is a type for response of handlers
type Response struct {
	Err    string `json:"err"`
	Result string `json:"result"`
}

//Motor is a model for information of motor
type Motor struct {
	PelakNumber string `json:"pelakNumber"`
	BodyNumber  string `json:"bodyNumber"`
	ModelName   string `json:"modelName"`
	ModelYear   string `json:"modelYear"`
	Color       string `json:"color"`
}

//Factor is a model for infromation of a buy factor
type Factor struct {
	FactorNumber string   `json:"factorNumber"`
	Motor        []Motor  `json:"motor"`
	Price        string   `json:"price"`
	Date         string   `json:"date"`
	Customer     Customer `json:"customer"`
	Debts        []Debt   `json:"debts"`
	Shop         string   `json:"shop"`
}

//Customer is a model for people who we buy motor from
type Customer struct {
	CustomerName         string `json:"customerName"`
	CustomerLastName     string `json:"customerLastName"`
	CustomerMobile       string `json:"customerMobile"`
	CustomerNationalCode string `json:"customerNationalCode"`
}

//Debt is a a model for future payable accounts
type Debt struct {
	Date   string `json:"date"`
	Price  string `json:"price"`
	Status bool   `json:"status"`
}

// LookUp is a...
type LookUp struct {
	PelakNumber string `json:"pelakNumber"`
	Color       string `json:"color"`
	ModelName   string `json:"modelName"`
	BuyFactor   string `json:"buyFactor"`
}

// MotorsResult is a ...
type MotorsResult struct {
	Motors []LookUp `json:"motors"`
	Shop   string   `json:"shop"`
}

//LookUpResponse is a ...
type LookUpResponse struct {
	Result []MotorsResult `json:"result"`
	Err    string         `json:"err"`
}

//TimeFilter is a...
type TimeFilter struct {
	Shops []string `json:"shops"`
	From  string   `json:"from"`
	To    string   `json:"to"`
}

// SaleHistory is a...
type SaleHistory struct {
	PelakNumber string `json:"pelakNumber"`
	Color       string `json:"color"`
	ModelName   string `json:"modelName"`
	SellFactor  string `json:"sellFactor"`
	Price       string `json:"price"`
	Date        string `json:"date"`
}

// SaleResult is a...
type SaleResult struct {
	Sales []SaleHistory `json:"sales"`
	Shop  string        `json:"shop"`
}

// SaleHistoryResponse is a ...
type SaleHistoryResponse struct {
	Result []SaleResult `json:"result"`
	Err    string       `json:"err"`
}

// Account is a type for receiveable and payable accounts
type Account struct {
	FactorNumber string `json:"factorNumber"`
	PelakNumber  string `json:"pelakNumber"`
	Price        string `json:"price"`
	Date         string `json:"date"`
	// CustomerName     string `json:"customerName"`
	// CustomerLastName string `json:"customerLastName"`
	// CustomerMobile   string `json:"customerMobile"`
	Customer Customer `json:"customer"`
}

// AccountsResult is receivables or payables accounts of a specified shop
type AccountsResult struct {
	List []Account `json:"list"`
	Shop string    `json:"shop"`
}

// AccountsResponse is a type for sending to client as response of specified shops receivables or payables accounts
type AccountsResponse struct {
	Result []AccountsResult `json:"result"`
	Err    string           `json:"err"`
}

type partialRec struct {
}
