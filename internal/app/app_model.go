package app

import (
	"time"
)

type Delivery struct{
	Name string `json:"name" db:"Name"`
	Phone string `json:"phone" db:"Phone"`
	Zip string `json:"zip" db:"Zip"`
	City string `json:"city" db:"City"`
	Address string `json:"address" db:"Address"`
	Region string `json:"region" db:"Region"`
	Email string `json:"email" db:"Email"`
}

type Payment struct{
	Transaction string `json:"transaction" db:"Transaction"`
	RequestId string `json:"request_id" db:"RequestId"`
	Currency string `json:"currency" db:"Currency"`
	Provider string	`json:"provider" db:"Provider"`
	Amount uint32 `json:"amount" db:"Amount"`
	PaymentDt int64 `json:"payment_dt" db:"PaymentDt"`
	Bank string	`json:"bank" db:"Bank"`
	DeliveryCost uint32 `json:"delivery_cost" db:"DeliveryCost"`
	GoodsTotal uint32 `json:"goods_total" db:"GoodsTotal"`
	CustomFee uint32 `json:"custom_fee" db:"CustomFee"`
}

type Item struct{
	ChrtId uint64 `json:"chrt_id" db:"ChrtId"`
	TrackNumber string `json:"track_number" db:"TrackNumber"`
	Price uint32 `json:"price" db:"Price"`
	Rid string `json:"rid" db:"Rid"`
	Name string `json:"name" db:"Name"`
	Sale uint8 `json:"sale" db:"Sale"`
	Size string `json:"size" db:"Size"`
	TotalPrice uint32 `json:"total_price" db:"TotalPrice"`
	RmId uint64 `json:"rm_id" db:"RmId"`
	NmId uint32 `json:"nm_id" db:"NmId"`
	Brand string `json:"brand" db:"Brand"`
	Status uint8 `json:"status" db:"Status"`
}

type Order struct{
	OrderUid string `json:"order_uid" db:"OrderUid"`
	TrackNumber string `json:"track_number" db:"TrackNumber"`
	Entry string `json:"entry" db:"Entry"`
	Delivery Delivery `json:"delivery"`
	Payment Payment `json:"payment"`
	Items []Item `json:"items"`
	Locale string `json:"locale" db:"Locale"`
	InternalSignature string `json:"internal_signature" db:"InternalSignature"`
   	CustomerId string `json:"customer_id" db:"CustomerId"`
    DeliveryService string `json:"delivery_service" db:"DeliveryService"`
    Shardkey string `json:"shardkey" db:"Shardkey"`
   	SmId uint8 `json:"sm_id" db:"SmId"`
    DateCreated time.Time `json:"date_created" db:"DateCreated"`
    OofShard string `json:"oof_shard" db:"OofShard"`
}