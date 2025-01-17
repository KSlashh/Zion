/*
 * Copyright (C) 2021 The Zion Authors
 * This file is part of The Zion library.
 *
 * The Zion is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The Zion is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The Zion.  If not, see <http://www.gnu.org/licenses/>.
 */
package test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native"
	scom "github.com/ethereum/go-ethereum/contracts/native/header_sync/common"
	"github.com/ethereum/go-ethereum/contracts/native/header_sync/eth"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/ethereum/go-ethereum/crypto"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/stretchr/testify/assert"
)

func TestSyncGenesisHeader(t *testing.T) {
	header7152785, _ := hex.DecodeString("7b22706172656e7448617368223a22307837633137326261396464383763363163616531643162613031626239356366353830366465313835653966613030313065343137656666353135356335366537222c2273686133556e636c6573223a22307831646363346465386465633735643761616238356235363762366363643431616433313234353162393438613734313366306131343266643430643439333437222c226d696e6572223a22307836333562343736346431393339646661636433613830313437323631353961626332373762656363222c227374617465526f6f74223a22307838366233376564663162343537663566393664333162303963343865306133653863333738616138666531386331623431343331363331623037643430343931222c227472616e73616374696f6e73526f6f74223a22307835626663393365366461343263383138343134393766366538303762373731366161613064313432396132313130653661323737373566336631373538343536222c227265636569707473526f6f74223a22307839303533306462303165613761636466306430613334323536343934343463333633303365383231363464616334303532303464653839633837356532636464222c226c6f6773426c6f6f6d223a2230786230303031383430303030303030343230303030303134343030303030303038303030303030303038303030303030303030383030393030303030343030303030303030303030313030303030303030363031303832303030303031303030303030303030303030303830303031303830303830303830343031323030303130303038303038303030303031303030303030303030303038343030303230303038303031303030303030303030303030303030303430303030303030303031303430303230303030303230383230303030303030303030343030303030383030303030303030303030303030303030323030303038303130303130383030343130303030303830303030303030303038303432303030303030323030303030333030303031303032303030303030303030323030303031303031303030303030303230303032303030303030303030303030303130303030313030303030303030303030303030323030303030303030303038306130303030313033303030303030303030303832303030303030313430303830323030303030303031303030303030303034303030303030303430303031303030303030303630303230313030303130303030343230303034303230303030303032303138303830303030303030303034303030303035383030303030323030303030303032323030303830222c22646966666963756c7479223a2230783130326535363063222c226e756d626572223a223078366432343931222c226761734c696d6974223a223078376131323164222c2267617355736564223a223078363330633536222c2274696d657374616d70223a2230783565323431626434222c22657874726144617461223a2230786465383330323035306438663530363137323639373437393264343537343638363537323635373536643836333132653333333832653330383236633639222c226d697848617368223a22307837386637386462346565353132336139386530363361663339646663663965633464333863373739383434376462323235386665383232363437613133643465222c226e6f6e6365223a22307832376132383132336631393365663439222c2268617368223a22307839306131626339633566326532396365316636303562323366336661366262303634666462653535336138313632626664666133623962623864303630306537227d")
	param := new(scom.SyncGenesisHeaderParam)
	param.ChainID = ethChainID
	param.GenesisHeader = header7152785

	input, err := utils.PackMethodWithStruct(scom.ABI, scom.MethodSyncGenesisHeader, param)
	assert.Nil(t, err)

	caller := crypto.PubkeyToAddress(*acct)
	blockNumber := big.NewInt(1)
	extra := uint64(10)
	contractRef := native.NewContractRef(sdb, caller, caller, blockNumber, common.Hash{}, scom.GasTable[scom.MethodSyncGenesisHeader]+extra, nil)
	ret, leftOverGas, err := contractRef.NativeCall(caller, utils.HeaderSyncContractAddress, input)

	assert.Nil(t, err)

	result, err := utils.PackOutputs(scom.ABI, scom.MethodSyncGenesisHeader, true)
	assert.Nil(t, err)
	assert.Equal(t, ret, result)
	assert.Equal(t, leftOverGas, extra)

	contract := native.NewNativeContract(sdb, contractRef)
	height := getEthLatestHeight(contract)
	assert.Equal(t, uint64(7152785), height)
	header7152785Hash := getEthHeaderHashByHeight(contract, 7152785)
	assert.Equal(t, true, bytes.Equal(ethcommon.HexToHash("90a1bc9c5f2e29ce1f605b23f3fa6bb064fdbe553a8162bfdfa3b9bb8d0600e7").Bytes(), header7152785Hash.Bytes()))
	header7152785_formstore := getEthHeaderByHash(contract, header7152785Hash)
	assert.Equal(t, true, bytes.Equal(header7152785_formstore, header7152785))
}

func getEthHeaderByHash(native *native.NativeContract, hash ethcommon.Hash) []byte {
	headerStore, _ := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(ethChainID), hash.Bytes()))
	headerBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil
	}
	var headerWithDifficultySum eth.HeaderWithDifficultySum
	if err := json.Unmarshal(headerBytes, &headerWithDifficultySum); err != nil {
		return nil
	}
	headerOnly, err := json.Marshal(headerWithDifficultySum.Header)
	if err != nil {
		return nil
	}
	return headerOnly
}

func getEthHeaderHashByHeight(native *native.NativeContract, height uint64) ethcommon.Hash {
	headerStore, _ := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(ethChainID), utils.GetUint64Bytes(height)))
	hashBytes, _ := cstates.GetValueFromRawStorageItem(headerStore)
	return ethcommon.BytesToHash(hashBytes)
}

func getEthLatestHeight(native *native.NativeContract) uint64 {
	contractAddress := utils.HeaderSyncContractAddress
	key := append([]byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(ethChainID)...)
	// try to get storage
	result, err := native.GetCacheDB().Get(utils.ConcatKey(contractAddress, key))
	if err != nil {
		return 0
	}
	if result == nil || len(result) == 0 {
		return 0
	} else {
		heightBytes, _ := cstates.GetValueFromRawStorageItem(result)
		return binary.LittleEndian.Uint64(heightBytes)
	}
}
