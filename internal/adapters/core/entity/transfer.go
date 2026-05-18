package entity

type Transfer struct {
	ID            int64  `json:"id"`
	FromAccountID int64  `json:"from_account_id"`
	ToAccountID   int64  `json:"to_account_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	FromBalance   int64  `json:"from_balance"`
	ToBalance     int64  `json:"to_balance"`
}

type CreateTransferRequest struct {
	From        int64  `json:"from" binding:"required"`
	To          int64  `json:"to" binding:"required"`
	Amount      int64  `json:"amount" binding:"required"`
	Currency    string `json:"currency" binding:"required"`
	FromBalance int64  `json:"from_balance"`
	ToBalance   int64  `json:"to_balance"`
}

type UpdateTransferStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type TransferResponse struct {
	ID          int64  `json:"id"`
	Status      string `json:"status"`
	FromBalance int64  `json:"from_balance"`
	ToBalance   int64  `json:"to_balance"`
}

type IdempotencyRecord struct {
	Key        string
	TransferID int64
	Status     string
}
