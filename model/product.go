package model

import (
	"errors"
	"fmt"
	"sync"

	workerpool "github.com/AhmedYasen/jumia_mds_challenge/workerPool"
)

type Product struct {
	Sku     string `json:"sku" csv:"sku" binding:"required"`
	Name    string `json:"name" csv:"name"`
	Stock   int64  `json:"stock" csv:"stock_change" binding:"required"`
	Country string `json:"country" csv:"country"`
}

type DbErrorType int

const (
	NONE DbErrorType = iota
	DB_ERR
	RECORD_NOT_FOUND
)

func nonReturnStmt(query string, args ...any) error {

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	stmt, err := tx.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(args...)

	if err != nil {
		return err
	}

	return nil
}

func (p *Product) InsertProduct() error {

	err := nonReturnStmt("INSERT OR IGNORE INTO product (sku, name) VALUES (?, ?)", p.Sku, p.Name)

	if err != nil {
		return err
	}

	err = nonReturnStmt("INSERT INTO country_stock (sku, country, stock) VALUES (?, ?, ?)", p.Sku, p.Country, p.Stock)
	if err != nil {
		return err
	}

	return nil

}

func GetBySku(sku string) ([]Product, bool) {

	rows, err := dbConn.Query("SELECT product.sku, product.name, country_stock.country, country_stock.stock FROM product INNER JOIN country_stock ON product.sku=country_stock.sku WHERE product.sku=?", sku)

	if err != nil {
		fmt.Println("err: ", err)
		return nil, false
	}
	defer rows.Close()

	products := make([]Product, 0)

	for rows.Next() {
		product := Product{}
		err := rows.Scan(&product.Sku, &product.Name, &product.Country, &product.Stock)
		if err != nil {
			return nil, false
		}

		products = append(products, product)
		fmt.Println(product)
	}

	err = rows.Err()

	if err != nil {
		return nil, false
	}

	return products, true
}

func checkIfStockExists(sku string, country string) (error, DbErrorType) {
	row := dbConn.QueryRow("SELECT COUNT(*) FROM country_stock WHERE sku=? AND country=?", sku, country)

	if row.Err() != nil {
		return row.Err(), DB_ERR
	}

	var value int

	row.Scan(&value)

	if value > 0 {
		return nil, NONE
	} else {
		return errors.New("RECORD_NOT_FOUND"), RECORD_NOT_FOUND
	}

}

func StockUpdate(sku string, country string, change int64) (error, DbErrorType) {
	err, errType := checkIfStockExists(sku, country)

	if err != nil {
		return err, errType
	}

	err = nonReturnStmt("UPDATE country_stock SET stock = stock + ? WHERE sku = ? AND country = ?", change, sku, country)

	if err != nil {
		return err, DB_ERR
	}

	return nil, NONE
}

func BulkStockUpdate(products []*Product) error {
	// steps
	// ---------
	// - prepare workerChannel to send jobs (each job is part of products slice)
	// - prepare waitGroup (wg) to wait all jobs to be successfully finished
	// - prepare waitGroupChannel to notify that all jobs successfully finished
	// - prepare errChannel to send resulted error from the worker
	// - prepare workerCloseChannel slice to send cancel notification for workers when error happened
	// - init workerPool and run it
	// - send jobs (each job is a slice of the whole products slice)
	// - wait on errChannel and wgChannel

	// this function splits the products slice to <buffSize>
	// and send them to the worker-channel to be served
	const buffSize = 50
	// a channel to send jobs of product slice
	var workerChannel = make(chan workerpool.Job[[]*Product], buffSize)
	defer close(workerChannel)

	// wait group to delect that all workers finished with no errors
	var wg sync.WaitGroup
	var wgChannel = make(chan struct{})

	go func() {
		wg.Wait()
		close(wgChannel)
	}()

	// error channel to receive the error happend
	var errChannel = make(chan error)

	// array of channels to notify close when there is an error occurred
	var workerCloseChannel [buffSize]chan bool
	for i := range workerCloseChannel {
		workerCloseChannel[i] = make(chan bool)
	}

	// worker pool initialization
	var workerPool = workerpool.New(10, workerChannel)

	go workerPool.Run(func(p []*Product, id uint64) {
		wg.Add(1)
		defer wg.Done()
		for _, v := range p {

			select {
			case <-workerCloseChannel[id]:
				{
					// stop if stop signal received
					return
				}
			default:
				{
					// update
					err, errType := StockUpdate(v.Sku, v.Country, v.Stock)
					if err != nil {
						switch errType {
						case RECORD_NOT_FOUND: // insert if the product not recorded
							// stock can't be consumed before it is exists
							if v.Stock < 0 {
								v.Stock = 0
							}
							if err := v.InsertProduct(); err != nil {
								errChannel <- err
								return
							}
						default:
							{
								errChannel <- err
								return
							}
						}
					}
				}
			}
		}
	})

	// send jobs to worker-pool
	// ------------------------

	productsLen := len(products)
	step := productsLen/buffSize + 1 // slice size (which will sent to worker-pool)
	id := uint64(0)

	for loopIndex := 0; loopIndex < productsLen; {
		// calculate slice end
		end := loopIndex + step
		if end >= productsLen {
			end = productsLen
		}
		// send job
		workerChannel <- workerpool.Job[[]*Product]{Element: products[loopIndex:end], Id: id}
		id++
		// claculate next slice start
		loopIndex = end
	}

	// if error received => stop all workers and return error
	// if wgChannel closed it means all workers called wg.Done() with no errors
	select {
	case err := <-errChannel:
		{
			if err != nil {
				for _, ch := range workerCloseChannel {
					ch <- false
				}
			}
			return err
		}
	case <-wgChannel:
		{
			return nil
		}
	}

}
