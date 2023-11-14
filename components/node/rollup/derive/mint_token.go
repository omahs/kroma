package derive

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"

	"github.com/kroma-network/kroma/bindings/predeploys"
	"github.com/kroma-network/kroma/utils/service/solabi"
)

const (
	MintTokenFuncSignature = "mint(uint256)"
	MintTokenArguments     = 1
	MintTokenLen           = 4 + 32*MintTokenArguments
	MintTxGas              = 1_000_000
)

var (
	MintTokenFuncBytes4       = crypto.Keccak256([]byte(MintTokenFuncSignature))[:4]
	MintTokenDepositorAddress = common.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddead0002")
	KromaTokenMinterAddress   = predeploys.KromaTokenMinterAddr
)

// MintToken presents the information stored in a L1Block.setL1BlockValues call
type MintToken struct {
	TotalAmount *big.Int
}

// Binary Format
// +---------+--------------------------+
// | Bytes   | Field                    |
// +---------+--------------------------+
// | 4       | Function signature       |
// | 32      | TotalAmount              |
// +---------+--------------------------+

func (m *MintToken) MarshalBinary() ([]byte, error) {
	w := bytes.NewBuffer(make([]byte, 0, MintTokenLen))
	if err := solabi.WriteSignature(w, MintTokenFuncBytes4); err != nil {
		return nil, err
	}
	if err := solabi.WriteUint256(w, m.TotalAmount); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *MintToken) UnmarshalBinary(data []byte) error {
	if len(data) != MintTokenLen {
		return fmt.Errorf("data is unexpected length: %d", len(data))
	}
	reader := bytes.NewReader(data)

	var err error
	if _, err := solabi.ReadAndValidateSignature(reader, MintTokenFuncBytes4); err != nil {
		return err
	}
	if m.TotalAmount, err = solabi.ReadUint256(reader); err != nil {
		return err
	}
	if !solabi.EmptyReader(reader) {
		return errors.New("too many bytes")
	}
	return nil
}

// MintTokenDepositTxData is the inverse of MintTokenDeposit.
func MintTokenDepositTxData(data []byte) (MintToken, error) {
	var m MintToken
	err := m.UnmarshalBinary(data)
	return m, err
}

// MintTokenDeposit creates a mint token transaction.
func MintTokenDeposit() (*types.DepositTx, error) {
	infoDat := MintToken{
		TotalAmount: new(big.Int).SetUint64(params.Ether),
	}
	data, err := infoDat.MarshalBinary()
	if err != nil {
		return nil, err
	}

	source := MintTokenDepositSource{}
	// Set a very large gas limit with to ensure
	// that the mint token transaction does not run out of gas.
	return &types.DepositTx{
		SourceHash: source.SourceHash(),
		From:       MintTokenDepositorAddress,
		To:         &KromaTokenMinterAddress,
		Mint:       nil,
		Value:      big.NewInt(0),
		Gas:        MintTxGas,
		Data:       data,
	}, nil
}

// MintTokenDepositBytes returns a serialized mint token transaction.
func MintTokenDepositBytes() ([]byte, error) {
	dep, err := MintTokenDeposit()
	if err != nil {
		return nil, fmt.Errorf("failed to create mint token tx: %w", err)
	}
	l1Tx := types.NewTx(dep)
	opaqueL1Tx, err := l1Tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to encode mint token tx: %w", err)
	}
	return opaqueL1Tx, nil
}
