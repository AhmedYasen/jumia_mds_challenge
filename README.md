# How to use
use `go run main.go --help` for getting list of commands and `go run main.go server run` to run with the default arguments

## Commands
### Server commands
- `go run main.go server run` -> run with default arguments add `--help` for more information about the default values; 
##### list of arguments
- `--ip` for setting ip address
- `--port` or `-p` for setting port
- `--jobs` or `-j` for setting max. number of running jobs for one request(worker)
- `--queue` or `-q` for setting number of maximum request in a queue

### DB commands [WIP]


# REST APIs

###  Get a product by sku

#### request
```
GET    /product/:sku
```
#### response
`OK 200`
##### **Body** `[json objects(products)]`
```
[<Product>]
Array of the product, each element belongs to a different country
```
---
###  Consume stock from a product
#### request
```
POST   /product/consume
```
##### **Body** `json object`
```
{
	"sku": string[required]
	"stock": int[required]
	"country": string[option]
}
```
##### notes
- if you **didn't specify country** it will **consume all** country's stock
- stock can be sent with minus or plus value at the end it will take it as minus value **for example** send `stock: 2` or `stock: -2` will be processed as `-2`
- if the stock doesn't exist it will return with `404 status`

#### response
`OK 200` or `NOT FOUND 404`


---
### Bulk Update Status
#### request
```
GET    /product/bulk-update-status/:id
```
#### response
`OK 200`
##### **body** `string status`
`status is one of these values [WAITING, PROCESSING, FAILED, DONE]`

---

### Bulk update
#### request
```
POST   /product/bulk-update
```
##### **Body** `form-data`
```
file: <path-to-file>
```
#### response
`Accepted 202`
##### **Body** `string with url for status checking`

##### notes:
- if product doesn't exists and the stock with **minus** value the product will be created with `stock = 0`

## Bulk update steps
- a request with `csv` file sent to `bulk-update` api
- the api starts to `unmarshall` the file
- set the status of the job to `waiting`
- send the `job` to the `worker pool` 
- response with a `url-path` to check the job status

