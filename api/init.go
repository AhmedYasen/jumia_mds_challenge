package server

import (
	"time"

	"github.com/AhmedYasen/jumia_mds_challenge/model"
	workerpool "github.com/AhmedYasen/jumia_mds_challenge/workerPool"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// used to send jobs to the worker pool
var workerChannelSender chan<- workerpool.Job[[]*model.Product]

// hold request status of bulk-updates
var bulkUpdateStatus map[uint64]workerpool.STATUS
var router = gin.Default()

//hold last job id (just counter)
var id uint64 = 0

// this channel used to update bulk-update-status
var statusChannel chan workerpool.JobStatus

func Init() {
	router.GET("/product/:sku", GetProductBySku)
	router.POST("/product/consume", ConsumeStock)
	// router.POST("/product/insert", InsertProduct)
	router.POST("/product/bulk-update", BulkStockUpdate)
	router.GET("/product/bulk-update-status/:id", BulkUpdateStatus)
}

func Run(ip string, workerPoolJobs uint64, requestQueueSize uint64) {

	bulkWorkerInit(workerPoolJobs, requestQueueSize)

	// Run the server
	if err := router.Run(ip); err != nil {
		log.Error(err)
	}
}

func bulkWorkerInit(workerPoolJobs uint64, requestQueueSize uint64) {
	// this channel used to update bulk-update-status
	statusChannel = make(chan workerpool.JobStatus, workerPoolJobs)

	bulkUpdateStatus = make(map[uint64]workerpool.STATUS)

	// this channel used to send jobs to the worker-pool
	var workerChannel = make(chan workerpool.Job[[]*model.Product], requestQueueSize)

	// worker pool initialization
	var workerPool = workerpool.New(workerPoolJobs, workerChannel)

	// get the sender part of the jobs channel
	workerChannelSender = workerChannel

	// update the bulk-update-status requests here
	go func() {
		for st := range statusChannel {
			log.Infof("Status received: %v", st)
			bulkUpdateStatus[st.Id] = st.Status
		}
	}()

	// Run the bulk update worker pool
	go workerPool.Run(func(p []*model.Product, id uint64) {
		// the begining of processing the job
		statusChannel <- workerpool.JobStatus{Id: id, Status: workerpool.PROCESSING}
		// processing
		if err := model.BulkStockUpdate(p); err != nil {
			log.Warnf("Job (%v) Failed: %v", id, err)
			// processing failed
			statusChannel <- workerpool.JobStatus{Id: id, Status: workerpool.FAILED}
		} else {
			log.Infof(" %v Job Done", time.Now().Format("2020/08/08 - 20:22:34"))
			// processing done
			statusChannel <- workerpool.JobStatus{Id: id, Status: workerpool.DONE}
		}
	})
}
