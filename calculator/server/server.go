package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"

	"grpc-go-course/calculator/calculatorpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct{}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Sum function was invoked with %v\n", req)
	num1 := req.FirstNumber
	num2 := req.SecondNumber
	result := num1 + num2
	res := &calculatorpb.SumResponse{
		Result: result,
	}
	return res, nil
}

func (*server) Prime(req *calculatorpb.PrimeRequest, stream calculatorpb.CalculatorService_PrimeServer) error {
	fmt.Printf("Prime function was invoked with %v\n", req)
	num := req.GetNumber()
	divisor := int32(2)
	for num > 1 {
		if num%divisor == 0 {
			stream.Send(&calculatorpb.PrimeResponse{
				Result: divisor,
			})
			num = num / divisor
		} else {
			divisor += 1
		}
	}
	return nil
}

func (*server) ComputedAvg(stream calculatorpb.CalculatorService_ComputedAvgServer) error {
	fmt.Println("ComputedAvg function was invoked with a streaming request")
	var numbers []int64

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// we have finished the client stream
			sum := int64(0)
			for _, number := range numbers {
				sum += number
				fmt.Printf("current sum: %v\n", sum)
			}
			result := float64(float64(sum) / float64(len(numbers)))
			return stream.SendAndClose(&calculatorpb.ComputedAvgResponse{
				Avgresult: result,
			})
		}
		if err != nil {
			log.Fatalf("error while reading client stream: %v", err)
		}

		numbers = append(numbers, req.GetNumber())
	}
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	fmt.Println("FindMaximum function was invoked with a streaming request")
	currentMaxNumber := int32(0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}

		num := req.GetNumber()
		if currentMaxNumber < num {
			currentMaxNumber = num

			err := stream.Send(&calculatorpb.FindMaximumResponse{
				MaximumNumber: currentMaxNumber,
			})
			if err != nil {
				return err
			}
		}
	}
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Printf("SquareRoot function was invoked with %v\n", req)
	number := req.GetNumber()

	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Received a negative number: %v", number))
	}

	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func main() {
	fmt.Println("Starting Calculator Service ...")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
