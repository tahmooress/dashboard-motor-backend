package customedb

import (
	"database/sql"
	"fmt"
)

//information of db for connection
const (
	host     = "localhost"
	port     = 5432
	user     = "mac"
	password = "900587101"
	dbname   = "test"
)

//SetDbInfo is a function for db initiation
func SetDbInfo() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
}

//Env is a wrapper type of a DB to pass to the handlers
type Env struct {
	DB *sql.DB
}

//queries for database
const (
	UserQuery             = `SELECT password, accessibility FROM users WHERE username=$1`
	CreateNewUser         = `INSERT INTO users(username, password, accessibility) VALUES($1,$2,$3)`
	InsertNewMotor        = `INSERT INTO motors (pelak_number, body_number, color,model_name, model_year) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`
	InsertBuyFactor       = `INSERT INTO factors (factor_type, shop ,factor_number, price, date, payed_amount, customer_name, customer_last_name, customer_mobile, customer_national_code, motors) VALUES ('buy', $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	InsertSellFactor      = `INSERT INTO factors (factor_type, shop ,factor_number, price, date, payed_amount, customer_name, customer_last_name, customer_mobile, customer_national_code, motors) VALUES ('sell', $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	InsertToInventory     = `INSERT INTO inventory (shop_name, motor, status, ref_buy_factor, ref_buy_type) VALUES ($1, $2, true, $3, 'buy')`
	InsertNewPayed        = `INSERT INTO accounts (account_type, price, date, status, shop, factor_number, factor_type) VALUES ('payable', $1, $2, false, $3, $4, 'buy')`
	InsertNewReceive      = `INSERT INTO accounts (account_type, price, date, status, shop, factor_number, factor_type) VALUES ('receiveable', $1, $2, false, $3, $4, 'sell')`
	UpdateInventoryStatus = `UPDATE inventory SET status = false, ref_sell_type = 'sell', ref_sell_factor = $1 WHERE shop_name = $2 AND motor = $3`
	SelectStock           = `SELECT inventory.motor, inventory.ref_buy_factor, motors.color, motors.model_name FROM inventory INNER JOIN motors ON inventory.motor = motors.pelak_number WHERE inventory.status = true AND inventory.shop_name = $1`
	SalesHistory          = `SELECT factors.factor_number, factors.price, factors.date, motors.pelak_number, motors.color, motors.model_name FROM factors INNER JOIN inventory ON factors.factor_number = inventory.ref_sell_factor INNER JOIN motors ON motors.pelak_number = inventory.motor WHERE factors.factor_type = 'sell' AND factors.shop=$1 AND factors.date >= $2 AND factors.date <= $3 ORDER BY factors.date ASC`
	UpdateReceive         = `UPDATE accounts SET status = true WHERE accounts.account_type = 'receiveable' AND accounts.factor_type = 'sell' AND accounts.shop = $1 AND accounts.factor_number = $2 AND accounts.date = $3`
	UpdatePay             = `UPDATE accounts SET status = true WHERE accounts.account_type = 'payable' AND accounts.factor_type = 'buy' AND accounts.shop = $1 AND accounts.factor_number = $2 AND accounts.date = $3`
	SelectAllPayAccounts  = `SELECT accounts.factor_number, accounts.price, accounts.date, inventory.motor, factors.customer_name, factors.customer_last_name, factors.customer_mobile FROM accounts INNER JOIN inventory ON accounts.factor_number = inventory.ref_buy_factor INNER JOIN factors ON inventory.ref_buy_factor = factors.factor_number WHERE factors.factor_type = 'buy' AND inventory.shop_name = $1 AND accounts.account_type = 'payable' AND accounts.status = false AND accounts.date >= $2 AND accounts.date <= $3  ORDER BY accounts.date ASC`
	SelectAllRecAccounts  = `SELECT accounts.factor_number, accounts.price, accounts.date, inventory.motor, factors.customer_name, factors.customer_last_name, factors.customer_mobile FROM accounts INNER JOIN inventory ON accounts.factor_number = inventory.ref_sell_factor INNER JOIN factors ON inventory.ref_sell_factor = factors.factor_number WHERE factors.factor_type = 'sell' AND inventory.shop_name = $1 AND accounts.account_type = 'receiveable' AND accounts.status = false AND accounts.date >= $2 AND accounts.date <= $3 ORDER BY accounts.date ASC`
	UpdateReceivePartly   = `UPDATE accounts SET price = $1 WHERE accounts.account_type = 'receiveable' AND accounts.factor_type = 'sell' AND accounts.status = false AND accounts.shop = $2 AND accounts.factor_number = $3 AND accounts.date = $4`
)
