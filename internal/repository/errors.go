package repository

import "errors"

var ErrOrderAlreadyExist = errors.New("order already exists")
var ErrOrderAlreadyCreatedByUser = errors.New("order already exists by user")
var ErrNoOrders = errors.New("no orders found")
