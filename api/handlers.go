package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/AhmedYasen/jumia_mds_challenge/model"
	workerpool "github.com/AhmedYasen/jumia_mds_challenge/workerPool"
	"github.com/gin-gonic/gin"
	"github.com/gocarina/gocsv"
)

type productReq struct {
	Sku           string `json:"sku" binding:"required"`
	Country       string `json:"country"`
	StockConsumed int64  `json:"stock_consumed" binding:"required"`
}

func GetProductBySku(c *gin.Context) {
	sku := c.Param("sku")
	product, exists := model.GetBySku(sku)

	if exists {
		c.IndentedJSON(http.StatusOK, product)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func ConsumeStock(c *gin.Context) {
	var req productReq
	if err := c.BindJSON(&req); err == nil {
		if req.StockConsumed > 0 {
			req.StockConsumed *= -1
		}
		if err, _ := model.StockUpdate(req.Sku, req.Country, req.StockConsumed); err == nil {
			c.Status(http.StatusOK)
		} else {
			c.String(http.StatusNotFound, err.Error())
		}
	} else {
		erro := fmt.Sprintf("%v", err)
		c.String(http.StatusBadRequest, erro)
	}

}

func InsertProduct(c *gin.Context) {
	var product model.Product
	if err := c.BindJSON(&product); err == nil {
		err = product.InsertProduct()
	} else {
		erro := fmt.Sprintf("%v", err)
		c.String(http.StatusBadRequest, erro)
	}
}

func BulkStockUpdate(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	products := []*model.Product{}
	if err := gocsv.Unmarshal(file, &products); err != nil {
		fmt.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}
	// bulk-update
	// job-id
	id++
	// set job-status to waiting
	go func() { statusChannel <- workerpool.JobStatus{Id: id, Status: workerpool.WAITING} }()
	// send job to a buffering channel (this channel represents a queue)
	go func() { workerChannelSender <- workerpool.Job[[]*model.Product]{Element: products, Id: id} }()
	// response with check-status url-path
	c.String(http.StatusAccepted, "GET /product/bulk-update-status/%v to know status", id)
}

func BulkUpdateStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 10)
	if err != nil {
		ctx.String(http.StatusBadRequest, "%v", err)
		return
	}
	ctx.String(http.StatusOK, bulkUpdateStatus[id].String())
}
