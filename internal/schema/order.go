package schema

type OrderUID string

type SeqNumber uint64

type Order struct {
	OrderUID        OrderUID `json:"order_uid"`
	TrackNumber     string   `json:"track_number"`
	Entry           string   `json:"entry"`
	Delivery        Delivery
	Payment         Payment
	Items           Items
	Locale          string `json:"locale"`
	InternalSign    string `json:"internal_signature"`
	CustomerID      string `json:"customer_id"`
	DeliveryService string `json:"delivery_service"`
	Shardkey        int    `json:"shardkey"`
	SmID            int    `json:"sm_id"`
	DateCreated     string `json:"date_created"`
	OofShard        int    `json:"oof_shard"`
}

type Delivery struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Zip    int    `json:"zip"`
	City   string `json:"city"`
	Adress string `json:"adress"`
	Region string `json:"region"`
	Email  string `json:"email"`
}

type Payment struct {
	Transaction   string `json:"transaction"`
	RequestID     string `json:"request_id"`
	Currency      string `json:"currency"`
	Provider      string `json:"provider"`
	Amount        int    `json:"amount"`
	PaymentDT     int    `json:"payment_dt"`
	Bank          string `json:"bank"`
	DeliveryConst int    `json:"delivery_cost"`
	GoodsTotal    int    `json:"goods_total"`
	CustomFee     int    `json:"custom_fee"`
}

type Items struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        int    `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}
