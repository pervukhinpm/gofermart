package repository

import "errors"

var ErrOrderAlreadyExist = errors.New("order already exists")
var ErrOrderAlreadyCreatedByUser = errors.New("order already exists by user")
var ErrNoOrders = errors.New("no orders found")
var ErrUserDuplicated = errors.New("duplicated user")
var ErrNoAuthUser = errors.New("no auth user")
var ErrNoWithdrawals = errors.New("withdrawals not found")
