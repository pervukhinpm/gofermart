package model

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusRegistered OrderStatus = "REGISTERED"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
	OrderStatusInvalid    OrderStatus = "INVALID"
)
