package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"dashboard.motor/customedb"
	"dashboard.motor/utility"
	"github.com/dgrijalva/jwt-go"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

//SECRETKEY is the key for signing token
const SECRETKEY = "secret"

//AuthMiddleWare is middleWare that handles the authentication process
func AuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header["Authorization"][0], "bearer ")
		if len(authHeader) != 2 {
			fmt.Println("malformed token")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("malformed token"))
		}
		token, err := jwt.Parse(authHeader[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(SECRETKEY), nil
		})
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), "props", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			fmt.Println(err)
		}
	})
}

//LoginHandler is responsible for handling login actions
func LoginHandler(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var user utility.User
		var res utility.LoginResponse
		body, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(body, &user); err != nil {
			res.Err = err.Error()
			json.NewEncoder(w).Encode(res)
			return
		}
		var temp utility.User
		err := env.DB.QueryRow(customedb.UserQuery, user.Username).Scan(&temp.Password, pq.Array(&temp.Accessibility))
		if err != nil {
			res.Err = err.Error()
			json.NewEncoder(w).Encode(res)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(temp.Password), []byte(user.Password))
		if err != nil {
			res.Err = "password is wrong!"
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(res)
			return
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user": user.Username,
			"exp":  time.Now().Add(time.Hour * time.Duration(1)).Unix(),
			"iat":  time.Now().Unix(),
		})
		tokenString, err := token.SignedString([]byte(SECRETKEY))
		if err != nil {
			res.Err = "error occurd while generating token"
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = tokenString
		res.Accessibility = temp.Accessibility
		json.NewEncoder(w).Encode(res)
	})
}

//CreateUser is handler for creating new admin user in db
func CreateUser(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var user utility.User
		var res utility.Response
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			res.Err = "error dar ghesmate readAll"
			json.NewEncoder(w).Encode(res)
			return
		}
		err = json.Unmarshal(body, &user)
		if err != nil {
			res.Err = err.Error()
			json.NewEncoder(w).Encode(res)
			return
		}
		if len(user.Accessibility) == 0 {
			res.Err = "accessibility cant be empty, Try again!"
			json.NewEncoder(w).Encode(res)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 5)
		if err != nil {
			res.Err = "some error occurd during hashing password, Try again!"
			json.NewEncoder(w).Encode(res)
			return
		}
		result, err := env.DB.Exec(customedb.CreateNewUser, user.Username, hash, pq.Array(user.Accessibility))
		if err != nil {
			res.Err = fmt.Sprintf("during inserting data to db %s error occurd and result was %s", err.Error(), result)
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = fmt.Sprintf("user created successfully")
		json.NewEncoder(w).Encode(res)
	})
}

//HandleBuy is a handler for handling incoming buy factor
func HandleBuy(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var factor utility.Factor
		var res utility.Response
		//reading request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			res.Err = fmt.Sprintf("during readALl: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		//unmarshal request body in factor
		err = json.Unmarshal(body, &factor)
		if err != nil {
			res.Err = fmt.Sprintf("during unmarshaling: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		//check if shop is not valid return err to client
		valids := []string{"shop_a", "shop_b", "shop_c", "warehouse"}
		check := false
		for _, v := range valids {
			if v == factor.Shop {
				check = true
			}
		}
		if !check {
			res.Err = fmt.Sprint("shop is not valid")
			json.NewEncoder(w).Encode(res)
			return
		}
		// create transactions
		ctx := context.Background()
		tx, err := env.DB.BeginTx(ctx, nil)
		if err != nil {
			res.Err = err.Error()
			json.NewEncoder(w).Encode(res)
			return
		}
		// motors is a slice of motor.pelakNumber and it used to insert into buy factor.motor
		var motors []string
		for _, m := range factor.Motor {
			motors = append(motors, m.PelakNumber)

		}
		//check if motors is empty to return error
		if len(motors) < 1 {
			res.Err = fmt.Sprintf("motor cant be empty")
			json.NewEncoder(w).Encode(res)
			return
		}
		//payedAmount is temproray to later define a payedAmount from client to send to server
		payedAmount := "000"
		_, err = tx.ExecContext(ctx, customedb.InsertBuyFactor, factor.Shop, factor.FactorNumber, factor.Price, factor.Date, payedAmount, factor.Customer.CustomerName, factor.Customer.CustomerLastName, factor.Customer.CustomerMobile, factor.Customer.CustomerNationalCode, pq.Array(motors))
		if err != nil {
			tx.Rollback()
			res.Err = fmt.Sprintf("during inserting buy factor to factors table this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		for _, f := range factor.Debts {
			_, err = tx.ExecContext(ctx, customedb.InsertNewPayed, f.Price, f.Date, factor.Shop, factor.FactorNumber)
			if err != nil {
				tx.Rollback()
				res.Err = fmt.Sprintf("during inserting payfactor to accounts table this error happened: %s", err.Error())
				json.NewEncoder(w).Encode(res)
				return
			}
		}
		// loop through Motors and insert one by one in motors and inventory tables
		for _, motor := range factor.Motor {
			_, err = tx.ExecContext(ctx, customedb.InsertNewMotor, motor.PelakNumber, motor.BodyNumber, motor.Color, motor.ModelName, motor.ModelYear)
			if err != nil {
				res.Err = fmt.Sprintf("during inserting new motors to motors table this error happened: %s", err.Error())
				json.NewEncoder(w).Encode(res)
				tx.Rollback()
				return
			}
			_, err = tx.ExecContext(ctx, customedb.InsertToInventory, factor.Shop, motor.PelakNumber, factor.FactorNumber)
			if err != nil {
				res.Err = fmt.Sprintf("during inserting new motors to inventory table this error happened: %s", err.Error())
				json.NewEncoder(w).Encode(res)
				tx.Rollback()
				return
			}
		}
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			res.Err = fmt.Sprintf("during inserting payfactor to accounts table this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = "success"
		json.NewEncoder(w).Encode(res)
		return
	})
}

//HandleSell is a handler for handling the sell request from client
func HandleSell(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var res utility.Response
		var factor utility.Factor
		//read request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			res.Err = fmt.Sprintf("error happend during reading request body: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		// unmarshal body to factor data type
		err = json.Unmarshal(body, &factor)
		if err != nil {
			res.Err = fmt.Sprintf("error happend during unmarshaling request body: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		//check if shop is not valid return err to client
		valids := []string{"shop_a", "shop_b", "shop_c", "warehouse"}
		check := false
		for _, v := range valids {
			if v == factor.Shop {
				check = true
			}
		}
		if !check {
			res.Err = fmt.Sprint("shop is not valid")
			json.NewEncoder(w).Encode(res)
			return
		}
		// create transactions
		ctx := context.Background()
		tx, err := env.DB.BeginTx(ctx, nil)
		if err != nil {
			res.Err = err.Error()
			json.NewEncoder(w).Encode(res)
			return
		}
		//motors is a slice of motor.pelakNumber and it used to insert into buy factor.motor
		var motors []string
		for _, m := range factor.Motor {
			motors = append(motors, m.PelakNumber)
		}
		//check if motors is empty to return error
		if len(motors) < 1 {
			res.Err = fmt.Sprintf("motor cant be empty")
			json.NewEncoder(w).Encode(res)
			return
		}
		//payedAmount is temproray to later define a payedAmount from client to send to server
		payedAmount := "000"
		//inset new factor into sell factor in db
		_, err = tx.ExecContext(ctx, customedb.InsertSellFactor, factor.Shop, factor.FactorNumber, factor.Price, factor.Date, payedAmount, factor.Customer.CustomerName, factor.Customer.CustomerLastName, factor.Customer.CustomerMobile, factor.Customer.CustomerNationalCode, pq.Array(motors))
		if err != nil {
			res.Err = fmt.Sprintf("during inserting sell factor into factors table this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		// loop through debts and create new receiveable record in accounts db
		for _, f := range factor.Debts {
			_, err = tx.ExecContext(ctx, customedb.InsertNewReceive, f.Price, f.Date, factor.Shop, factor.FactorNumber)
			if err != nil {
				res.Err = fmt.Sprintf("during inserting new receiveable in accounts this error happened: %s", err.Error())
				json.NewEncoder(w).Encode(res)
				return
			}
		}
		//In database change the inventory status to false and insert buy_factor_number
		for _, motor := range factor.Motor {
			result, err := tx.ExecContext(ctx, customedb.UpdateInventoryStatus, factor.FactorNumber, factor.Shop, motor.PelakNumber)
			if err != nil {
				res.Err = fmt.Sprintf("during updating inventory this error happend: %s", err.Error())
				json.NewEncoder(w).Encode(res)
				return
			}
			rows, err := result.RowsAffected()
			if err != nil {
				res.Err = fmt.Sprintf("during reslut.rowsaffected this error happened: %s", err.Error())
				tx.Rollback()
				json.NewEncoder(w).Encode(res)
				return
			}
			if rows == 0 {
				tx.Rollback()
				res.Err = fmt.Sprintf("motor and shop are not march")
				json.NewEncoder(w).Encode(res)
				return

			}
		}
		err = tx.Commit()
		if err != nil {
			res.Err = fmt.Sprintf("during comminting transactions this error happend: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = "success"
		json.NewEncoder(w).Encode(res)
		return
	})
}

//StockHandle is handler to retun stock of specified shops to client
func StockHandle(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		//locations is a type for incoming shops to look up their stokcs
		var locations struct {
			Shops []string `json:"shops"`
		}
		//res is response type for returning to the client
		var res utility.LookUpResponse
		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &locations)
		if err != nil {
			res.Err = fmt.Sprintf("during unmarshaling request this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		//loop through shops and for each of them loop up stocks
		for _, shop := range locations.Shops {
			var mrs utility.MotorsResult
			mrs.Shop = shop
			rows, err := env.DB.Query(customedb.SelectStock, shop)
			if err != nil {
				res.Err = fmt.Sprintf("during look up stocks, query to shop: %s this error happened: %s", shop, err.Error())
				json.NewEncoder(w).Encode(res)
				return
			}
			defer rows.Close()
			for rows.Next() {
				var temp utility.LookUp
				err := rows.Scan(&temp.PelakNumber, &temp.BuyFactor, &temp.Color, &temp.ModelName)
				if err != nil {
					res.Err = fmt.Sprintf("during scaning rows for shop: %s this error happened: %s", shop, err.Error())
					json.NewEncoder(w).Encode(res)
					return
				}
				err = rows.Err()
				if err != nil {
					res.Err = fmt.Sprintf("during chceks rows.err for shop: %s this error happened: %s", shop, err.Error())
					json.NewEncoder(w).Encode(res)
					return
				}
				mrs.Motors = append(mrs.Motors, temp)
			}
			if len(mrs.Motors) > 0 {
				res.Result = append(res.Result, mrs)
			}
		}
		json.NewEncoder(w).Encode(res)
	})
}

//HandleSaleHistory is a handler for returning sales history of specified shops to client
func HandleSaleHistory(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content_Type", "application/json")
		//filter is a type which receives from client and specified shops and time frame for saleHIstory
		var filter utility.TimeFilter
		var res utility.SaleHistoryResponse
		//reading request body
		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &filter)
		if err != nil {
			res.Err = fmt.Sprintf("during unmarshaling json data this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		//loop through shops and reading sales records of that shop
		for _, shop := range filter.Shops {
			var saleResult utility.SaleResult
			saleResult.Shop = shop
			rows, err := env.DB.Query(customedb.SalesHistory, shop, filter.From, filter.To)
			if err != nil {
				res.Err = fmt.Sprintf("during query to db for shop: %s this error happened: %s ", shop, err.Error())
				json.NewEncoder(w).Encode(res)
				return
			}
			defer rows.Close()
			for rows.Next() {
				var temp utility.SaleHistory
				err = rows.Scan(&temp.SellFactor, &temp.Price, &temp.Date, &temp.PelakNumber, &temp.Color, &temp.ModelName)
				if err != nil {
					res.Err = fmt.Sprintf("during scaning rows for shop: %s this error happened: %s", shop, err.Error())
					json.NewEncoder(w).Encode(res)
					return
				}
				err = rows.Err()
				if err != nil {
					res.Err = fmt.Sprintf("during scaning rows for shop: %s this error happened: %s", shop, err.Error())
					json.NewEncoder(w).Encode(res)
					return
				}

				saleResult.Sales = append(saleResult.Sales, temp)
			}
			if len(saleResult.Sales) > 0 {
				res.Result = append(res.Result, saleResult)
			}
		}
		json.NewEncoder(w).Encode(res)
		return
	})
}

//UpdateReceive is handler for updating specified shop and factors receive accounts in database
func UpdateReceive(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content_Type", "application/json")
		//factor is a data type for unmarshling data that comes from client
		var factor utility.Factor
		var res utility.Response
		//read body request
		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &factor)
		if err != nil {
			res.Err = fmt.Sprintf("during unmarshaling factor this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		//check for shop which is valid or not
		valids := []string{"shop_a", "shop_b", "shop_c", "warehouse"}
		check := false
		for _, v := range valids {
			if v == factor.Shop {
				check = true
			}
		}
		if !check {
			res.Err = fmt.Sprint("shop is not valid")
			json.NewEncoder(w).Encode(res)
			return
		}
		//query to db for updating status of a recieve account record for specified date, shop and factor number to true
		result, err := env.DB.Exec(customedb.UpdateReceive, factor.Shop, factor.FactorNumber, factor.Date)
		if err != nil {
			res.Err = fmt.Sprintf("during updating accounts table for receiving accounts in db this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		//check for mismatch parameters and records
		if n, _ := result.RowsAffected(); n == 0 {
			res.Err = fmt.Sprintf("some parameter are not match")
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = "account updated successfully"
		json.NewEncoder(w).Encode(res)
		return
	})
}

//UpdatePayable is handler for updating specified shop and factors payable accounts in database
func UpdatePayable(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content_Type", "application/json")
		//factor is a data type for unmarshling data that comes from client
		var factor utility.Factor
		var res utility.Response
		body, _ := ioutil.ReadAll(r.Body)
		//unmarshal body to factor data type
		err := json.Unmarshal(body, &factor)
		if err != nil {
			res.Err = fmt.Sprintf("during unmarshaling data this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		//check for shop which is valid or not
		valids := []string{"shop_a", "shop_b", "shop_c", "warehouse"}
		check := false
		for _, v := range valids {
			if v == factor.Shop {
				check = true
			}
		}
		if !check {
			res.Err = fmt.Sprint("shop is not valid")
			json.NewEncoder(w).Encode(res)
			return
		}
		//query to db for updating status of a payable account record for specified date, shop and factor number to true
		result, err := env.DB.Exec(customedb.UpdatePay, factor.Shop, factor.FactorNumber, factor.Date)
		if err != nil {
			res.Err = fmt.Sprintf("during updating accounts table for payable accounts in db this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		if n, _ := result.RowsAffected(); n == 0 {
			res.Err = fmt.Sprintf("some parameter are not match")
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = fmt.Sprint("account updated successfully")
		json.NewEncoder(w).Encode(res)
		return
	})
}

//HandleUnpayedPay is a handler for returning list of all the payables accounts of specified multiple shops
func HandleUnpayedPay(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content_Type", "application/json")
		//filter is a data type for unmarshaling shops and time frame which comes from clinet
		var filter utility.TimeFilter
		var res utility.AccountsResponse
		//read request body
		body, _ := ioutil.ReadAll(r.Body)
		//unmarshal request body into locations variable
		err := json.Unmarshal(body, &filter)
		if err != nil {
			res.Err = fmt.Sprintf("during unmarshaling shops this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		for _, shop := range filter.Shops {
			var accountResult utility.AccountsResult
			accountResult.Shop = shop
			rows, err := env.DB.Query(customedb.SelectAllPayAccounts, shop, filter.From, filter.To)
			if err != nil {
				res.Err = fmt.Sprintf("during query to accounts table for pay accounts of shop: %s this error happened: %s", shop, err.Error())
				json.NewEncoder(w).Encode(res)
				return
			}
			//close rows in case of error or operations ends
			defer rows.Close()
			//loop through rows and append them to result
			for rows.Next() {
				var temp utility.Account
				err = rows.Scan(&temp.FactorNumber, &temp.Price, &temp.Date, &temp.PelakNumber, &temp.Customer.CustomerName, &temp.Customer.CustomerLastName, &temp.Customer.CustomerMobile)
				if err != nil {
					res.Err = fmt.Sprintf("during scaning rows for shop %s this error happened: %s", shop, err.Error())
					json.NewEncoder(w).Encode(res)
					return
				}
				err = rows.Err()
				if err != nil {
					res.Err = fmt.Sprintf("during scaning rows for shop: %s this error happened: %s", shop, err.Error())
					json.NewEncoder(w).Encode(res)
					return
				}
				accountResult.List = append(accountResult.List, temp)
			}
			//check to get rid of shops which contains no records
			if len(accountResult.List) > 0 {
				res.Result = append(res.Result, accountResult)
			}
		}
		json.NewEncoder(w).Encode(res)
		return
	})
}

//HandleUnpayedRec is a handler for returning list of all the receivables accounts of specified multiple shops
func HandleUnpayedRec(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content_Type", "application/json")
		//filter is a data type for unmarshaling shops which comes from clinet into it
		var filter utility.TimeFilter
		var res utility.AccountsResponse
		//read request body
		body, _ := ioutil.ReadAll(r.Body)
		//unmarshal body into locations
		err := json.Unmarshal(body, &filter)
		if err != nil {
			res.Err = fmt.Sprintf("during unmarshaling request data this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		fmt.Println("filter: ", filter)
		for _, shop := range filter.Shops {
			var accountResult utility.AccountsResult
			accountResult.Shop = shop
			rows, err := env.DB.Query(customedb.SelectAllRecAccounts, shop, filter.From, filter.To)
			if err != nil {
				res.Err = fmt.Sprintf("during query to db for payable accounts of shop: %s this error happened: %s", shop, err.Error())
				json.NewEncoder(w).Encode(res)
				return
			}
			//close rows in case of error or ending the proccess
			defer rows.Close()
			//loop through rows and append them to the result
			for rows.Next() {
				var temp utility.Account
				err = rows.Scan(&temp.FactorNumber, &temp.Price, &temp.Date, &temp.PelakNumber, &temp.Customer.CustomerName, &temp.Customer.CustomerLastName, &temp.Customer.CustomerMobile)
				if err != nil {
					res.Err = fmt.Sprintf("during scaning rows for shop: %s this error happened: %s", shop, err.Error())
					json.NewEncoder(w).Encode(res)
					return
				}
				err = rows.Err()
				if err != nil {
					res.Err = fmt.Sprintf("during scaning rows for shop: %s this error happened: %s", shop, err.Error())
					json.NewEncoder(w).Encode(res)
					return
				}
				accountResult.List = append(accountResult.List, temp)
			}
			//check to get rid of shops which contains no records
			if len(accountResult.List) > 0 {
				res.Result = append(res.Result, accountResult)
			}
		}
		json.NewEncoder(w).Encode(res)
		return
	})
}

//PartlyUpdateReceives is a handler for partialy update receive accounts
func PartlyUpdateReceives(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content_Type", "application/json")
		var factor utility.Factor
		var res utility.Response
		body, _ := ioutil.ReadAll(r.Body)
		//decode json request body to factor
		err := json.Unmarshal(body, &factor)
		if err != nil {
			res.Err = fmt.Sprintf("during unmarshaling data this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		//check for shop which is valid or not
		valids := []string{"shop_a", "shop_b", "shop_c", "warehouse"}
		check := false
		for _, v := range valids {
			if v == factor.Shop {
				check = true
			}
		}
		if !check {
			res.Err = fmt.Sprint("shop is not valid")
			json.NewEncoder(w).Encode(res)
			return
		}
		//test
		fmt.Printf("from partly: %v", factor)
		//query to db for updating price amount of payable account
		result, err := env.DB.Exec(customedb.UpdateReceivePartly, factor.Price, factor.Shop, factor.FactorNumber, factor.Date)
		if err != nil {
			res.Err = fmt.Sprintf("during query to database for updating pay amount for recieve account this error happended: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		if n, _ := result.RowsAffected(); n == 0 {
			res.Err = fmt.Sprint("some parameters are not match so nothing updated")
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = "acount successfully updated"
		json.NewEncoder(w).Encode(res)
		return
	})
}

//PartlyUpdatePays is a handler for partialy update payable accounts
func PartlyUpdatePays(env *customedb.Env) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content_Type", "application/json")
		var factor utility.Factor
		var res utility.Response
		body, _ := ioutil.ReadAll(r.Body)
		//decode json request body to factor
		err := json.Unmarshal(body, &factor)
		if err != nil {
			res.Err = fmt.Sprintf("during unmarshaling data this error happened: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		//check for shop which is valid or not
		valids := []string{"shop_a", "shop_b", "shop_c", "warehouse"}
		check := false
		for _, v := range valids {
			if v == factor.Shop {
				check = true
			}
		}
		if !check {
			res.Err = fmt.Sprint("shop is not valid")
			json.NewEncoder(w).Encode(res)
			return
		}
		result, err := env.DB.Exec(customedb.UpdatePaysPartly, factor.Price, factor.Shop, factor.FactorNumber, factor.Date)
		if err != nil {
			res.Err = fmt.Sprintf("بروز رسانی حساب پرداختی با مشکل رو به رو شد: %s", err.Error())
			json.NewEncoder(w).Encode(res)
			return
		}
		if n, _ := result.RowsAffected(); n == 0 {
			res.Err = fmt.Sprint("مشخصات وارد شده برای بروز رسانی حساب پرداختی صحیح نسیت")
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = "بروز رسانی با موفقیت انجام شد"
		json.NewEncoder(w).Encode(res)
		return
	})
}
