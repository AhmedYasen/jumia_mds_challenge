package cmd

import (
	"fmt"

	server "github.com/AhmedYasen/jumia_mds_challenge/api"
	"github.com/AhmedYasen/jumia_mds_challenge/model"
	"github.com/spf13/cobra"
)

var srvrRootCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"srvr"},
	Short:   "Root of commands which operate on the server",
}

var ip string
var port string
var workerPoolJobs uint64
var requestQueueSize uint64
var path = "./mds"
var run = &cobra.Command{
	Use:   "run",
	Short: "run the server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := model.ConnectDatabase(path); err != nil {
			fmt.Printf("Database connection err: %v", err)
		} else {
			fmt.Println("Database connection success")
		}
		server.Init()
		server.Run(fmt.Sprintf("%v:%v", ip, port), workerPoolJobs, requestQueueSize)
	},
}

func init() {
	srvrRootCmd.Flags().StringVar(&ip, "ip", "127.0.0.1", "Set server ip")
	srvrRootCmd.Flags().StringVarP(&port, "port", "p", "8080", "Set server port")
	srvrRootCmd.Flags().StringVar(&path, "path", "./mds.db", "database path")
	srvrRootCmd.Flags().Uint64VarP(&workerPoolJobs, "jobs", "j", uint64(5), "Request max. jobs number")
	srvrRootCmd.Flags().Uint64VarP(&requestQueueSize, "queue", "q", uint64(100), "Request queue size")
	srvrRootCmd.AddCommand(run)
	rootCmd.AddCommand(srvrRootCmd)
}
