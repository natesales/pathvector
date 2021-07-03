package cmd

import (
	"context"
	"io"
	"log"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/natesales/pathvector/proto"
)

var (
	server string
	asn    uint32
)

func init() {
	reloadCmd.Flags().StringVarP(&server, "server", "s", "localhost:8084", "remote gRPC endpoint")
	reloadCmd.Flags().Uint32VarP(&asn, "asn", "a", 0, "ASN to reload (0 for all ASNs)")
	rootCmd.AddCommand(reloadCmd)
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.Dial(server, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("gRPC dial: %v", err)
		}

		client := protobuf.NewReloadServiceClient(conn)
		in := &protobuf.ReloadRequest{Asn: asn}
		stream, err := client.FetchResponse(context.Background(), in)
		if err != nil {
			log.Fatalf("gRPC stream open: %v", err)
		}

		done := make(chan bool)
		go func() {
			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					done <- true
					return
				}
				if err != nil {
					log.Fatalf("gRPC stream receive: %v", err)
				}
				log.Print(resp.Message)
			}
		}()
		<-done
	},
}
