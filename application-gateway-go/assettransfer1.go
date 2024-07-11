/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

const (
	mspID        = "Org1MSP"
	cryptoPath   = "../../test-network/organizations/peerOrganizations/org1.example.com"
	certPath     = cryptoPath + "/users/User1@org1.example.com/msp/signcerts/cert.pem"
	keyPath      = cryptoPath + "/users/User1@org1.example.com/msp/keystore/"
	tlsCertPath  = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint = "localhost:7051"
	gatewayPeer  = "peer0.org1.example.com"
)

var now = time.Now()
var assetId = fmt.Sprintf("asset%d", now.Unix()*1e3+int64(now.Nanosecond())/1e6)

// runPythonScript 执行指定的 Python 脚本并返回其输出
func runPythonScript(script string) (string, error) {
	cmd := exec.Command("python3", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func main() {
	var (
		initLedgerCount, getAllAssetsCount, createAssetCount, readAssetByIDCount, updateMerchantSigCount, updateAssetPaymentSuccessCount int
		initLedgerLatency, getAllAssetsLatency, createAssetLatency, readAssetByIDLatency, updateMerchantSigLatency, updateAssetPaymentSuccessLatency time.Duration
	)
	startT := time.Now()

	// 记录并输出 newGrpcConnection() 的时间
	start := time.Now()
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()
	elapsed := time.Since(start)
	fmt.Printf("newGrpcConnection took %v\n", elapsed)

	// 记录并输出 newIdentity() 和 newSign() 的时间
	start = time.Now()
	id := newIdentity()
	elapsed = time.Since(start)
	fmt.Printf("newIdentity took %v\n", elapsed)

	start = time.Now()
	sign := newSign()
	elapsed = time.Since(start)
	fmt.Printf("newSign took %v\n", elapsed)

	// 记录并输出 client.Connect 的时间
	start = time.Now()
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gw.Close()
	elapsed = time.Since(start)
	fmt.Printf("client.Connect took %v\n", elapsed)

	chaincodeName := "testgo"
	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)
	fmt.Println("connect to contract succeed: ", contract)

	// 记录并输出每个合约调用的时间
	start = time.Now()
	initLedger(contract)
	elapsed = time.Since(start)
	fmt.Printf("initLedger took %v\n", elapsed)
	initLedgerLatency += elapsed
	initLedgerCount++

	start = time.Now()
	getAllAssets(contract)
	elapsed = time.Since(start)
	fmt.Printf("getAllAssets took %v\n", elapsed)
	getAllAssetsLatency += elapsed
	getAllAssetsCount++

	start = time.Now()
	createAsset(contract)
	elapsed = time.Since(start)
	fmt.Printf("createAsset took %v\n", elapsed)
	createAssetLatency += elapsed
	createAssetCount++

	start = time.Now()
	readAssetByID(contract)
	elapsed = time.Since(start)
	fmt.Printf("readAssetByID took %v\n", elapsed)
	readAssetByIDLatency += elapsed
	readAssetByIDCount++

	start = time.Now()
	updateMerchantSig(contract)
	elapsed = time.Since(start)
	fmt.Printf("updateMerchantSig took %v\n", elapsed)
	updateMerchantSigLatency += elapsed
	updateMerchantSigCount++

	start = time.Now()
	readAssetByID(contract)
	elapsed = time.Since(start)
	fmt.Printf("readAssetByID took %v\n", elapsed)
	readAssetByIDLatency += elapsed
	readAssetByIDCount++

	start = time.Now()
	updateAssetPaymentSuccess(contract)
	elapsed = time.Since(start)
	fmt.Printf("updateAssetPaymentSuccess took %v\n", elapsed)
	updateAssetPaymentSuccessLatency += elapsed
	updateAssetPaymentSuccessCount++

	start = time.Now()
	readAssetByID(contract)
	elapsed = time.Since(start)
	fmt.Printf("readAssetByID took %v\n", elapsed)
	readAssetByIDLatency += elapsed
	readAssetByIDCount++

	// 记录并输出运行 Python 脚本的时间
	start = time.Now()
	output, err := runPythonScript("run_docker.py")
	elapsed = time.Since(start)
	fmt.Printf("runPythonScript took %v\n", elapsed)

	if err != nil {
		fmt.Println("Error running Python script:", err)
	} else {
		fmt.Println("Python script output:", output)
	}

	tc := time.Since(startT) // 计算总耗时
	fmt.Printf("Total time cost = %v\n", tc)

	// 计算并输出每种交易类型的TPS和平均延迟
	totalDuration := time.Since(startT)

	calculateAndPrintTPS := func(name string, count int, latency time.Duration) {
		if count > 0 {
			tps := float64(count) / totalDuration.Seconds()
			avgLatency := latency / time.Duration(count)
			fmt.Printf("%s - Transactions: %d, TPS: %.2f, Average Latency: %v\n", name, count, tps, avgLatency)
		} else {
			fmt.Printf("%s - No transactions\n", name)
		}
	}

	calculateAndPrintTPS("initLedger", initLedgerCount, initLedgerLatency)
	calculateAndPrintTPS("getAllAssets", getAllAssetsCount, getAllAssetsLatency)
	calculateAndPrintTPS("createAsset", createAssetCount, createAssetLatency)
	calculateAndPrintTPS("readAssetByID", readAssetByIDCount, readAssetByIDLatency)
	calculateAndPrintTPS("updateMerchantSig", updateMerchantSigCount, updateMerchantSigLatency)
	calculateAndPrintTPS("updateAssetPaymentSuccess", updateAssetPaymentSuccessCount, updateAssetPaymentSuccessLatency)
}


// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	files, err := os.ReadDir(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := os.ReadFile(path.Join(keyPath, files[0].Name()))

	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

// This type of transaction would typically only be run once by an application the first time it was started after its
// initial deployment. A new version of the chaincode deployed later would likely not need to run an "init" function.
func initLedger(contract *client.Contract) {
	fmt.Printf("\n--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger \n")
	_, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

// Evaluate a transaction to query ledger state.
func getAllAssets(contract *client.Contract) {
	fmt.Println("\n--> Evaluate Transaction: GetAllAssets, function returns all the current assets on the ledger")

	evaluateResult, err := contract.EvaluateTransaction("GetAllAssets")

	// 打印原始返回的数据
	fmt.Printf("Raw Evaluate Result: %s\n", string(evaluateResult))
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	// 继续您原有的处理逻辑
	result := formatJSON(evaluateResult)
	fmt.Printf("*** Result:%s\n", result)
}

func createAsset(contract *client.Contract) {
	// 定义创建资产所需的参数
	assetId := assetId // 示例资产ID，实际应用中可能需要动态生成或从用户输入获取
	proposalTimeStamp := "1622548800"
	amount := "1000"              // 示例金额
	buyer := "BuyerA"             // 买家示例
	buyerSig := "BuyerASignature" // 买家签名示例
	merchant := "MerchantA"       // 商家示例

	fmt.Printf("\n--> Submit Transaction: CreateAsset, creates new asset with ID %s, ProposalTimeStamp %s, Amount %s, Buyer %s, BuyerSig %s, and Merchant %s\n",
		assetId, proposalTimeStamp, amount, buyer, buyerSig, merchant)

	// 调用智能合约的CreateAsset函数
	_, err := contract.SubmitTransaction("CreateAsset", assetId, proposalTimeStamp, amount, buyer, buyerSig, merchant)
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

// Evaluate a transaction by assetID to query ledger state.
func readAssetByID(contract *client.Contract) {
	fmt.Printf("\n--> Evaluate Transaction: ReadAsset, function returns asset attributes\n")

	evaluateResult, err := contract.EvaluateTransaction("ReadAsset", assetId)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result:%s\n", result)
}

// Evaluate a transaction by assetID to query ledger state.
func updateMerchantSig(contract *client.Contract) {
	fmt.Printf("\n--> Evaluate Transaction: UpdateSig, Updating merchant's signature\n")

	_, err := contract.SubmitTransaction("UpdateAssetMerchantSig", assetId, "MerchantSig")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("*** UpdateSig committed successfully\n")
}

// Evaluate a transaction by assetID to query ledger state.
func updateAssetPaymentSuccess(contract *client.Contract) {
	fmt.Printf("\n--> Evaluate Transaction: updateAssetPaymentSuccess, Updating merchant's signature\n")

	_, err := contract.SubmitTransaction("UpdateAssetPaymentSuccess", assetId, "true")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("*** updateAssetPaymentSuccess committed successfully\n")
}

// Submit transaction, passing in the wrong number of arguments ,expected to throw an error containing details of any error responses from the smart contract.
func exampleErrorHandling(contract *client.Contract) {
	fmt.Println("\n--> Submit Transaction: UpdateAsset asset70, asset70 does not exist and should return an error")

	_, err := contract.SubmitTransaction("UpdateAsset", "asset70", "blue", "5", "Tomoko", "300")
	if err == nil {
		panic("******** FAILED to return an error")
	}

	fmt.Println("*** Successfully caught the error:")

	switch err := err.(type) {
	case *client.EndorseError:
		fmt.Printf("Endorse error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
	case *client.SubmitError:
		fmt.Printf("Submit error for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
	case *client.CommitStatusError:
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Timeout waiting for transaction %s commit status: %s", err.TransactionID, err)
		} else {
			fmt.Printf("Error obtaining commit status for transaction %s with gRPC status %v: %s\n", err.TransactionID, status.Code(err), err)
		}
	case *client.CommitError:
		fmt.Printf("Transaction %s failed to commit with status %d: %s\n", err.TransactionID, int32(err.Code), err)
	default:
		panic(fmt.Errorf("unexpected error type %T: %w", err, err))
	}

	// Any error that originates from a peer or orderer node external to the gateway will have its details
	// embedded within the gRPC status error. The following code shows how to extract that.
	statusErr := status.Convert(err)

	details := statusErr.Details()
	if len(details) > 0 {
		fmt.Println("Error Details:")

		for _, detail := range details {
			switch detail := detail.(type) {
			case *gateway.ErrorDetail:
				fmt.Printf("- address: %s, mspId: %s, message: %s\n", detail.Address, detail.MspId, detail.Message)
			}
		}
	}
}

// Format JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}
