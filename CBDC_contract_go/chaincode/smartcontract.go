package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	AppraisedValue         int    `json:"appraisedValue"`
	ID                     string `json:"id"`
	ProposalTimeStamp      int64  `json:"proposalTimeStamp"`
	DeliveryTimeStamp      int64  `json:"deliveryTimeStamp"`
	Amount                 int    `json:"amount"`
	Buyer                  string `json:"buyer"`
	Merchant               string `json:"merchant"`
	BuyerSig               string `json:"buyerSig"`
	MerchantSig            string `json:"merchantSig"`
	PlatformSig            string `json:"platformSig"`
	PartySig               string `json:"partySig"`
	PaymentSuccess         bool   `json:"paymentSuccess"`
	DeliverSuccess         bool   `json:"deliverSuccess"`
	BuyerConfirmDeliverSig string `json:"buyerConfirmDeliverSig"`
	TradeSuccess           bool   `json:"tradeSuccess"`
}


// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{
			ID:                "init0",
			ProposalTimeStamp: 1622548800, // 示例时间戳
			Amount:            1000,
			Buyer:             "BuyerA",
			BuyerSig:          "BuyerASignature",
			Merchant:          "MerchantA",
		},
		{
			ID:                "init1",
			ProposalTimeStamp: 1622548800, // 示例时间戳
			Amount:            1000,
			Buyer:             "BuyerB",
			BuyerSig:          "BuyerASignature",
			Merchant:          "MerchantB",
		},
		{
			ID:                "init2",
			ProposalTimeStamp: 1622548800, // 示例时间戳
			Amount:            1000,
			Buyer:             "BuyerC",
			BuyerSig:          "BuyerASignature",
			Merchant:          "MerchantC",
		},
		{
			ID:                "init3",
			ProposalTimeStamp: 1622548800, // 示例时间戳
			Amount:            1000,
			Buyer:             "BuyerD",
			BuyerSig:          "BuyerASignature",
			Merchant:          "MerchantD",
		},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, proposalTimeStamp int64, amount int, buyer string, buyerSig string, merchant string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	// 创建一个新的Asset实例，只包含必需的字段
	asset := Asset{
		ID:                id,
		ProposalTimeStamp: proposalTimeStamp,
		Amount:            amount,
		Buyer:             buyer,
		BuyerSig:          buyerSig,
		Merchant:          merchant,
		// 其他字段保留其零值或默认值
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAssetDeliveryTimeStamp updates the DeliveryTimeStamp of an asset.
func (s *SmartContract) UpdateAssetDeliveryTimeStamp(ctx contractapi.TransactionContextInterface, id string, newTimeStamp int64) error {
	return s.updateAssetField(ctx, id, func(asset *Asset) {
		asset.DeliveryTimeStamp = newTimeStamp
	})
}

// UpdateAssetPaymentSuccess updates the PaymentSuccess of an asset.
func (s *SmartContract) UpdateAssetPaymentSuccess(ctx contractapi.TransactionContextInterface, id string, newStatus bool) error {
	return s.updateAssetField(ctx, id, func(asset *Asset) {
		asset.PaymentSuccess = newStatus
	})
}

// UpdateAssetMerchantSig updates the MerchantSig of an asset.
func (s *SmartContract) UpdateAssetMerchantSig(ctx contractapi.TransactionContextInterface, id string, newSig string) error {
	return s.updateAssetField(ctx, id, func(asset *Asset) {
		asset.MerchantSig = newSig
	})
}

// updateAssetField is a helper function to update an asset field
func (s *SmartContract) updateAssetField(ctx contractapi.TransactionContextInterface, id string, updateFunc func(*Asset)) error {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return err
	}

	// Call the passed function to update the asset
	updateFunc(&asset)

	// Serialize the updated asset and write it back to the state
	updatedAssetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, updatedAssetJSON)
}

// DeleteAsset deletes a given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
