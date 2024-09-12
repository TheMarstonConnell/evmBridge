package types

import (
	"encoding/json"
)

type InstantiateMsg struct {
	BindingsCodeId int `json:"bindings_code_id"`
}

type ExecuteFactoryMsg struct {
	CreateBindingsV2    *ExecuteMsgCreateBindingsV2    `json:"create_bindings_v2,omitempty"`
	FundBindings        *ExecuteMsgFundBindings        `json:"fund_bindings,omitempty"`
	CallBindings        *ExecuteMsgCallBindings        `json:"call_bindings,omitempty"`
	CallStorageBindings *ExecuteMsgCallStorageBindings `json:"call_storage_bindings,omitempty"`
}

type ExecuteMsgCreateBindingsV2 struct {
	UserEvmAddress *string `json:"user_evm_address,omitempty"`
}

type ExecuteMsgCallBindings struct {
	EvmAddress *string     `json:"evm_address,omitempty"`
	Msg        *ExecuteMsg `json:"msg,omitempty"`
}

type ExecuteMsgFundBindings struct {
	EvmAddress *string `json:"evm_address,omitempty"`
	Amount     *int64  `json:"amount,omitempty"`
}

type ExecuteMsgCallStorageBindings struct {
	EvmAddress *string     `json:"evm_address,omitempty"`
	Msg        *ExecuteMsg `json:"msg,omitempty"`
}

// ToString returns a json string representation of the message
func (m *ExecuteFactoryMsg) ToString() string {
	return toString(m)
}

// Encode returns a json byte representation of the message
func (m *ExecuteFactoryMsg) Encode() []byte {
	return encode(m)
}

func toString(v any) string {
	return string(encode(v))
}

func encode(v any) []byte {
	jsonBz, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return jsonBz
}
